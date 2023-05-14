package api

import (
	"embed"
	"github.com/GameDev-Tube/famemma.PRE2E/bindings"
	"github.com/GameDev-Tube/famemma.PRE2E/config"
	"github.com/GameDev-Tube/famemma.PRE2E/services/db"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Api struct {
	db              *db.Db
	h               *handler
	frontend        embed.FS
	sockBind        string
	externalUrl     string
	externalIsHttps bool
	p2eAddress      string
	bepAddress      string
}

func New(db *db.Db, binding *bindings.P2EContract, rpc *ethclient.Client, cfg config.Config, frontend embed.FS) *Api {
	return &Api{
		db: db,
		h: &handler{
			db:       db,
			rpc:      rpc,
			binding:  binding,
			adminkey: cfg.AuthKey,
		},
		frontend:        frontend,
		sockBind:        cfg.Listen,
		externalUrl:     cfg.ExternalUrl,
		externalIsHttps: cfg.ExternalIsHttps,
		p2eAddress:      cfg.P2EContract,
		bepAddress:      cfg.BEP20Contract,
	}
}

type handler struct {
	db       *db.Db
	rpc      *ethclient.Client
	binding  *bindings.P2EContract
	adminkey string
}
