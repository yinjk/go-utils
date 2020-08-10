/*
 @Desc

 @Date 2020-06-23 20:28
 @Author yinjk
*/
package syncs

import "sync"

type TaskExecutor struct {
	sync.Mutex
	goroutine int            //协程池总协程数
	waitLen   int            //最大等待队列数
	reject    func(f func()) //超过最大等待的拒绝策略
	taskQueue chan func()    //任务排队队列

	stopChan chan struct{} //结束协程池，当前正在执行的任务会执行完，等待队列中的任务不会被执行
	stop     bool
}

type ExecutorBuilder struct {
	goroutine int            //协程池总协程数
	waitLen   int            //最大等待队列数
	reject    func(f func()) //超过最大等待的拒绝策略
}

func (b *ExecutorBuilder) Goroutine(r int) *ExecutorBuilder {
	b.goroutine = r
	return b
}

func (b *ExecutorBuilder) WaitLen(wl int) *ExecutorBuilder {
	b.waitLen = wl
	return b
}
func (b *ExecutorBuilder) Reject(reject func(f func())) *ExecutorBuilder {
	b.reject = reject
	return b
}

func (b *ExecutorBuilder) Build() *TaskExecutor {
	if b.goroutine == 0 {
		b.goroutine = 10
	}
	if b.waitLen == 0 {
		b.waitLen = 10
	}
	return NewTaskExecutor(b.goroutine, b.waitLen, b.reject)
}

/*
	g       :
	waitQueue :

*/
func NewTaskExecutor(goroutine, waitLen int, reject func(f func())) *TaskExecutor {
	t := &TaskExecutor{
		goroutine: goroutine,
		waitLen:   waitLen,
		reject:    reject,
		taskQueue: make(chan func(), waitLen),
		stopChan:  make(chan struct{}, goroutine),
		stop:      false,
	}
	if t.reject == nil {
		t.reject = t.defaultReject
	}
	t.start()
	return t
}

func (t TaskExecutor) start() {
	for i := 0; i < t.goroutine; i++ {
		go func() {
			for {
				select {
				case task := <-t.taskQueue:
					task()
				case <-t.stopChan:
					return
				}
			}
		}()
	}
}

func (t TaskExecutor) Shutdown() {
	t.Lock()
	defer t.Unlock()
	if t.stop {
		return
	}
	for i := 0; i < t.goroutine; i++ {
		t.stopChan <- struct{}{}
	}
	t.stop = true
}

func (t TaskExecutor) Execute(f func()) {
	queueLen := len(t.taskQueue)
	if queueLen >= t.waitLen {
		t.reject(f)
		return
	}
	t.taskQueue <- f
}

// 默认拒绝策略：将被拒绝的task重新加入task队列中，直到加入成功
func (t TaskExecutor) defaultReject(f func()) {
	t.taskQueue <- f
}
