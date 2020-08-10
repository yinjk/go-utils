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

type expireCache struct {
	//上一次时间单位为秒
	createTime time.Time
	ttl        time.Duration
	//数据
	Data interface{}
}

func newExpireCache(value interface{}, ttl time.Duration) expireCache {
	expireData := expireCache{}
	expireData.createTime = time.Now()
	expireData.ttl = ttl
	expireData.Data = value
	return expireData
}

type ExpireCache struct {
	sync.Mutex
	data map[string]*expireCache
	back bool
	size int
}

func NewExpireCache(b ...bool) *ExpireCache {
	em := &ExpireCache{
		data: make(map[string]*expireCache),
	}
	if b != nil && len(b) > 0 {
		em.back = b[0]
	}
	if em.back {
		go em.background()
	}
	return em
}

func (em *ExpireCache) background() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			em.clear()
		}
	}
}

func (em *ExpireCache) clear() {
	em.ForEach(func(key string, data interface{}) {
	})
}
func (em *ExpireCache) ForEach(test func(key string, data interface{})) {
	if em.size == 0 {
		return
	}
	for key, _ := range em.data {
		value := em.Get(key)
		if value != nil {
			test(key, value)
		}
	}
}

//获取并刷新
func (em *ExpireCache) GetAndFlush(key string) interface{} {
	em.Lock()
	defer em.Unlock()
	expireData := em.data[key]
	if expireData == nil {
		return nil
	}
	now := time.Now()
	if expireData.ttl <= 0 {
		expireData.createTime = now
		return expireData.Data
	}
	if expireData.createTime.Add(expireData.ttl).Before(now) {
		delete(em.data, key)
		em.size--
		return nil
	}
	expireData.createTime = now
	return expireData.Data
}

//获取数据,并进行惰性删除
func (em *ExpireCache) Get(key string) interface{} {
	em.Lock()
	defer em.Unlock()
	expireData := em.data[key]
	if expireData == nil {
		return nil
	}
	if expireData.ttl <= 0 {
		return expireData.Data
	}
	if expireData.createTime.Add(expireData.ttl).Before(time.Now()) {
		delete(em.data, key)
		em.size--
		return nil
	}
	return expireData.Data
}

//Put 存放数据
func (em *ExpireCache) Put(key string, value interface{}, duration ...time.Duration) {
	em.Lock()
	defer em.Unlock()
	if em.size >= _maxSize {
		em.recycle()
	}
	var ttl time.Duration = 0
	if len(duration) > 0 {
		ttl = duration[0]
	}
	expireData := newExpireCache(value, ttl)
	em.data[key] = &expireData
	em.size++
}

//Update 更新cache中的数据，并且可以重新设置过期时间
func (em *ExpireCache) Update(key string, value interface{}, duration ...time.Duration) {
	em.Lock()
	defer em.Unlock()
	var ttl time.Duration = 0
	if len(duration) > 0 {
		ttl = duration[0]
	}
	expireData := em.data[key]
	if expireData == nil {
		em.Put(key, value, ttl)
		return
	}
	expireData.Data = value
	if ttl != 0 {
		expireData.createTime = time.Now()
		expireData.ttl = ttl
	}
	em.data[key] = expireData
}

//Expire 重置过期时间，设置新的ttl
func (em *ExpireCache) Expire(key string, ttl time.Duration) {
	em.Lock()
	defer em.Unlock()
	value := em.Get(key)
	if value == nil {
		return
	}
	em.Update(key, value, ttl)
}

//TTL the cache key current ttl
func (em *ExpireCache) TTL(key string) time.Duration {
	em.Lock()
	defer em.Unlock()
	expireData := em.data[key]
	if expireData == nil {
		return 0
	}
	ttl := time.Now().Sub(expireData.createTime)
	return expireData.ttl - ttl
}

//Size the cached data count
func (em *ExpireCache) Size() int {
	if em.back {
		return em.size
	} else {
		em.clear()
		return em.size
	}
}

//移除掉对应key的数据
func (em *ExpireCache) Remove(key string) interface{} {
	em.Lock()
	defer em.Unlock()
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

//回收过期key，这里进行条件触发删除
func (em *ExpireCache) recycle() {
	now := time.Now()
	var deleteKey []string
	for key := range em.data {
		temp := em.data[key]
		if temp.createTime.Add(temp.ttl).Before(now) {
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
