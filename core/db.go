package core

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/BOPR/common"
	"github.com/BOPR/config"
	"github.com/globalsign/mgo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type IDB interface {
	// Account related DB functions
	// FetchSiblings(accID uint64) (accs []UserState, err error)
	GetAllAccounts() (accs []UserState, err error)
	GetAccount(accID uint64) (UserState, error)
	InsertBulkAccounts(accounts []UserState) error
	InsertGenAccounts(genAccs []config.GenUserState) error
	GetAccountCount() (int, error)

	// Tx related functions
	InsertTx(t *Tx) error
	PopTxs() (txs []Tx, err error)

	// Batch related functions
	InsertBatchInfo(root ByteArray, index uint64) error
	GetAllBatches() (batches []Batch, err error)
	GetLatestBatch() (Batch, error)
	GetBatchCount() (int, error)

	// common functions
	GetBatchCollection() *mgo.Collection
	GetTransactionCollection() *mgo.Collection
	GetAccountCollection() *mgo.Collection
}

// global DB instance created while doing init
var DBInstance DB

type DB struct {
	Instance *gorm.DB
	Bazooka  Bazooka
	Logger   log.Logger
}

// NewDB creates a new DB instance
// NOTE: it uses the configrations present in the config.toml file
// returns error if not able to open the DB
func NewDB() (DB, error) {
	if err := config.ParseAndInitGlobalConfig(""); err != nil {
		return DB{}, err
	}
	db, err := gorm.Open(config.GlobalCfg.DB, config.GlobalCfg.FormattedDBURL())
	if err != nil {
		return DB{}, err
	}
	db.LogMode(config.GlobalCfg.DBLogMode)
	// create logger
	logger := common.Logger.With("module", "DB")

	bz, err := NewPreLoadedBazooka()
	if err != nil {
		return DB{}, err
	}
	return DB{Instance: db, Logger: logger, Bazooka: bz}, nil
}

func (db *DB) Close() {
	db.Instance.Close()
}
