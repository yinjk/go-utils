// @Desc
// @Author  inori
// @Update
package list

import (
	"fmt"
	"strconv"
	"testing"
)

func TestStreamOf(_ *testing.T) {
	s := StreamOf("1", "2", "3", "8", "5", "4", "7", "10", "2", "10", "1").
		Distinct()
	res := Map(s, func(i string) int64 {
		n, _ := strconv.ParseInt(i, 10, 64)
		return n
	}).Filter(func(t int64) bool {
		return t%2 == 0
	}).Sorted(func(o1, o2 int64) bool {
		return o1 > o2
	}).ToArray()
	fmt.Println(res)
}
