package main

import (
	"bytes"
	"encoding/json"
	"github.com/coreos/dex/client"
	"github.com/coreos/dex/client/manager"
	"github.com/coreos/dex/connector"
	"github.com/coreos/dex/db"
	"github.com/coreos/go-oidc/oidc"
)

func newDBDriver(dsn string) (driver, error) {
	dbc, err := db.NewConnection(db.Config{DSN: dsn})
	if err != nil {
		return nil, err
	}

	drv := &dbDriver{
		cfgRepo:   db.NewConnectorConfigRepo(dbc),
		ciManager: manager.NewClientManager(db.NewClientRepo(dbc), db.TransactionFactory(dbc), manager.ManagerOptions{}),
	}

	return drv, nil
}

type dbDriver struct {
	ciManager *manager.ClientManager
	cfgRepo   *db.ConnectorConfigRepo
}

func (d *dbDriver) NewClient(meta oidc.ClientMetadata) (*oidc.ClientCredentials, error) {
	if err := meta.Valid(); err != nil {
		return nil, err
	}
	cli := client.Client{
		Metadata: meta,
	}
	return d.ciManager.New(cli, nil)
}

func (d *dbDriver) ConnectorConfigs() ([]interface{}, error) {
	var configs []interface{}
	if configModels, err := d.cfgRepo.All(); err != nil {
		return configs, err
	} else {
		for _, configModel := range configModels {
			configs = append(configs, configModel)
		}
	}
	return configs, nil
}

func (d *dbDriver) SetConnectorConfigs(cfgs []interface{}) error {
	b, marshalErr := json.Marshal(cfgs)
	if marshalErr != nil {
		return marshalErr
	}
	configs, readErr := connector.ReadConfigs(bytes.NewReader(b))
	if readErr != nil {
		return readErr
	}
	return d.cfgRepo.Set(configs)
}
