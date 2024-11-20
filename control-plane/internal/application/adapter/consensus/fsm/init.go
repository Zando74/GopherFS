package fsm

import (
	"sync"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"github.com/dgraph-io/badger/v2"
)

type BadgerDB struct {
	once sync.Once
	db   *badger.DB
}

func NewBadgerDB() *badger.DB {
	log := logger.LoggerSingleton.GetInstance()
	cfg := config.ConfigSingleton.GetInstance()
	badgerOpt := badger.DefaultOptions(cfg.Consensus.RaftDir).
		WithLogger(log).WithLoggingLevel(badger.INFO)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		log.Fatal(err)
	}

	return badgerDB
}

func (bs *BadgerDB) GetInstance() *badger.DB {
	bs.once.Do(func() {
		bs.db = NewBadgerDB()
	})
	return bs.db
}

var BabdgerSingleton BadgerDB
