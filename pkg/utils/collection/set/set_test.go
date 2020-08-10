/**
 *
 * @author yinjk
 * @create 2019-03-12 20:01
 */
package set

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestNewHashSet(t *testing.T) {
	set := NewHashSet()
	set.Add("abc")
	set.Add("adc")
	set.Add("abc")
	set.Add("adc")
	set.Add("adc")
	set.Add("afc")
	fmt.Println(set.Contains("amc"))
	fmt.Println(set.Contains("adc"))
	fmt.Println(set)
	for _, v := range set.Elements() {
		fmt.Println(v)
	}
}

func TestNextIntN(t *testing.T) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
	}
}

type Student struct {
}

type Teacher struct {
}

type Manager struct {
}

func Test_Reflect(t *testing.T) {
	fmt.Println(reflect.TypeOf(&Student{}))
	fmt.Println(reflect.TypeOf(&Student{}))

	st1 := reflect.TypeOf(&Student{})
	st2 := reflect.TypeOf(&Student{})
	te1 := reflect.TypeOf(&Teacher{})
	te2 := reflect.TypeOf(&Teacher{})
	fmt.Println(st1 == st2)
	fmt.Println(te1 == te2)
	fmt.Println(st1 == te1)

}
