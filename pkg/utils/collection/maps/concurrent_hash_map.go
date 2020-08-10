/**
 * 同步hash map：该数据结构是一个线程安全的HashMap, 底层使用性能比较低的全量加锁方式实现，
 * 为追求更高的性能可以使用分段锁的方式来实现，详情见segment_hash_map。由于该结构性能低下，应该尽量避免使用
 * sync包提供了一个Map类，我们应该使用该类来完成线程安全的map的使用（由于之前不知道该接口，所以自己实现了一个）
 * @author yinjk
 * @create 2019-05-07 17:43
 */
package maps

import (
	"github.com/yinjk/go-utils/pkg/utils/collection/list"
	"github.com/yinjk/go-utils/pkg/utils/collection/set"
	"sync"
)

type ConcurrentHashMap struct {
	lock sync.Mutex
	data map[interface{}]interface{}
	sync.Map
}

func NewConcurrentHashMap() *ConcurrentHashMap {
	return &ConcurrentHashMap{
		data: make(map[interface{}]interface{}),
	}
}

//Put put one data to the map
func (m *ConcurrentHashMap) Put(key interface{}, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = value
}

//PutAll put more data to the map
func (m *ConcurrentHashMap) PutAll(maps map[interface{}]interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for k, v := range maps {
		m.data[k] = v
	}
}

//PutIfAbsent 如果map中没有则添加,返回true，如果map中有则返回false表示没有添加
func (m *ConcurrentHashMap) PutIfAbsent(key interface{}, value interface{}) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.data[key]; !ok {
		m.data[key] = value
		return true
	}
	return false
}

//Get get one data on the map
func (m *ConcurrentHashMap) Get(key interface{}) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.data[key]
}

//GetOrDefault get one data, and returns the default data if key not found
func (m *ConcurrentHashMap) GetOrDefault(key interface{}, defaultVal interface{}) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	if value, ok := m.data[key]; ok { //map中查到该key了
		return value
	}
	return defaultVal
}

//Remove remove one data equals the key, and return the old data who is deleted
func (m *ConcurrentHashMap) Remove(key interface{}) (old interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	old = m.data[key]
	delete(m.data, key)
	return
}

//KeySet returns all keys to this map
func (m *ConcurrentHashMap) KeySet(key interface{}) (keys *set.HashSet) {
	m.lock.Lock()
	defer m.lock.Unlock()

	keys = set.NewHashSet()
	for k := range m.data {
		keys.Add(k)
	}
	return
}

//Values returns all values to this map
func (m *ConcurrentHashMap) Values(key interface{}) (values list.List) {
	m.lock.Lock()
	defer m.lock.Unlock()

	values = list.NewArrayListWithCapacity(len(m.data))
	for _, v := range m.data {
		values.Add(v)
	}
	return
}

//Clear clear the all data and fast to gc
func (m *ConcurrentHashMap) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data = make(map[interface{}]interface{})
}

//Size returns the size for the concurrent hash map
func (m *ConcurrentHashMap) Size() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.data)
}
