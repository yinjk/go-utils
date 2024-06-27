/*
流式处理接口的一种简单实现，该接口提供了一种基于流式处理的功能，而该实现是参考了java的Stream流框架的源码，
并且通过go的channel管道使用扇入扇出的特性来实现，而pipeline中的sink函数则可以理解成我们的处理函数，数据通过
入参的管道流进来，处理之后通过出参的管道流出去，每个操作都会生成一个sink函数，而这些sink处理函数最终会连成一个链条，使得数据
从一个sink流出之后能流入下一个sink，

	流入管道   —————————   流出管道  —————————   流出管道

原始数据 =========| 去重处理 |==========| 过滤处理 |========== 最终数据

	—————————            —————————

如上所示，每一个处理都是一个sink函数，而sink函数的入参用于接收流入管道中的数据，而出参这是一个处理之后的数据流管道，
这个管道会作为下一个sink函数的入参，这样数据流就串联起来了

使用场景：当我们的集合类处理一些复杂场景的时候会比较吃力，而使用流式处理却能轻易的实现，例如:我们需要做一个分页的功能，如果使用集合类我们需要
从原始集合中算出当前页的数据，并将这些数据复制到新的集合中去，如果要对数据排序，又需要在新的集合中去排序，这时候假如我们又只需要原始数据中的一个属性，
我们又需要遍历这个新的集合并且取出我们需要的属性，并把这些新的属性又放在一个新的集合中，这样的操作就会显得很复杂，而如果我们使用流式处理就会很轻易的
处理这样的逻辑，比如，先使用skip和limit函数分好页再调用sorted函数对分页之后的数据排序，最后调用map函数将排好序的数据流转换成我们需要的数据格式即可。
但是需要注意的一点是，每一个处理操作都会对应一次数据流的流动，所以应该尽量减少流式调用的层数，能在一个处理中完成的事就别调用两次来完成，这回带来较大的性能损耗。

	@author yinjk
	@create 2019-05-13 19:40
*/
package list

import (
	"sort"
)

type sourceConsumer[T comparable] func(out chan<- T)
type limitedPipeline[T comparable] struct {
	source   sourceConsumer[T]
	head     *limitedPipeline[T]
	previous *limitedPipeline[T]
	next     *limitedPipeline[T]

	sink  func(in chan T) (out chan T) //该方法实现扇入扇出流式处理，每次stream调用，数据都会从一个通道流入下一个通道
	depth int
}

func newLimitedPipeline[T comparable](source []T) (p *limitedPipeline[T]) {
	p = &limitedPipeline[T]{
		depth: 0,
	}
	p.source = func(out chan<- T) {
		for _, t := range source {
			out <- t
		}
	}
	p.head = p
	return
}
func newLimitedPipelineFromCh[T comparable](source <-chan T) (p *limitedPipeline[T]) {
	p = &limitedPipeline[T]{
		depth: 0,
	}
	p.source = func(out chan<- T) {
		for t := range source {
			out <- t
		}
	}
	p.head = p
	return
}
func newPipelineFromPreview[T comparable](preview *limitedPipeline[T]) (p *limitedPipeline[T]) {
	p = &limitedPipeline[T]{
		previous: preview,
		head:     preview.head,
		depth:    preview.depth + 1,
	}
	preview.next = p
	return
}

func StreamOf[T comparable](value ...T) Stream[T] {
	return newLimitedPipeline(value)
}

func StreamOfSlice[T comparable](slice []T) Stream[T] {
	return newLimitedPipeline(slice)
}

func (p *limitedPipeline[T]) isBatch() bool {
	return true
}

// Distinct  Returns a stream consisting of the distinct elements of this stream.
func (p *limitedPipeline[T]) Distinct() Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		out := make(chan T)
		go func() {
			distinctMap := make(map[T]bool)
			for t := range in {
				if _, ok := distinctMap[t]; ok {
					continue
				}
				distinctMap[t] = true
				out <- t
			}
			close(out)
		}()
		return out
	}
	return current
}

// Distinct  Returns a stream consisting of the distinct elements of this stream.
func (p *limitedPipeline[T]) DistinctComplex(equals func(o1, o2 T) bool) Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		out := make(chan T)
		go func() {
			distinctMap := make(map[T]bool)
			for t := range in {
				if _, ok := distinctMap[t]; ok {
					continue
				}
				distinctMap[t] = true
				out <- t
			}
			close(out)
		}()
		return out
	}
	return current
}

// Filter Returns a stream consisting of the elements of this stream that match the given test func.
func (p *limitedPipeline[T]) Filter(test func(v T) bool) Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		out := make(chan T)
		go func() {
			for t := range in {
				if test(t) {
					out <- t
				}
			}
			close(out)
		}()
		return out
	}
	return current
}

// Skip Returns a stream consisting of the remaining elements of this stream
//
//	after discarding the first n elements of the stream.
//	If this stream contains fewer than n elements then an
//	empty stream will be returned.
func (p *limitedPipeline[T]) Skip(n int) Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		if n < 0 {
			panic("stream.skip args: [n] must to >= 0")
		}
		if n == 0 {
			return in
		}
		out := make(chan T)
		go func() {
			index := 0
			for t := range in {
				if index < n { //跳过
					index++
					continue
				}
				out <- t
			}
			close(out)
		}()
		return out
	}
	return current
}

// Limit Returns a stream consisting of the elements of this stream, truncated to be no longer than maxSize in length.
func (p *limitedPipeline[T]) Limit(maxSize int) Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		out := make(chan T)
		go func() {
			index := 0
			for t := range in {
				if index >= maxSize {
					//break
					continue
				}
				index++
				out <- t
			}
			close(out)
		}()
		return out
	}
	return current
}

// Sorted Returns a stream consisting of the elements of this stream, sorted according to natural order,
// it will use default lessFunc to sort the original element if the lessFunc is nil.
func (p *limitedPipeline[T]) Sorted(lessFunc func(o1, o2 T) bool) Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		out := make(chan T)
		go func() {
			l := make([]T, 0)
			for t := range in {
				l = append(l, t)
			}
			sort.SliceStable(l, func(i, j int) bool {
				return lessFunc(l[i], l[j])
			})
			for _, t := range l {
				out <- t
			}
			close(out)
		}()
		return out
	}
	return current
}

// Peek Like ForEach() function, but this func will return a Stream consisting of the all elements
func (p *limitedPipeline[T]) Peek(accept func(t T)) Stream[T] {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) chan T {
		out := make(chan T)
		go func() {
			for t := range in {
				accept(t)
				out <- t
			}
			close(out)
		}()
		return out
	}
	return current
}

// ========== terminal operation ==========

// Count return the count of elements in this stream
func (p *limitedPipeline[T]) Count() int {
	current := newPipelineFromPreview(p)
	count := 0
	current.sink = func(in chan T) (out chan T) {
		for range in {
			count++
		}
		return nil
	}
	p.handSink()
	return count
}

// Min Returns the minimum element of this stream according to the provided lessFunc,
// if the lessFunc is nil will use the default func
func (p *limitedPipeline[T]) Min(lessFunc func(o1, o2 T) bool) T {
	current := newPipelineFromPreview(p)
	var minimum T
	current.sink = func(in chan T) (out chan T) {
		index := 0
		for t := range in {
			if index == 0 {
				minimum = t
				index++
				continue
			}
			if lessFunc(t, minimum) {
				minimum = t
			}
			index++
		}
		return nil
	}
	p.handSink()
	return minimum
}

// Max Returns the maximum element of this stream according to the provided lessFun,
// if the lessFunc is nil will use the default func
func (p *limitedPipeline[T]) Max(lessFunc func(o1, o2 T) bool) T {
	current := newPipelineFromPreview(p)
	var maximum T
	current.sink = func(in chan T) (out chan T) {
		index := 0
		for t := range in {
			if index == 0 {
				maximum = t
				index++
				continue
			}
			if lessFunc(maximum, t) {
				maximum = t
			}
			index++
		}
		return nil
	}
	p.handSink()
	return maximum
}

// FindFirst Returns the first element of this stream
func (p *limitedPipeline[T]) FindFirst() T {
	current := newPipelineFromPreview(p)
	var first T
	current.sink = func(in chan T) (out chan T) {
		first = <-in
		return nil
	}
	p.handSink()
	return first
}

// FindLast Returns the last element of this stream
func (p *limitedPipeline[T]) FindLast() T {
	current := newPipelineFromPreview(p)
	var last T
	current.sink = func(in chan T) (out chan T) {
		for t := range in {
			last = t
		}
		return nil
	}
	p.handSink()
	return last
}

// AnyMatch Returns whether any elements of this stream match the provided test func
func (p *limitedPipeline[T]) AnyMatch(test func(t T) bool) bool {
	current := newPipelineFromPreview(p)
	matched := false
	current.sink = func(in chan T) (out chan T) {
		for t := range in {
			if test(t) {
				matched = true
				return
			}
		}
		return nil
	}
	p.handSink()
	return matched
}

// AllMatch Returns whether all elements of this stream match the provided test func
func (p *limitedPipeline[T]) AllMatch(test func(t T) bool) bool {
	current := newPipelineFromPreview(p)
	matched := true
	current.sink = func(in chan T) (out chan T) {
		for t := range in {
			if !test(t) {
				matched = false
				return
			}
		}
		return nil
	}
	p.handSink()
	return matched
}

// NoneMatch Returns whether no elements of this stream match the provided test func
func (p *limitedPipeline[T]) NoneMatch(test func(t T) bool) bool {
	current := newPipelineFromPreview(p)
	matched := false
	current.sink = func(in chan T) (out chan T) {
		for t := range in {
			if test(t) {
				matched = true
				return
			}
		}
		return nil
	}
	p.handSink()
	return !matched
}

// ForEach Performs an action for each element of this stream.
func (p *limitedPipeline[T]) ForEach(consumer func(t T)) {
	current := newPipelineFromPreview(p)
	current.sink = func(in chan T) (out chan T) {
		for t := range in {
			consumer(t)
		}
		return nil
	}
	p.handSink()
}

// ToArray to array list
func (p *limitedPipeline[T]) ToArray() []T {
	var res = make([]T, 0)
	p.ForEach(func(t T) {
		res = append(res, t)
	})
	return res
}

// handSink Do execute the limitedPipeline chain
func (p *limitedPipeline[T]) handSink() {
	var out chan T
	out = make(chan T)
	go func() {
		p.head.source(out)
		close(out)
	}()

	var outLink chan T
	sourceStage := p.head
	for current := sourceStage.next; current != nil; current = current.next {
		if current == sourceStage.next {
			outLink = current.sink(out)
			continue
		}
		outLink = current.sink(outLink)
	}
}
