/*
 @Desc

 @Date 2020-05-18 11:04
 @Author yinjk
*/
package maps

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestExpireCache_Get(t *testing.T) {
	cache := NewExpireCache()
	cache.Put("forevear", 1)
	cache.Put("2s", 1, time.Second*2)
	cache.Put("1s", 1, time.Second*1)
	cache.Put("5s", 1, time.Second*5)
	cache.Put("10s", 1, time.Second*10)
	cache.Put("1m", 1, time.Minute*1)
	i := 0
	for true {
		i++
		fmt.Println("-------"+strconv.Itoa(i)+"--------", cache.Size())
		cache.ForEach(func(key string, data interface{}) {
			fmt.Println(key)
		})
		time.Sleep(time.Second)
	}
}

func TestNewExpireMap(t *testing.T) {
	expireMap := NewExpireMap(time.Second * 2)
	expireMap.Put("aa", time.Second)
	expireMap.Put("bb", time.Second)
	i := 0
	for true {
		i++
		fmt.Println("-------" + strconv.Itoa(i) + "--------")
		expireMap.ForEach(func(key string, data interface{}) {
			fmt.Println(key)
		})
		time.Sleep(time.Second)
	}

}
