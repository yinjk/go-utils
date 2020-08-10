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
	"errors"
	"reflect"
)

type pipeline struct {
	sourceList    List
	sourceStage   *pipeline
	previousStage *pipeline
	nextStage     *pipeline
	sink          func(in chan interface{}) (out chan interface{}) //该方法实现扇入扇出流式处理，每次stream调用，数据都会从一个通道流入下一个通道

	depth int
}

func NewPipeline(source List) (p *pipeline) {
	p = &pipeline{
		sourceList: source,
		depth:      0,
	}
	p.sourceStage = p
	return
}
func pipelineWithPre(previousStage *pipeline) (p *pipeline) {
	p = &pipeline{
		previousStage: previousStage,
		sourceStage:   previousStage.sourceStage,
		sourceList:    previousStage.sourceList,
		depth:         previousStage.depth + 1,
	}
	previousStage.nextStage = p
	return
}

func StreamOf(value ...interface{}) Stream {
	l := NewArrayListWithValue(value...)
	return NewPipeline(l)
}

func StreamOfSlice(slice interface{}) Stream {
	l := NewArrayListWithSlice(slice)
	return NewPipeline(l)
}

//Distinct  Returns a stream consisting of the distinct elements of this stream.
func (p *pipeline) Distinct() Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
		go func() {
			distinctMap := make(map[interface{}]bool)
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

//Distinct  Returns a stream consisting of the distinct elements of this stream.
func (p *pipeline) DistinctComplex(equals func(o1, o2 interface{}) bool) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
		go func() {
			distinctMap := make(map[interface{}]bool)
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

//Filter Returns a stream consisting of the elements of this stream that match the given test func.
func (p *pipeline) Filter(test func(t interface{}) bool) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
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

//Skip Returns a stream consisting of the remaining elements of this stream
//     after discarding the first n elements of the stream.
//     If this stream contains fewer than n elements then an
//     empty stream will be returned.
func (p *pipeline) Skip(n int) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		if n < 0 {
			panic("stream.skip args: [n] must to >= 0")
		}
		if n == 0 {
			return in
		}
		out := make(chan interface{})
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

//Limit Returns a stream consisting of the elements of this stream, truncated to be no longer than maxSize in length.
func (p *pipeline) Limit(maxSize int) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
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

//Sorted Returns a stream consisting of the elements of this stream, sorted according to natural order,
// it will use default lessFunc to sort the original element if the lessFunc is nil.
func (p *pipeline) Sorted(lessFunc func(o1, o2 interface{}) bool) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
		go func() {
			l := NewArrayList()
			for t := range in {
				l.Add(t)
			}
			l.Sort(lessFunc)
			l.ForEach(func(t interface{}) {
				out <- t
			})
			close(out)
		}()
		return out
	}
	return current
}

//Maps Returns a stream consisting of the results of applying the given function to the elements of this stream.
func (p *pipeline) Maps(apply func(t interface{}) (r interface{})) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
		go func() {
			for t := range in {
				out <- apply(t)
			}
			close(out)
		}()
		return out
	}
	return current
}

//Peek Like ForEach() function, but this func will return a Stream consisting of the all elements
func (p *pipeline) Peek(accept func(t interface{})) Stream {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) chan interface{} {
		out := make(chan interface{})
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

//Count return the count of elements in this stream
func (p *pipeline) Count() int {
	current := pipelineWithPre(p)
	count := 0
	current.sink = func(in chan interface{}) (out chan interface{}) {
		for range in {
			count++
		}
		return nil
	}
	p.handSink()
	return count
}

//Min Returns the minimum element of this stream according to the provided lessFunc,
//if the lessFunc is nil will use the default func
func (p *pipeline) Min(lessFunc func(o1, o2 interface{}) bool) interface{} {
	if lessFunc == nil {
		lessFunc = defaultLess
	}
	current := pipelineWithPre(p)
	var minimum interface{}
	current.sink = func(in chan interface{}) (out chan interface{}) {
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

//Max Returns the maximum element of this stream according to the provided lessFun,
//if the lessFunc is nil will use the default func
func (p *pipeline) Max(lessFunc func(o1, o2 interface{}) bool) interface{} {
	if lessFunc == nil {
		lessFunc = defaultLess
	}
	current := pipelineWithPre(p)
	var maximum interface{}
	current.sink = func(in chan interface{}) (out chan interface{}) {
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

//FindFirst Returns the first element of this stream
func (p *pipeline) FindFirst() interface{} {
	current := pipelineWithPre(p)
	var first interface{}
	current.sink = func(in chan interface{}) (out chan interface{}) {
		first = <-in
		return nil
	}
	p.handSink()
	return first
}

//FindLast Returns the last element of this stream
func (p *pipeline) FindLast() interface{} {
	current := pipelineWithPre(p)
	var last interface{}
	current.sink = func(in chan interface{}) (out chan interface{}) {
		for t := range in {
			last = t
		}
		return nil
	}
	p.handSink()
	return last
}

//AnyMatch Returns whether any elements of this stream match the provided test func
func (p *pipeline) AnyMatch(test func(t interface{}) bool) bool {
	current := pipelineWithPre(p)
	matched := false
	current.sink = func(in chan interface{}) (out chan interface{}) {
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

//AllMatch Returns whether all elements of this stream match the provided test func
func (p *pipeline) AllMatch(test func(t interface{}) bool) bool {
	current := pipelineWithPre(p)
	matched := true
	current.sink = func(in chan interface{}) (out chan interface{}) {
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

//NoneMatch Returns whether no elements of this stream match the provided test func
func (p *pipeline) NoneMatch(test func(t interface{}) bool) bool {
	current := pipelineWithPre(p)
	matched := false
	current.sink = func(in chan interface{}) (out chan interface{}) {
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

//ToList Returns the list type of this stream, this func is a terminal operation
func (p *pipeline) ToList() List {
	current := pipelineWithPre(p)
	list := NewArrayList()
	current.sink = func(in chan interface{}) (out chan interface{}) {
		for t := range in {
			list.Add(t)
		}
		return nil
	}
	p.handSink()
	return list
}

//ForEach Performs an action for each element of this stream.
func (p *pipeline) ForEach(consumer func(t interface{})) {
	current := pipelineWithPre(p)
	current.sink = func(in chan interface{}) (out chan interface{}) {
		for t := range in {
			consumer(t)
		}
		return nil
	}
	p.handSink()
}

//Unmarshal unmarshal to the type slice
func (p *pipeline) Unmarshal(slice interface{}) {
	dest := reflect.Indirect(reflect.ValueOf(slice))
	if dest.Kind() != reflect.Slice {
		panic(errors.New("Unmarshal must accept a slice type "))
	}
	dest.Set(reflect.MakeSlice(dest.Type(), 0, 0))
	p.ForEach(func(t interface{}) {
		elem := reflect.ValueOf(t)
		dest.Set(reflect.Append(dest, elem))
	})
}

//handSink Do execute the pipeline chain
func (p *pipeline) handSink() {
	var out chan interface{}
	out = make(chan interface{})
	go func() {
		p.sourceStage.sourceList.ForEach(func(t interface{}) {
			out <- t
		})
		close(out)
	}()

	var outLink chan interface{}
	sourceStage := p.sourceStage
	for current := sourceStage.nextStage; current != nil; current = current.nextStage {
		if current == sourceStage.nextStage {
			outLink = current.sink(out)
			continue
		}
		outLink = current.sink(outLink)
	}
}
