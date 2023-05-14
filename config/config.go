package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	RPC            string `json:"rpc_addr"`
	PSQLDsn        string `json:"psql_dsn"`
	IndexerStartAt uint64 `json:"indexer_start_block"`
	AuthKey        string `json:"auth_key"`
	P2EContract    string `json:"p2e_contract"`
	BEP20Contract  string `json:"bep20_contract"`
	Listen         string `json:"listen"`
	// for JS
	ExternalUrl string `json:"external_url"`
	// for JS, ie if reverse ssl-wrapping proxy was used
	ExternalIsHttps bool `json:"is_https"`
}

func LoadConfig() (Config, error) {
	f, err := os.OpenFile("config.json", os.O_RDONLY, 0600)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	cfgdata, err := io.ReadAll(f)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	out := Config{}
	err = json.Unmarshal(cfgdata, &out)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	return out, nil
}
