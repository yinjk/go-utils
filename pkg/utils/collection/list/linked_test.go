package list

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

/**
 * @author     ：yinchong
 * @create     ：2019/6/13 9:34
 * @description：
 */
func Test(t *testing.T) {
	linkedList := NewLinkedList()
	linkedList.Add("test")
	linkedList.Add("hello")
	linkedList.Add("nice")
	linkedList.Add("good")
	value := linkedList.IndexOf("hello")
	fmt.Println(value)
	flag := linkedList.RemoveValue("hello")
	fmt.Println(flag)
	linkedList.Clear()
	value = linkedList.IndexOf("hello")
	fmt.Println(value)
	//linkedList.ForEach(func(t interface{}) {
	//	fmt.Println(t)
	//})
}

func TestArrayList_Add(t *testing.T) {
	parse := resource.MustParse("12")
	fmt.Println(parse.Value())
}
