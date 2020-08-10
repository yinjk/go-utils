/**
 *
 * @author yinjk
 * @create 2019-05-13 20:17
 */
package list

import (
	"fmt"
	"github.com/yinjk/go-utils/pkg/utils/times"
	"testing"
)

func TestArrayList_Stream(t *testing.T) {
	list := NewArrayListWithValue(1, 2, 3, 4, 5, 6, 6, 7, 5, 6, 9, 8, 16, 14, 11, 13, 46, 23, 523, 3, 5, 6, 723, 5, 613, 5, 32, 556, 14, 516)
	list.Stream().Distinct().Filter(func(t interface{}) bool {
		return t.(int) != 1
	}).Skip(10).Sorted(nil).Limit(10).Maps(func(t interface{}) (r interface{}) {
		return t.(int) * 10
	}).ForEach(func(t interface{}) {
		fmt.Println(t)
	})
}

func TestArrayList_Stream_Count(t *testing.T) {
	list := NewArrayListWithValue(1, 2, 3, 4, 5, 6, 6, 7, 5, 6, 9, 8, 16, 14, 11, 13, 46, 23, 523, 3, 5, 6, 723, 5, 613, 5, 32, 556, 14, 516)
	fmt.Println("the list len: ", list.Size())
	fmt.Println("the stream count: ", list.Stream().Skip(3).Limit(10).Count())
	fmt.Println("the stream min: ", list.Stream().Min(nil))
	fmt.Println("the stream max: ", list.Stream().Max(nil))
	fmt.Println("the stream first: ", list.Stream().Skip(1).FindFirst())
	fmt.Println("the stream last: ", list.Stream().Skip(1).Limit(10).FindLast())
	fmt.Println("the stream any matched -1: ", list.Stream().Skip(0).Limit(10).AnyMatch(func(t interface{}) bool {
		return t == -1
	}))
	fmt.Println("the stream any matched 1: ", list.Stream().Skip(0).Limit(10).AnyMatch(func(t interface{}) bool {
		return t == 1
	}))
	fmt.Println("the stream all matched 1: ", list.Stream().Skip(0).Limit(10).AllMatch(func(t interface{}) bool {
		return t == 1
	}))
	fmt.Println("the stream all matched any: ", list.Stream().Skip(0).Limit(10).AllMatch(func(t interface{}) bool {
		return true
	}))
	fmt.Println("the stream none matched 1: ", list.Stream().Skip(0).Limit(10).NoneMatch(func(t interface{}) bool {
		return t == 1
	}))
	fmt.Println("the stream none matched -1: ", list.Stream().Skip(0).Limit(10).NoneMatch(func(t interface{}) bool {
		return t == -1
	}))
}

func TestPipeline_ToList(t *testing.T) {
	var ints []int
	list := StreamOf(1, 2, 3, 4, 4, 5, 3, 4, 6, 0, 8, 6, 7, 8).Distinct().Sorted(nil).ToList()
	list.ForEach(func(t interface{}) {
		fmt.Println(t)
	})
	StreamOf(1, 2, 4, 5, 6, 7, 8, 89, 90, 5).Sorted(nil).Unmarshal(&ints)
	fmt.Println(ints)
}

type Students struct {
	Name  string
	Score int
}

func Test_efficientTest(t *testing.T) {
	slice := []*Students{{"张三", 1}, {"fds", 2}, {"rwq", 43}, {"rew", 2}, {"张gw三", 100},
		{"张g三", 34}, {"fd", 342}, {"twq", 32}, {"fdsa", 64}, {"gq", 76},
		{"w", 34}, {"fa", 75}, {"tew", 65}, {"fa", 54}, {"gqw", 86},
		{"tw", 65}, {"ewr", 6}, {"ga", 86}, {"张v三", 76}, {"gq", 96},
		{"req", 53}, {"ewr", 53}, {"fdas", 56}, {"张gdas三", 87}, {"张gwq三", 87}}
	list := NewArrayListWithSlice(slice)
	watch := times.NewStopWatch(true)
	for i := 0; i < 1000; i++ {
		useSlice(slice)
	}
	watch.PrettyPrint("useSlice")
	watch.Reset()
	for i := 0; i < 1000; i++ {
		useStream(list)
	}
	watch.PrettyPrint("useStream")

}

func useSlice(slice []*Students) (names []string) {
	disticName := make(map[string]bool)
	for _, stu := range slice {
		if stu.Score < 60 {
			continue
		}
		if _, ok := disticName[stu.Name]; ok {
			continue
		}
		names = append(names, stu.Name)
		disticName[stu.Name] = true
	}
	return
}

func useStream(list List) (names []string) {
	list.Stream().Filter(func(t interface{}) bool {
		return t.(*Students).Score > 60
	}).Maps(func(t interface{}) (r interface{}) {
		return t.(*Students).Name
	}).Distinct().Unmarshal(&names)
	return
}

func TestStreamOf(t *testing.T) {
	StreamOf(1, 2, 3, 9, 4, 5, 6, 8).Filter(func(t interface{}) bool {
		i := t.(int)
		if i == 4 {
			panic("111")
		}
		return true
	}).ForEach(func(t interface{}) {
		fmt.Println(t)
	})
}
