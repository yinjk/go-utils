/**
 *
 * @author yinjk
 * @create 2019-06-14 12:40
 */
package maps

import (
	"fmt"
	"testing"
	"time"
)

//验证并发性
func TestConcurrentHashMap_Put(t *testing.T) {
	chm := NewConcurrentHashMap()
	for i := 0; i < 10000; i++ {
		go func(index int) {
			chm.Put(index, "1")
		}(i)
	}
	time.Sleep(time.Second * 2)
	fmt.Println(chm.Size())
}
