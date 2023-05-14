package engine

import "github.com/GameDev-Tube/famemma.PRE2E/services/db"

type Engine struct {
	db *db.Db
}

func New(db *db.Db) *Engine {
	return &Engine{
		db: db,
	}
}
