/**
 * 底层基于数组实现，在添加元素的性能上略高于原生slice，性能提升大概1.5倍
 * 相对于LinkedList来说，在遍历、随机访问、顺序添加元素上有更好的性能表现，
 * 同样的，在随机删除、随机插入上，LinkedList有更好的性能表现
 *
 * ArrayList的设计是借鉴了java的集合类，封装了一些非常便利的方法，如快速查找、添加、删除元素等，并且提供了向Stream流转换的方法。
 * 适用场景：所有对集合的复杂操作都可以考虑使用该类来替换原生的slice切片。
 * @author yinjk
 * @create 2019-05-07 17:43
 */
package list

import (
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
)

const (
	_defaultCapacity = 10
	_maxArraySize    = 1<<31 - 9
)

type ArrayList struct {
	internalSlice []interface{}
	size          int
	lessFun       func(o1, o2 interface{}) bool
}

func NewArrayListWithValue(value ...interface{}) (list *ArrayList) {
	internalSlice := value
	return &ArrayList{
		internalSlice: internalSlice,
		size:          len(value),
	}
}

func NewArrayListWithSlice(value interface{}) (list *ArrayList) {
	dest := reflect.Indirect(reflect.ValueOf(value))
	if dest.Kind() != reflect.Slice {
		panic(errors.New("NewArrayListWithSlice must accept a slice type"))
	}
	size := dest.Len()
	list = NewArrayListWithCapacity(size)
	for i := 0; i < size; i++ {
		list.Add(dest.Index(i).Interface())
	}
	return list
}

func NewArrayListWithCapacity(initialCapacity int) (list *ArrayList) {
	if initialCapacity <= 0 {
		initialCapacity = _defaultCapacity
	}
	return &ArrayList{
		internalSlice: make([]interface{}, initialCapacity, initialCapacity),
		size:          0,
	}
}

func NewArrayList(option ...Option) *ArrayList {
	if option == nil || len(option) == 0 {
		return NewArrayListWithCapacity(_defaultCapacity)
	}
	o := &options{}
	apply(o, option)
	if o.capacity > 0 {
		return NewArrayListWithCapacity(o.capacity)
	}
	if o.slice != nil {
		return NewArrayListWithSlice(o.slice)
	}
	if o.values != nil {
		return NewArrayListWithValue(o.values...)
	}
	return NewArrayListWithCapacity(_defaultCapacity)
}

func EmptyList() *ArrayList {
	return NewArrayListWithCapacity(0)
}

// Get one element by index
func (l *ArrayList) Get(index int) (value interface{}) {
	if index >= l.size || index < 0 {
		panic(errors.New("index out of size, Index: " + strconv.Itoa(index) + ", Size: " + strconv.Itoa(l.size)))
	}
	return l.internalSlice[index]
}

// Get one element by test func
func (l *ArrayList) GetEq(test func(t interface{}) bool) (value interface{}) {
	l.ForEach(func(t1 interface{}) {
		if test(t1) {
			value = t1
		}
	})
	return value
}

// get the value index in list, if not found will return -1
func (l *ArrayList) IndexOf(o interface{}) (index int) {
	for i := 0; i < l.size; i++ {
		if o == l.internalSlice[i] {
			return i
		}
	}
	return -1
}

// add values
func (l *ArrayList) Add(values ...interface{}) {
	l.ensureCapacityInternal(l.size + len(values)) // Sure the size of array is enough!!
	for _, value := range values {
		l.internalSlice[l.size] = value
		l.size++
	}
}

// add a slice to list, the main purpose is to be compatible with slice
func (l *ArrayList) AddSlice(value interface{}) {
	dest := reflect.Indirect(reflect.ValueOf(value))
	if dest.Kind() != reflect.Slice {
		panic(errors.New("AddSlice must accept a slice type"))
	}
	size := dest.Len()
	for i := 0; i < size; i++ {
		l.Add(dest.Index(i).Interface())
	}
}

// remove one element by index
func (l *ArrayList) Remove(index int) (value interface{}) {
	if index >= l.size {
		panic(errors.New("index out of size, Index: " + strconv.Itoa(index) + ", Size: " + strconv.Itoa(l.size)))
	}
	old := l.internalSlice[index] // Get the element we want to delete
	for i := index; i < l.size; i++ {
		if i == l.size-1 { // Last element set nil to let GC do its work
			l.internalSlice[i] = nil
		}
		l.internalSlice[i] = l.internalSlice[i+1] //Move all elements one bit forward after index
	}
	l.size--
	return old
}

// clear list element and fast to gc
func (l *ArrayList) Clear() {
	// clear to let GC do its work
	for i := 0; i < l.size; i++ {
		l.internalSlice[i] = nil
	}
	l.size = 0
}

// remove one value in list
func (l *ArrayList) RemoveValue(val interface{}) (removed bool) {
	if val == nil {
		return false
	}
	index := -1
	l.Range(func(i int, value interface{}) (isBreak bool) {
		if val == value {
			index = i
			return true
		}
		return false
	})
	if index == -1 {
		return false
	}
	l.Remove(index)
	return true
}

// return the list size
func (l *ArrayList) Size() int {
	return l.size
}

// range the list and can break the range when the iter func return true value
func (l *ArrayList) Range(accept func(i int, value interface{}) (isBreak bool)) {
	for i := 0; i < l.size; i++ {
		if accept(i, l.internalSlice[i]) {
			break
		}
	}
}

func (l *ArrayList) ForEach(accept func(t interface{})) {
	for i := 0; i < l.size; i++ {
		accept(l.internalSlice[i])
	}
}

func (l *ArrayList) Sort(lessFun func(o1, o2 interface{}) bool) {
	if lessFun != nil {
		l.lessFun = lessFun
	}
	sort.Sort(l)
}

// unmarshal the list to slice
func (l *ArrayList) Unmarshal(value interface{}) {
	dest := reflect.Indirect(reflect.ValueOf(value))
	if dest.Kind() != reflect.Slice {
		panic(errors.New("Unmarshal must accept a slice type"))
	}
	dest.Set(reflect.MakeSlice(dest.Type(), l.size, l.size))
	l.Range(func(i int, v interface{}) (isBreak bool) {
		elem := reflect.ValueOf(v)
		dest.Index(i).Set(elem)
		return false
		//dest.Set(reflect.Append(dest, elem))
	})
}

// get the ArrayList internal slice struct with interface{}
func (l *ArrayList) ToSlice() []interface{} {
	return l.internalSlice[:l.size]
}

func (l *ArrayList) Len() int {
	return l.size
}

func (l *ArrayList) Less(i, j int) bool {
	if l.lessFun == nil {
		return defaultLess(l.internalSlice[i], l.internalSlice[j])
	}
	return l.lessFun(l.internalSlice[i], l.internalSlice[j])
}

func (l *ArrayList) Swap(i, j int) {
	temp := l.internalSlice[i]
	l.internalSlice[i] = l.internalSlice[j]
	l.internalSlice[j] = temp
}

func (l *ArrayList) Stream() Stream {
	return NewPipeline(l)
}

func (l *ArrayList) ensureCapacityInternal(minCapacity int) {
	if minCapacity-len(l.internalSlice) > 0 {
		l.grow(minCapacity)
	}

}

func (l *ArrayList) grow(minCapacity int) {
	oldCapacity := len(l.internalSlice)
	newCapacity := oldCapacity + (oldCapacity >> 1)
	if newCapacity-minCapacity < 0 {
		newCapacity = minCapacity
	}
	if newCapacity-_maxArraySize > 0 {
		newCapacity = _maxArraySize
	}
	newSlice := make([]interface{}, newCapacity, newCapacity)
	copy(newSlice, l.internalSlice)
	l.internalSlice = newSlice
}
