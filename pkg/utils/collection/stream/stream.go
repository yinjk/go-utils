/**
 * 流式处理接口
 * @author yinjk
 * @create 2019-05-13 18:01
 */
package list

func Map[I, O comparable](in Stream[I], m func(i I) O) Stream[O] {
	outCh := make(chan O)
	out := newLimitedPipelineFromCh(outCh)
	//in.Peek(func(t I) {
	//	o := m(t)
	//	outCh <- o
	//})
	go func() {
		in.ForEach(func(v I) {
			o := m(v)
			outCh <- o
		})
		close(outCh)
	}()

	return out
}

type Stream[T comparable] interface {
	isBatch() bool

	Distinct() Stream[T]

	Filter(test func(v T) bool) Stream[T]

	Skip(n int) Stream[T]

	Limit(maxSize int) Stream[T]

	Sorted(lessFun func(o1, o2 T) bool) Stream[T]

	Peek(accept func(t T)) Stream[T]

	//terminal operation

	Count() int

	Min(lessFun func(o1, o2 T) bool) T

	Max(lessFun func(o1, o2 T) bool) T

	FindFirst() T

	FindLast() T

	AnyMatch(test func(t T) bool) bool

	AllMatch(test func(t T) bool) bool

	NoneMatch(test func(t T) bool) bool

	ForEach(consumer func(t T))

	ToArray() []T
}
