/**
 *
 * @author yinjk
 * @create 2019-06-19 11:44
 */
package maps

import (
	"github.com/yinjk/go-utils/pkg/utils/collection/list"
	"github.com/yinjk/go-utils/pkg/utils/collection/set"
)

type Map interface {
	//Put put one data to the map
	Put(key interface{}, value interface{})

	//PutAll put more data to the map
	PutAll(maps map[interface{}]interface{})

	//PutIfAbsent 如果map中没有则添加,返回true，如果map中有则返回false表示没有添加
	PutIfAbsent(key interface{}, value interface{}) bool

	//Get get one data on the map
	Get(key interface{}) interface{}

	//GetOrDefault get one data, and returns the default data if key not found
	GetOrDefault(key interface{}, defaultVal interface{}) (value interface{})

	//Remove remove one data equals the key, and return the old data who is deleted
	Remove(key interface{}) (old interface{})

	//KeySet returns all keys to this map
	KeySet(key interface{}) (keys *set.HashSet)

	//Values returns all values to this map
	Values(key interface{}) (values list.List)

	//Clear clear the all data and fast to gc
	Clear()

	//Size returns the size for the concurrent hash map
	Size() int
}
