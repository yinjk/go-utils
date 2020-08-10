//@Desc
//@Date 2019-11-20 22:01
//@Author yinjk
package influx

import (
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	iclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/prometheus/common/log"
)

type Config struct {
	Addr     string
	Username string
	Password string
	Database string
}

type DB struct {
	iclient.Client
	Database string
}

func New(config *Config) *DB {
	influxClient, err := iclient.NewHTTPClient(iclient.HTTPConfig{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
	})
	if err != nil {
		log.Errorf("StatisticsHandler New influx db failed, err=%s \n", err)
		panic(err)
	}
	return &DB{
		Client:   influxClient,
		Database: config.Database,
	}
}
