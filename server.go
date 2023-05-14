package main

import (
	"embed"
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/config"
	"github.com/GameDev-Tube/famemma.PRE2E/services/api"
	"github.com/GameDev-Tube/famemma.PRE2E/services/db"
	"github.com/GameDev-Tube/famemma.PRE2E/services/engine"
	"github.com/GameDev-Tube/famemma.PRE2E/services/indexer"
	"os"
)

//go:embed fe/*.html fe/*.css fe/*.js fe/abi.json
var feFiles embed.FS

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Print("\n\n\n\n\n")
		fmt.Println("failed to load config: ", err)
		fmt.Println("does file \"config.json\" exist?")
		return
	}
	database, err := db.New(cfg.PSQLDsn, cfg.IndexerStartAt)
	if err != nil {
		fmt.Print("\n\n\n\n\n")
		fmt.Println(err.Error())
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "dropdb" {
		fmt.Printf("\nDROPPING TABLES\n")
		if err := database.Drop(); err != nil {
			fmt.Printf("\n\n\n%v\n", err)
		} else {
			fmt.Printf("OK\n")
		}
		return
	}
	idx, rpc, bind, err := indexer.New(database, cfg.P2EContract, cfg.RPC)
	if err != nil {
		fmt.Print("\n\n\n\n\n")
		fmt.Println(err.Error())
		return
	}
	eng := engine.New(database)
	srv := api.New(database, bind, rpc, cfg, feFiles)

	go idx.Run()
	go eng.Run()
	srv.Run()
}
