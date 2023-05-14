package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Db struct {
	orm        *gorm.DB
	startBlock uint64
}

func New(dsn string, startBlock uint64) (*Db, error) {
	lgr := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Error,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
	orm, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: lgr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Db{
		orm:        orm,
		startBlock: startBlock,
	}, orm.AutoMigrate(&IndexerStatus{}, &Prediction{}, &GameState{}, &SwapList{})
}
