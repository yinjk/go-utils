/**
 * list集合类统一接口
 * @author yinjk
 * @create 2019-05-07 22:55
 */
package list

import (
	"sort"
)

type options struct {
	capacity int
	slice    interface{}
	values   []interface{}
}

type Option func(o *options)

func Capacity(c int) Option {
	return func(o *options) {
		o.capacity = c
	}
}
func Slice(s interface{}) Option {
	return func(o *options) {
		o.slice = s
	}
}
func Values(v ...interface{}) Option {
	return func(o *options) {
		o.values = v
	}
}
func apply(o *options, opts []Option) {
	for _, f := range opts {
		f(o)
	}
}

type List interface {
	sort.Interface

	Get(index int) (value interface{})

	GetEq(test func(t interface{}) bool) (value interface{})

	IndexOf(o interface{}) (index int)

	Add(values ...interface{})

	AddSlice(value interface{})

	Remove(index int) interface{}

	RemoveValue(val interface{}) bool

	Clear()

	Size() int

	Range(accept func(i int, value interface{}) (breaks bool))

	ForEach(accept func(t interface{}))

	Sort(lessFunc func(o1, o2 interface{}) bool)

	Unmarshal(value interface{})

	ToSlice() []interface{}

	Stream() Stream
}

type Comparable interface {
	LessTo(o interface{}) bool
}

// the default less to sort
func defaultLess(o1, o2 interface{}) bool {
	switch o1.(type) {
	case Comparable:
		c1 := o1.(Comparable)
		return c1.LessTo(o2)
	case int:
		i1 := o1.(int)
		i2 := o2.(int)
		return i1 < i2
	case int8:
		i1 := o1.(int8)
		i2 := o2.(int8)
		return i1 < i2
	case int16:
		i1 := o1.(int16)
		i2 := o2.(int16)
		return i1 < i2
	case int32:
		i1 := o1.(int32)
		i2 := o2.(int32)
		return i1 < i2
	case int64:
		i1 := o1.(int64)
		i2 := o2.(int64)
		return i1 < i2
	case float32:
		f1 := o1.(float32)
		f2 := o2.(float32)
		return f1 < f2
	case float64:
		f1 := o1.(float64)
		f2 := o2.(float64)
		return f1 < f2
	case string:
		s1 := o1.(string)
		s2 := o2.(string)
		return s1 < s2
	case uint:
		i1 := o1.(uint)
		i2 := o2.(uint)
		return i1 < i2
	case uint8:
		i1 := o1.(uint8)
		i2 := o2.(uint8)
		return i1 < i2
	case uint16:
		i1 := o1.(uint16)
		i2 := o2.(uint16)
		return i1 < i2
	case uint32:
		i1 := o1.(uint32)
		i2 := o2.(uint32)
		return i1 < i2
	case uint64:
		i1 := o1.(uint64)
		i2 := o2.(uint64)
		return i1 < i2
	default:
		panic("nil point exception, the  compare func is nil, to set it when use sort on list")
	}
}
