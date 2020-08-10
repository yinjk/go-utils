/**
 *
 * @author yinjk
 * @create 2019-05-07 17:46
 */
package list

import (
	"fmt"
	"github.com/yinjk/go-utils/pkg/utils/times"
	"sort"
	"testing"
)

func TestList_Add(t *testing.T) {
	watch := times.NewStopWatch(true)
	list := NewArrayListWithCapacity(10)
	for i := 0; i < 10000000; i++ {
		list.Add(i)
	}
	var r []int
	watch.PrettyPrint("add")
	list.Unmarshal(&r)
	watch.PrettyPrint("total")
	fmt.Println(list.Size())
}

func TestSlice_Add(t *testing.T) {
	watch := times.NewStopWatch(true)
	list := make([]interface{}, 0, 10)
	for i := 0; i < 10000000; i++ {
		list = append(list, i)
	}
	watch.PrettyPrint("total")
	fmt.Println(len(list))
}

type Student struct {
	Age int
}

func (s Student) LessTo(o interface{}) bool {
	if s2, ok := o.(Student); ok {
		return s.Age < s2.Age
	}
	if s2, ok := o.(*Student); ok {
		return s.Age < s2.Age
	}
	return false
}

func TestList_Add1(t *testing.T) {
	list := NewArrayList()
	var remove *Student
	for i := 0; i < 16; i++ {
		student := &Student{Age: i}
		list.Add(student)
		if i == 3 {
			remove = student
		}
	}
	list.Remove(11)
	list.RemoveValue(remove)
	list.RemoveValue(nil)
	var r []*Student
	list.Unmarshal(&r)
	for _, v := range r {
		fmt.Println(v)
	}
}

func TestNewArrayList(t *testing.T) {
	list1 := NewArrayList(Values(1, 2, 3, 4, 5, 6, 778, 2, 8, 9, 1))
	fmt.Println(list1.Size())
	slice := []string{"1", "2", "2", "3", "4", "7"}
	list2 := NewArrayList(Slice(&slice))
	fmt.Println(list2.Size())

}

func TestNewArrayListWithSlice(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	list := NewArrayListWithSlice(a)
	list.Range(func(i int, value interface{}) (isBreak bool) {
		fmt.Println(value)
		return false
	})
}

func TestArrayList_Sort(t *testing.T) {
	list := NewArrayListWithValue("a", "e", "d", "c", "b")
	sort.Sort(list)
	list.Range(func(i int, value interface{}) (isBreak bool) {
		fmt.Println(value)
		return false
	})

	list2 := NewArrayListWithValue(&Student{Age: 3}, &Student{Age: 2}, &Student{Age: 1}, &Student{Age: 5}, &Student{Age: 9})
	//list2.Sort(func(o1, o2 interface{}) bool {
	//	student1 := o1.(*Student)
	//	student2 := o2.(*Student)
	//	return student1.Age < student2.Age
	//})
	list2.Sort(nil)
	list2.Range(func(i int, value interface{}) (isBreak bool) {
		fmt.Println(value)
		return false
	})
}
