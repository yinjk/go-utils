/**
 *
 * @author yinjk
 * @create 2019-01-30 16:26
 */
package redis

import (
	"fmt"
	"testing"
)

var (
	pool *Pool
)

func init() {
	pool = GetRedisPool(nil)
}

func TestClient_HmSet(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	map1 := make(map[string]interface{})
	map1["name"] = "张珊"
	map1["sex"] = "女"
	map1["age"] = 18
	map1["list"] = []string{"111", "222", "333"}
	if err := client.HmSet("map1", map1); err != nil {
		panic(err)
	}
}

func TestClient_HGet(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	fmt.Println(client.HGet("map1", "list"))
}

func TestClient_ZGetMaxScore(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	i, _ := client.ZGetMaxScore("students")
	fmt.Println(i)
}

func TestClient_DeleteKey(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	i, _ := client.DeleteKey("a1", "a2")
	fmt.Println(i)
}

func TestClient_ZClearAndAddSet(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	args := []interface{}{"1", "2", "a", "b", "c", "e"}
	ok := client.ZClearAndAddSet("zSetKey", args)
	fmt.Println(ok)
}

func TestClient_ZRemove(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	i, _ := client.ZRemove("zSetKey", "b", "c", "v")
	fmt.Println(i)
}

func TestClient_ZRange(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	result := client.ZRange("zSetKey", 0, -1)
	fmt.Println(result)
}

func TestClient_ZRevRange(t *testing.T) {
	client, _ := pool.Get()
	defer client.Close()
	result := client.ZRevRange("zSetKey", 0, -1)
	fmt.Println(result)
}
