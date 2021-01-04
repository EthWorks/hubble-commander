package db

import (
	"math/big"

	"github.com/BOPR/bazooka"
	"github.com/BOPR/core"
)

func (db *DB) GetLatestBatch() (batch core.Batch, err error) {
	if err := db.Instance.Order("batch_id desc").First(&batch).Error; err != nil {
		return batch, err
	}
	return batch, nil
}

func (db *DB) GetBatchCount() (int, error) {
	var count int
	if err := db.Instance.Table("batches").Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}

func (db *DB) AddNewBatch(batch core.Batch, commitments []core.Commitment) error {
	batchCount, err := db.GetBatchCount()
	if err != nil {
		return err
	}

	// this is because batch ID starts from 0
	batch.BatchID = uint64(batchCount)

	err = db.addCommitmentsWithBatchID(commitments, batch.BatchID)
	if err != nil {
		return err
	}

	return db.Instance.Create(batch).Error
}

func (db *DB) GetBatchByIndex(index uint64) (batch core.Batch, err error) {
	if err := db.Instance.Where("batch_id = ?", index).Find(&batch).Error; err != nil {
		return batch, err
	}
	return batch, nil
}

func (db *DB) CommitBatch(ID uint64) error {
	batch, err := db.GetBatchByIndex(ID)
	if err != nil {
		return err
	}

	if batch.Status == core.BATCH_COMMITTED {
		return core.ErrBatchAlreadyCommitted
	}

	batch.Status = core.BATCH_COMMITTED
	return db.Instance.Model(&batch).Where("batch_id = ?", batch.BatchID).Update(batch).Error
}

func (db *DB) addCommitmentsWithBatchID(commitments []core.Commitment, batchID uint64) error {
	if len(commitments) == 0 {
		return nil
	}
	dbCommitments := core.NewCommitments(commitments, batchID)
	return db.Instance.Create(&dbCommitments).Error
}

// GetCommitmentsForBatch getter for all commitments sent in a batch in order of offset
func (db *DB) GetCommitmentsForBatch(id uint64) (commitments []core.Commitment, err error) {
	if err = db.Instance.Model(&commitments).Where("batch_id = ?", id).Order("offset asc").Find(&commitments).Error; err != nil {
		return
	}
	return
}

func (db *DB) GetLastCommitmentMP(id uint64) (commitmentMP bazooka.TypesCommitmentInclusionProof, err error) {
	commitments, err := db.GetCommitmentsForBatch(id)
	if err != nil {
		return
	}

	var leaves []core.ByteArray
	for _, commitment := range commitments {
		leaf, err := commitment.CommitmentData.Leaf()
		if err != nil {
			return commitmentMP, err
		}
		leaves = append(leaves, leaf)
	}

	tree, err := core.NewTree(leaves)
	if err != nil {
		return
	}
	lastCommitment := len(commitments)
	_, witnesses, err := tree.GetWitnessForLeaf(uint64(lastCommitment))
	if err != nil {
		return
	}

	var solWitness [][32]byte

	for _, witness := range witnesses {
		solWitness = append(solWitness, witness)
	}

	commitmentMP = bazooka.TypesCommitmentInclusionProof{
		Commitment: commitments[lastCommitment].CommitmentData,
		Path:       big.NewInt(int64(lastCommitment)),
		Witness:    solWitness,
	}

	return commitmentMP, nil
}
