/**
 * redis 通用api实现类
 * @author yinjk
 * @create 2019-01-30 14:44
 */
package redis

import (
	"github.com/yinjk/go-utils/pkg/utils/convert"
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/common/log"
	"time"
)

type Config struct {
	Addr              string
	Password          string
	Database          string
	MaxIdle           int
	MaxActive         int
	IdleTimeout       int
	DefaultExpireTime int
}

type Client struct {
	conn redis.Conn
}

func GetRedisPool(config *Config) (pool *Pool) {
	p := &redis.Pool{
		MaxIdle:     config.MaxIdle,
		MaxActive:   config.MaxActive,
		IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Addr)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			// 鉴权 登陆
			if _, err = c.Do("AUTH", config.Password); err != nil {
				log.Error(err)
				return nil, err
			}
			// 选择db
			if _, err = c.Do("SELECT", config.Database); err != nil {
				log.Error(err)
				return nil, err
			}
			return c, nil
		},
	}
	return &Pool{pool: p}
}

type Pool struct {
	pool *redis.Pool
}

func (p Pool) Get() (client *Client, err error) {
	conn := p.pool.Get()
	return &Client{conn: conn}, conn.Err()
}

func (p Pool) Close() (err error) {
	return p.pool.Close()
}

func (c Client) Close() error {
	return c.conn.Close()
}

func (c Client) Multi() (err error) {
	_, err = c.conn.Do("MULTI") //开启redis事物
	return err
}

func (c Client) Exec() (err error) {
	_, err = c.conn.Do("EXEC") //执行事物
	return err
}

func (c Client) Discard() (err error) {
	_, err = c.conn.Do("DISCARD") //取消事物，放弃事物块内的所有命令
	return err
}

func (c Client) DeleteKey(keys ...interface{}) (removed int, err error) {
	return redis.Int(c.conn.Do("DEL", keys...))
}

func (c Client) Exists(key string) (b bool, err error) {
	return redis.Bool(c.conn.Do("EXISTS", key))
}

func (c Client) GetString(key string) (value string, err error) {
	return redis.String(c.conn.Do("get", key))
}

func (c Client) GetInt(key string) (value int, err error) {
	return redis.Int(c.conn.Do("get", key))
}

func (c Client) Set(key, value interface{}) (err error) {
	_, err = c.conn.Do("set", key, value)
	return err
}

func (c Client) Expire(key string, time int) (err error) {
	_, err = c.conn.Do("expire", key, time)
	return
}

func (c Client) HSet(key, field string, value interface{}) (err error) {
	_, err = c.conn.Do("hset", key, field, value)
	return err
}

func (c Client) HGet(key, field string) (value string, err error) {
	return redis.String(c.conn.Do("hget", key, field))
}

func (c Client) HGetAll(key string) (maps map[string]string, err error) {
	return redis.StringMap(c.conn.Do("hgetall", key))
}
func (c Client) HmSet(key string, data map[string]interface{}) (err error) {
	_, err = c.conn.Do("hmset", redis.Args{}.Add(key).AddFlat(data)...)
	return err
}

func (c Client) ZAdd(key string, score int, value interface{}) error {
	_, err := c.conn.Do("zadd", key, score, value)
	return err
}

func (c Client) ZCard(key string) (value int, err error) {
	return redis.Int(c.conn.Do("ZCARD", key))
}

func (c Client) ZScore(key, member string) (i int, err error) {
	return redis.Int(c.conn.Do("ZSCORE", key, member))
}

/**
 * 移除有序集合中的一个或多个成员
 * @param : 集合的key
 * @param : 集合key中的值
 * @return: 移除的值的个数
 * @author: yinjk
 * @time  : 2019/2/11 15:25
 */
func (c Client) ZRemove(key string, member ...interface{}) (i int, err error) {
	args := make([]interface{}, 1)
	args[0] = key
	args = append(args, member...)
	return redis.Int(c.conn.Do("ZREM", args...))
}

func (c Client) ZAddSet(key string, values []interface{}) (err error) {
	zSetArgs := make([]interface{}, len(values)*2+1) //拼装 zadd的参数
	zSetArgs[0] = key
	for i, v := range values {
		zSetArgs[2*i+1] = i
		zSetArgs[2*i+2] = v
	}
	_, err = c.conn.Do("zadd", zSetArgs...)
	return err
}

func (c Client) ZAddSetWithScore(key string, values []interface{}, genScore func(value interface{}) int) bool {
	zSetArgs := make([]interface{}, len(values)*2+1) //拼装 zadd的参数
	zSetArgs[0] = key
	for i, v := range values {
		zSetArgs[2*i+1] = genScore(v)
		zSetArgs[2*i+2] = v
	}
	_, err := c.conn.Do("zadd", zSetArgs...)
	if err != nil {
		panic(err)
		return false
	}
	return true
}

func (c Client) ZClearAndAddSet(key string, values []interface{}) (err error) {
	//开启redis事物
	if err = c.Multi(); err != nil {
		return
	}
	//先清空zSet
	if _, err = c.DeleteKey(key); err != nil {
		_ = c.Discard()
		return
	}
	//添加set
	result := c.ZAddSet(key, values)
	//执行事物
	if err = c.Exec(); err != nil {
		_ = c.Discard()
		return
	}
	return result
}

func (c Client) ZGetMaxScore(key string) (score int, err error) {
	values, err := redis.Strings(c.conn.Do("zrevrange", key, 0, 0, "withscores"))
	if err != nil || len(values) != 2 {
		return -1, err
	}
	return convert.StringToInt(values[1]), nil
}

type ZSetValue struct {
	Score int
	Value string
}

func (c Client) ZRange(key string, start, stop int) []ZSetValue {
	//返回的values数组为值和分数，格式如：[value1 score1 value2 score2 ...]
	values, err := redis.Strings(c.conn.Do("zrange", key, start, stop, "withscores"))
	if err != nil {
		return nil
	}
	//将values封装成ZSetValue类型数组
	isScore := false
	result := make([]ZSetValue, len(values)/2)
	var zSetValue ZSetValue
	for i, v := range values {
		if isScore {
			zSetValue.Score = convert.StringToInt(v)
			result[(i-1)/2] = zSetValue
			isScore = false
		} else {
			zSetValue = ZSetValue{Value: v}
			isScore = true
		}
	}
	return result
}

func (c Client) ZRevRange(key string, start, stop int) []ZSetValue {
	//返回的values数组为值和分数，格式如：[value1 score1 value2 score2 ...]
	values, err := redis.Strings(c.conn.Do("ZREVRANGE", key, start, stop, "withscores"))
	if err != nil {
		return nil
	}
	//将values封装成ZSetValue类型数组
	isScore := false
	result := make([]ZSetValue, len(values)/2)
	var zSetValue ZSetValue
	for i, v := range values {
		if isScore {
			zSetValue.Score = convert.StringToInt(v)
			result[(i-1)/2] = zSetValue
			isScore = false
		} else {
			zSetValue = ZSetValue{Value: v}
			isScore = true
		}
	}
	return result
}
