package aggregator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BOPR/bazooka"
	"github.com/BOPR/config"
	"github.com/BOPR/core"
	"github.com/BOPR/db"
	"github.com/BOPR/log"
)

const (
	AggregatingService = "aggregator"
)

// Aggregator is the service which is supposed to create batches
// It has the following tasks:
// 1. Pick txs from the mempool
// 2. Validate these trnsactions
// 3. Update the DB post running each tx
// 4. Finally create a batch of all the transactions and post on-chain
type Aggregator struct {
	// Base service
	core.BaseService

	// contract caller to interact with contracts
	LoadedBazooka bazooka.Bazooka

	// DB instance
	DB db.DB

	Bazooka bazooka.Bazooka

	// header listener subscription
	cancelAggregating context.CancelFunc

	// wait group
	wg sync.WaitGroup
}

// NewAggregator returns new aggregator object
func NewAggregator() *Aggregator {
	// create logger
	logger := log.Logger.With("module", AggregatingService)
	LoadedBazooka, err := bazooka.NewPreLoadedBazooka()
	if err != nil {
		panic(err)
	}
	aggregator := &Aggregator{}
	aggregator.BaseService = *core.NewBaseService(logger, AggregatingService, aggregator)
	DB, err := db.NewDB()
	if err != nil {
		panic(err)
	}

	bz, err := bazooka.NewPreLoadedBazooka()
	if err != nil {
		panic(err)
	}
	aggregator.Bazooka = bz
	aggregator.DB = DB
	aggregator.LoadedBazooka = LoadedBazooka
	return aggregator
}

// OnStart starts new block subscription
func (a *Aggregator) OnStart() error {
	err := a.BaseService.OnStart() // Always call the overridden method.
	if err != nil {
		return err
	}

	ctx, cancelAggregating := context.WithCancel(context.Background())
	a.cancelAggregating = cancelAggregating

	// start polling for checkpoint in buffer
	go a.startAggregating(ctx, config.GlobalCfg.PollingInterval)
	return nil
}

// OnStop stops all necessary go routines
func (a *Aggregator) OnStop() {
	a.BaseService.OnStop() // Always call the overridden method.
	a.DB.Close()
	// cancel ack process
	a.cancelAggregating()
}

func (a *Aggregator) startAggregating(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	// stop ticker when everything done
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if isCatchingUp, err := IsCatchingUp(a.Bazooka, a.DB); err != nil {
				return
			} else if isCatchingUp {
				a.Logger.Info("Commander catching up, aborting aggregation till next poll")
			} else {
				a.wg.Wait()
				a.wg.Add(1)
				go a.pickBatch()
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (a *Aggregator) pickBatch() {
	defer a.wg.Done()
	txs, err := a.DB.PopTxs()
	if err != nil {
		fmt.Println("Error while popping txs from mempool", "Error", err)
	}
	a.processAndSubmitBatch(txs)
}

func (a *Aggregator) processAndSubmitBatch(txs []core.Tx) {
	a.Logger.Info("Processing new batch", "numberOfTxs", len(txs))

	// Step-2
	commitments, err := a.processTxs(txs)
	if err != nil {
		fmt.Println("Error while processing tx", "error", err)
		return
	}

	// Step-3
	// Submit all commitments on-chain
	txHash, err := a.LoadedBazooka.SubmitBatch(commitments)
	if err != nil {
		fmt.Println("Error while submitting batch", "error", err)
		return
	}

	// Step-4
	// Record batch locally
	lastCommitment := commitments[len(commitments)-1]
	newBatch := core.NewBatch(lastCommitment.UpdatedRoot.String(), config.GlobalCfg.OperatorAddress, txHash, lastCommitment.BatchType, core.BATCH_BROADCASTED)
	err = a.DB.AddNewBatch(newBatch)
	if err != nil {
		return
	}
}

func (a *Aggregator) processTxs(txs []core.Tx) (commitments []core.Commitment, err error) {
	return db.ProcessTxs(&a.LoadedBazooka, &a.DB, txs, false)
}

// IsCatchingUp returns true/false according to the sync status of the node
func IsCatchingUp(b bazooka.Bazooka, db db.DB) (bool, error) {
	totalBatches, err := b.TotalBatches()
	if err != nil {
		return false, err
	}

	totalBatchedStored, err := db.GetBatchCount()
	if err != nil {
		return false, err
	}

	// if total batchse are greater than what we recorded we are still catching up
	if totalBatches > uint64(totalBatchedStored) {
		return true, err
	}

	return false, nil
}
