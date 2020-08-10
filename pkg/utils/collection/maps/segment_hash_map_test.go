/**
 *
 * @author yinjk
 * @create 2019-06-19 10:26
 */
package maps

import (
	"fmt"
	"github.com/yinjk/go-utils/pkg/utils/times"
	"strconv"
	"sync"
	"testing"
)

func TestMap_Size(t *testing.T) {
	segment := &Segment{}
	fmt.Println("point 1:", hash(segment))
	fmt.Println("point 2:", hash(segment))
	fmt.Println("point 3:", hash(segment))

	fmt.Println("struct 1:", hash(&Segment{}))
	fmt.Println("struct 2:", hash(&Segment{}))
	fmt.Println("struct 3:", hash(&Segment{}))

	fmt.Println("other struct 1:", hash(&Segment{}))
	fmt.Println("other struct 2:", hash(&ConcurrentHashMap{}))
	fmt.Println("other struct 3:", hash(&SegmentHashMap{}))

	fmt.Println("int 1:", hash(1))
	fmt.Println("int 1:", hash(1))
	fmt.Println("int 2:", hash(2))
	fmt.Println("string 1:", hash("hello"))
	fmt.Println("string 1:", hash("hello"))
	fmt.Println("string 2:", hash("nihao"))
	fmt.Println("bool 1:", hash(true))
	fmt.Println("bool 2:", hash(false))
}

func TestSegmentHashMap_Put(t *testing.T) {
	var shm Map
	group := sync.WaitGroup{}
	group.Add(10000)
	shm = NewSegmentHashMap()
	for i := 0; i < 10000; i++ {
		go func(index int) {
			shm.Put(index, "1")
			group.Done()
		}(i)
	}
	group.Wait()
	fmt.Println(shm.Size())
}

/**
 * 两种同步map性能比较测试
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/6/19 13:11
 */
func Test_Compare(t *testing.T) {
	var (
		goroutineCount = 1000 // 开启的协程数
		valueCount     = 100  // 每个协程put的元素个数
	)
	putTest(NewConcurrentHashMap(), goroutineCount, valueCount, "concurrentHashMap")
	putTest(NewSegmentHashMap(), goroutineCount, valueCount, "segmentHashMap")

}

func putTest(maps Map, goroutineCount, valueCount int, msg string) {
	group := sync.WaitGroup{}
	group.Add(goroutineCount)
	watch := times.NewStopWatch(true)
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			for j := 0; j < valueCount; j++ {
				key := strconv.Itoa(index) + "-" + strconv.Itoa(j)
				maps.Put(key, true)
			}
			group.Done()
		}(i)
	}
	group.Wait()
	fmt.Println("map put complete, and the result size is: ", maps.Size())
	watch.PrettyPrint(msg)
}
