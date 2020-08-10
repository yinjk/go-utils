/**
 *
 * @author yinjk
 * @create 2019-06-21 18:48
 */
package convert

import (
	"fmt"
	"testing"
)

type Student struct {
	Name string
	Age  int
	Sex  string
}

type StudentCopy struct {
	Name string
	Age  int
	Sex  int
}

func TestCopyProperties(t *testing.T) {
	var target *StudentCopy = &StudentCopy{}

	CopyProperties(&Student{Name: "zhangshan", Sex: "man", Age: 18}, target, "age")
	fmt.Println(target)
}
