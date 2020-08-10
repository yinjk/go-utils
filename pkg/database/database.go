/**
 *
 * @author yinjk
 * @create 2019-05-21 11:11
 */
package database

import (
	"github.com/yinjk/go-utils/pkg/database/influx"
	"github.com/yinjk/go-utils/pkg/database/mysql"
	"github.com/yinjk/go-utils/pkg/database/redis"
	"github.com/prometheus/common/log"
)

var baseDao *BaseDao

type Config struct {
	Mysql  *mysql.Config
	Redis  *redis.Config
	Influx *influx.Config
}

//BaseDao base dao
type BaseDao struct {
	db          *mysql.BaseOrm
	influxDB    *influx.DB
	redis       *redis.Pool
	redisExpire int32
}

func Get(config *Config) (dao *BaseDao) {
	if baseDao == nil {
		baseDao = New(config)
	}
	return baseDao
}

// New new a dao and return.
func New(config *Config) (dao *BaseDao) {
	var (
		db       *mysql.BaseOrm
		rs       *redis.Pool
		influxDB *influx.DB
	)
	if config.Mysql != nil {
		log.Info("init db connection")
		db = mysql.NewMySQL(config.Mysql)
	}
	if config.Influx != nil {
		influxDB = influx.New(config.Influx)
	}
	if config.Redis != nil {
		rs = redis.GetRedisPool(config.Redis)
	}
	dao = &BaseDao{
		// mysql
		db: db,
		// redis
		redis: rs,
		// influxDB
		influxDB: influxDB,
	}
	return
}

func (d *BaseDao) DB() *mysql.BaseOrm {
	return d.db
}

func (d *BaseDao) Redis() *redis.Pool {
	return d.redis
}

func (d *BaseDao) InfluxDB() *influx.DB {
	return d.influxDB
}

// Close close the resource.
func (d *BaseDao) Close() (err error) {
	if err = d.redis.Close(); err != nil {
		log.Error(err)
	}
	if err = d.db.Close(); err != nil {
		log.Error(err)
	}
	return
}
