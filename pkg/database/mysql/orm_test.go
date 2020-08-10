/*
 @Desc

 @Date 2020-04-26 16:01
 @Author yinjk
*/
package mysql

import (
	"fmt"
	"testing"
)

var config Config

func init() {
	config = Config{
		DataBase:    "test",
		DSN:         "root:123456@tcp(localhost:3306)/test?timeout=120s&parseTime=true&loc=Local&charset=utf8mb4,utf8",
		Active:      100,
		Idle:        100,
		IdleTimeout: 240,
		LogMode:     false,
	}
}

type User struct {
	ID   uint
	Name string
	//Sex  string
	Age int
}

func TestBaseOrm_FastQuery(t *testing.T) {
	db := NewMySQL(&config)
	//var result []map[interface{}]string
	//var result [][]interface{}
	var result []User
	if err := db.FastQuery("select distinct id, name as name, sex, age from user", &result); err != nil {
		panic(err)
	}
	fmt.Println(result)
}
