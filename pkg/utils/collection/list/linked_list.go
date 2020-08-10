/*
LinkedList底层是基于连表实现的，它也实现了List接口，和ArrayList实现了相同的接口，所以这两个类具有相同的方法和作用，
但是在使用场景上有些略微的不同，因为LinkedList底层是基于连表来实现的，所以，在随机删除和随机插入时，它不需要移动后面的数据，
从而使得它具有更高的性能，但是在读取和顺序插入时，由于它需要遍历整个集合，所以性能会更差。

使用场景：当我们的集合需要大量的随机插入和删除而不会频繁的读取和遍历时，就可以使用该结构来存储集合。
*/
package list

import (
	"github.com/pkg/errors"
	"reflect"
	"sort"
)

/**
 * @author     ：yinchong
 * @create     ：2019/6/13 9:01
 * @description：
 */
type Node struct {
	preNode  *Node
	nextNode *Node
	value    interface{}
}

//create inner node for linked list
func createNode(prev *Node, value interface{}) *Node {
	return &Node{
		preNode:  prev,
		value:    value,
		nextNode: nil,
	}
}

type LinkedList struct {
	head     *Node
	tail     *Node
	size     int
	lessFunc func(i, j interface{}) bool
}

func NewLinkedList() *LinkedList {
	return &LinkedList{
		head: nil,
		tail: nil,
		size: 0,
	}
}

func (l *LinkedList) Len() int {
	return l.size
}

func (l *LinkedList) Less(i, j int) bool {
	if l.lessFunc == nil {
		return defaultLess(l.Get(i), l.Get(j))
	}
	return l.lessFunc(l.Get(i), l.Get(j))
}

func (l *LinkedList) Swap(i, j int) {
	prev := l.node(i)
	next := l.node(j)
	value := prev.value
	prev.value = next.value
	next.value = value
}

//get value by index if index out of range will return -1
func (l *LinkedList) Get(index int) (value interface{}) {
	l.checkElementIndex(index)
	return l.node(index).value
}

//check value if exist in linkedlist
func (l *LinkedList) GetEq(test func(t interface{}) bool) (value interface{}) {
	if l.size == 0 {
		return true
	}
	if l.size == 1 {
		return test(l.head.value)
	}
	for head, tail := l.head, l.tail; head != tail; head, tail = head.nextNode, tail.preNode {
		if test(head.value) {
			return true
		}
		if test(tail.value) {
			return true
		}
	}
	return false
}

//find value in linkedlist index if not exist return -1
func (l *LinkedList) IndexOf(o interface{}) (index int) {
	if l.size == 0 {
		return -1
	}

	if l.size<<1 >= index {
		tempIndex := 0
		for node := l.head; node != nil; node = node.nextNode {
			if node.value == o {
				return tempIndex
			}
			tempIndex++
		}
		return -1
	}

	tempIndex := l.size - 1
	for node := l.tail; node != nil; node = node.preNode {
		if node.value == o {
			return tempIndex
		}
		tempIndex--
	}
	return -1
}

//add value to linkedlist
func (l *LinkedList) Add(values ...interface{}) {
	if values == nil || len(values) == 0 {
		return
	}
	for _, value := range values {
		last := l.tail
		newNode := createNode(last, value)
		l.tail = newNode
		if last == nil {
			l.head = newNode
		} else {
			last.nextNode = newNode
		}
		l.size++
	}
}

//add slice value to linkedlist
func (l *LinkedList) AddSlice(value interface{}) {
	dest := reflect.Indirect(reflect.ValueOf(value))
	if dest.Kind() != reflect.Slice {
		panic(errors.New("AddSlice must accept a slice type"))
	}
	size := dest.Len()
	for i := 0; i < size; i++ {
		l.Add(dest.Index(i).Interface())
	}
}

func (l *LinkedList) Remove(index int) interface{} {
	l.checkElementIndex(index)
	return l.unlink(l.node(index))
}

//unlink node from linkedlist
func (l *LinkedList) unlink(node *Node) interface{} {
	value := node.value
	next := node.nextNode
	prev := node.preNode
	if prev == nil {
		l.head = next
	} else {
		prev.nextNode = next
		node.preNode = nil
	}

	if next == nil {
		l.tail = prev
	} else {
		next.preNode = prev
		node.nextNode = nil
	}

	node.value = nil
	l.size--
	return value
}

//find node by index
func (l *LinkedList) node(index int) *Node {
	if index < (l.size >> 1) {
		first := l.head
		for i := 0; i < index; i++ {
			first = first.nextNode
		}
		return first
	} else {
		last := l.tail
		for i := l.size - 1; i > index; i-- {
			last = last.preNode
		}
		return last
	}
}

//check element index if exist
func (l *LinkedList) checkElementIndex(index int) {
	if index < 0 || index >= l.size {
		panic("index out of bound")
	}
}

func (l *LinkedList) RemoveValue(val interface{}) bool {
	if l.size == 0 {
		return false
	}
	index := l.IndexOf(val)
	if index == -1 {
		return false
	}
	l.unlink(l.node(index))
	return true
}

func (l *LinkedList) Clear() {
	l.tail = nil
	l.head = nil
	l.size = 0
}

func (l *LinkedList) Size() int {
	return l.size
}

func (l *LinkedList) Range(accept func(i int, value interface{}) (breaks bool)) {
	if l.size == 0 {
		return
	}
	index := 0
	for head := l.head; head != nil; head = head.nextNode {
		if accept(index, head.value) {
			break
		}
		index++
	}
}

func (l *LinkedList) ForEach(accept func(t interface{})) {
	if l.size == 0 {
		return
	}
	for head := l.head; head != nil; head = head.nextNode {
		accept(head.value)
	}
}

func (l *LinkedList) Sort(lessFunc func(o1, o2 interface{}) bool) {
	sort.Sort(l)
}

func (l *LinkedList) Unmarshal(value interface{}) {
	dest := reflect.Indirect(reflect.ValueOf(value))
	if dest.Kind() != reflect.Slice {
		panic(errors.New("Unmarshal must accept a slice type"))
	}
	dest.Set(reflect.MakeSlice(dest.Type(), l.size, l.size))
	l.Range(func(i int, v interface{}) (isBreak bool) {
		elem := reflect.ValueOf(v)
		dest.Index(i).Set(elem)
		return false
	})
}

func (l *LinkedList) ToSlice() []interface{} {
	if l.size == 0 {
		return nil
	}
	len := l.size
	array := make([]interface{}, len)
	l.Range(func(index int, value interface{}) (breaks bool) {
		array[index] = value
		return false
	})
	return array
}

func (l *LinkedList) Stream() Stream {
	return NewPipeline(l)
}
