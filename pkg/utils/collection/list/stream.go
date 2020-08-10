/**
 * 流式处理接口
 * @author yinjk
 * @create 2019-05-13 18:01
 */
package list

type Stream interface {
	Distinct() Stream

	Filter(test func(t interface{}) bool) Stream

	Skip(n int) Stream

	Limit(maxSize int) Stream

	Sorted(lessFun func(o1, o2 interface{}) bool) Stream

	Maps(apply func(t interface{}) (r interface{})) Stream

	Peek(accept func(t interface{})) Stream

	//terminal operation

	Count() int

	Min(lessFun func(o1, o2 interface{}) bool) interface{}

	Max(lessFun func(o1, o2 interface{}) bool) interface{}

	FindFirst() interface{}

	FindLast() interface{}

	AnyMatch(test func(t interface{}) bool) bool

	AllMatch(test func(t interface{}) bool) bool

	NoneMatch(test func(t interface{}) bool) bool

	ToList() List

	ForEach(consumer func(t interface{}))

	Unmarshal(slice interface{})
}
