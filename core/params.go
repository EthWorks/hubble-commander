package core

import (
	"math/big"
)

// Params stores all the parameters which are maintained on-chain and keeps updating
// them whenever they change on-chain
type Params struct {
	DBModel

	// Stake amount which coordinator needs to submit a new batch
	// Updates when syncer receives a stake update event from the contract
	// Used while sending new batch
	StakeAmount uint64 `json:"stakeAmount"`

	// MaxDepth is the maximum depth of the balances tree possible
	// If in case we want to increase it we will update the value on the contract
	// And then this will be updated
	MaxDepth uint64 `json:"maxDepth"`

	// DepositSubTreeHeight is the maximum height of the deposit subtree that the coordinator wants to merge
	// It is set on the contract and will be updated when that value changes
	MaxDepositSubTreeHeight uint64 `json:"maxDepositSubTreeHeight"`

	FinalisationTime uint64 `json:"finalisationTime"`
}

// Maintains sync information
type SyncStatus struct {
	ID string `json:"-" gorm:"primary_key;size:100;default:'6ba7b810-9dad-11d1-80b4-00c04fd430c8'"`
	// Last eth block seen by the syncer is persisted here so that we can resume sync from it
	LastEthBlockRecorded uint64 `json:"lastEthBlockRecorded"`
}

func (ss *SyncStatus) LastEthBlockBigInt() *big.Int {
	n := new(big.Int)
	return n.SetUint64(ss.LastEthBlockRecorded)
}

func (db *DB) InitSyncStatus(startBlock uint64) error {
	return db.Instance.Create(&SyncStatus{LastEthBlockRecorded: startBlock}).Error
}

func (db *DB) UpdateSyncStatusWithBlockNumber(blkNum uint64) error {
	syncStatus, err := db.GetSyncStatus()
	if err != nil {
		return err
	}
	var updatedSyncStatus SyncStatus
	updatedSyncStatus.LastEthBlockRecorded = blkNum
	if err := db.Instance.Table("sync_statuses").Where("id = ?", syncStatus.ID).Update(&updatedSyncStatus).Error; err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSyncStatus() (status SyncStatus, err error) {
	if err := db.Instance.First(&status).Error; err != nil {
		return status, err
	}
	return status, nil
}

// UpdateStakeAmount updates the stake amount
func (db *DB) UpdateStakeAmount(newStakeAmount uint64) error {
	var updatedParams Params
	updatedParams.StakeAmount = newStakeAmount
	if err := db.Instance.Table("params").Assign(Params{StakeAmount: newStakeAmount}).FirstOrCreate(&updatedParams).Error; err != nil {
		return err
	}
	return nil
}

// UpdateFinalisationTimePerBatch updates the max depth
func (db *DB) UpdateFinalisationTimePerBatch(finalisationDuration uint64) error {
	var updatedParams Params
	updatedParams.MaxDepth = finalisationDuration
	if err := db.Instance.Table("params").Assign(Params{FinalisationTime: finalisationDuration}).FirstOrCreate(&updatedParams).Error; err != nil {
		return err
	}
	return nil
}

// UpdateMaxDepth updates the max depth
func (db *DB) UpdateMaxDepth(newDepth uint64) error {
	var updatedParams Params
	updatedParams.MaxDepth = newDepth
	if err := db.Instance.Table("params").Assign(Params{MaxDepth: newDepth}).FirstOrCreate(&updatedParams).Error; err != nil {
		return err
	}
	return nil
}

// UpdateDepositSubTreeHeight updates the max height of deposit sub tree
func (db *DB) UpdateDepositSubTreeHeight(newHeight uint64) error {
	var updatedParams Params
	updatedParams.MaxDepositSubTreeHeight = newHeight
	if err := db.Instance.Table("params").Assign(Params{MaxDepositSubTreeHeight: newHeight}).FirstOrCreate(&updatedParams).Error; err != nil {
		return err
	}
	return nil
}

func (db *DB) GetBatchFinalisationTime() (uint64, error) {
	var params Params
	if err := db.Instance.First(&params).Error; err != nil {
		return 0, err
	}
	return params.FinalisationTime, nil
}

// GetParams gets params from the DB
func (db *DB) GetParams() (params Params, err error) {
	if err := db.Instance.First(&params).Error; err != nil {
		return params, err
	}
	return params, nil
}
