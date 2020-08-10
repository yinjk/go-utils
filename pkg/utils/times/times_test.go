/**
 *
 * @author yinjk
 * @create 2019-06-24 18:47
 */
package times

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeOutFunc(t *testing.T) {
	err := TimeOutFunc(func() (err error) {
		fmt.Println("你好")
		time.Sleep(time.Second)
		fmt.Println("世界")
		return nil
	}, time.Second)
	if err != nil {
		panic(err)
	}
}

func TestTimeOutWithResult(t *testing.T) {
	result, err := TimeOutWithResult(func() (result interface{}, err error) {
		result = "你好"
		time.Sleep(time.Second)
		result = result.(string) + "世界"
		return result, nil
	}, time.Second, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
