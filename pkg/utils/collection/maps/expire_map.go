/*
一种过期map，数据只会在map中保存一段固定时间，超过时间之后数据会失效被删除，
而真正的失效判断和删除操作是不会自动发生的，当用户调用get方法的时候才会去检测key是否失效并删除数据，
该map类似于redis中设置过期时间的key。

使用场景： 由于该map和redis的过期key类似，所以可以用作一起缓存场景，或则一些超时处理场景，比如gpu卡配置我们会等待配置的成功，
这里有一个超时，超时之后认为配置失败，我们就可以使用该map来实现。
*/
package maps

import (
	"sync"
	"time"
)

const (
	_maxSize           = 500
	_defaultExpireTime = time.Minute * 10
)

type expireData struct {
	//上一次时间单位为秒
	LastTime time.Time
	//数据
	Data interface{}
}

func newExpireData(value interface{}) expireData {
	expireData := expireData{}
	expireData.LastTime = time.Now()
	expireData.Data = value
	return expireData
}

type ExpireMap struct {
	data       map[string]*expireData
	lock       sync.Mutex
	size       int64
	expireTime time.Duration
}

func NewExpireMap(times ...time.Duration) *ExpireMap {
	if nil == times || len(times) == 0 {
		return &ExpireMap{
			expireTime: _defaultExpireTime,
		}
	}
	return &ExpireMap{expireTime: times[0]}
}

func (em *ExpireMap) ForEach(test func(key string, data interface{})) {
	if em.size == 0 {
		return
	}
	for key := range em.data {
		value := em.Get(key)
		if value != nil {
			test(key, value)
		}
	}
}

//获取并刷新
func (em *ExpireMap) GetAndFlush(key string) interface{} {
	em.lock.Lock()
	defer em.lock.Unlock()
	if em.data == nil {
		return nil
	}
	expireData := em.data[key]
	if expireData == nil {
		return nil
	}
	now := time.Now()
	if expireData.LastTime.Add(em.expireTime).Before(now) {
		delete(em.data, key)
		em.size--
		return nil
	}
	//刷新下过期时间 取消刷新
	expireData.LastTime = now
	return expireData.Data
}

//获取数据,并进行惰性删除
func (em *ExpireMap) Get(key string) interface{} {
	em.lock.Lock()
	defer em.lock.Unlock()
	if em.data == nil {
		return nil
	}
	expireData := em.data[key]
	if expireData == nil {
		return nil
	}
	if expireData.LastTime.Add(em.expireTime).Before(time.Now()) {
		delete(em.data, key)
		em.size--
		return nil
	}
	return expireData.Data
}

//存放数据
func (em *ExpireMap) Put(key string, value interface{}) {
	em.lock.Lock()
	defer em.lock.Unlock()
	if em.data == nil {
		em.data = make(map[string]*expireData)
	}
	if em.size >= _maxSize {
		em.recycle()
	}
	expireData := newExpireData(value)
	em.data[key] = &expireData
	em.size++
}

//移除掉对应key的数据
func (em *ExpireMap) Remove(key string) interface{} {
	em.lock.Lock()
	defer em.lock.Unlock()
	if em.data == nil {
		return nil
	}
	value := em.data[key]
	delete(em.data, key)
	em.size--
	if value != nil {
		return value.Data
	}
	return nil
}

//如果超过大小就回收掉过期token，这里进行条件触发删除
func (em *ExpireMap) recycle() {
	now := time.Now()
	var deleteKey []string
	for key := range em.data {
		temp := em.data[key]
		lastTime := temp.LastTime
		if lastTime.Add(em.expireTime).Before(now) {
			deleteKey = append(deleteKey, key)
		}
	}
	if deleteKey != nil && len(deleteKey) > 0 {
		for _, key := range deleteKey {
			delete(em.data, key)
			em.size--
		}
	}
}
