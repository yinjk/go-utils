/*
 @Desc

 @Date 2020-06-23 21:02
 @Author yinjk
*/
package syncs

import (
	"fmt"
	"testing"
	"time"
)

func TestExecutor(_ *testing.T) {
	executor := NewTaskExecutor(10, 1000, nil)
	for i := 0; i < 1000; i++ {
		executor.Execute(func() {
			time.Sleep(time.Second)
			fmt.Println(i)
		})
	}
	time.Sleep(time.Second * 20)
	fmt.Println("shutdown executor >>>>>>>>>>>>>>>>>>>>")
	executor.Shutdown()
	time.Sleep(time.Second * 20)

}
