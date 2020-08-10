/**
 *
 * @author yinjk
 * @create 2019-04-19 14:40
 */
package times

import (
	"fmt"
	"strconv"
	"time"
)

type StopWatch struct {
	startTime       time.Time
	lastTime        time.Time
	stopped         bool
	currentTaskName string
	printMethod     func(sw StopWatch, message string)
	lastPrintTime   time.Time //最后一次打印的时间
	active          bool
}

/**
 * 创建一个秒表，并且自动开始计时
 * @param : active:是否激活秒表，默认传true，当我们解决了性能问题，不想再打印时间日志时，将这里改为false即可（不用再手动去删除每个打印方法）
 * @return:
 * @author: yinjk
 * @time  : 2019/4/19 16:59
 */
func NewStopWatch(active bool) *StopWatch {
	watch := &StopWatch{stopped: false, active: active}
	watch.Start()
	return watch
}

/**
 * 开始计时，如果之前已经开始，调用此方法会重新开始计时
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:28
 */
func (sw *StopWatch) Start() {
	if !sw.active {
		return
	}
	sw.startTime = time.Now()
	sw.stopped = false
	sw.lastPrintTime = sw.startTime
}

func (sw *StopWatch) Pause() {
	//doing nothing
}

func (sw *StopWatch) Continue() {
	//doing nothing
}

/**
 * 停止计时，停止后获取时间将不会再变化
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:29
 */
func (sw *StopWatch) Stop() {
	if !sw.active {
		return
	}
	sw.lastTime = time.Now()
	sw.stopped = true
}

/**
 * 重新开始计时
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:29
 */
func (sw *StopWatch) Reset() {
	sw.Stop()
	sw.Start()
}

/**
 * 获取从开始到现在经过的时间秒数（保留3位小数）
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:30
 */
func (sw StopWatch) GetTimeSeconds() float64 {
	seconds := sw.getDuration().Seconds()
	seconds, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", seconds), 64)
	return seconds
}

/**
 * 同上，获取毫秒数
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:30
 */
func (sw StopWatch) GetTimeMillis() int64 {
	return sw.getDuration().Nanoseconds() / 1000000
}

/**
 * 同上，获取纳秒数
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:30
 */
func (sw StopWatch) GetTimeNanoseconds() int64 {
	return sw.getDuration().Nanoseconds()
}

/**
 * 获取字符串格式的时间
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:31
 */
func (sw StopWatch) GetTimeString() string {
	return sw.getDuration().String()
}

/**
 * 设置时间打印格式
 * @param : printFormat：这里传一个方法，设置之后，PrettyPrint方法会调用这里设置的方法进行打印
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:31
 */
func (sw *StopWatch) SetPrintFormat(printFormat func(sw StopWatch, message string)) {
	sw.printMethod = printFormat
}

/**
 * 打印时间，stopWatch提供了两种风格的时间打印方式：1.普通打印：每次打印与Start的时间差值，2：差值打印：每次打印与上一次调用该方法的时间差值
 * @param : message：需要输出的消息
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:33
 */
func (sw *StopWatch) PrettyPrint(message string) {
	if !sw.active {
		return
	}
	if sw.printMethod == nil {
		sw.GetDefaultPrintFormat()(*sw, message)
	} else {
		//这里不直接传指针是为了防止用户在printMethod中去操作我们的秒表，比如：将秒表停掉
		sw.printMethod(*sw, message)
	}
	sw.lastPrintTime = time.Now()
}

/**
 * 默认打印方式
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:35
 */
func (sw StopWatch) GetDefaultPrintFormat() func(sw StopWatch, message string) {
	return func(sw StopWatch, message string) {
		fmt.Println("---------------------------------------------")
		fmt.Println(strconv.Itoa(int(sw.GetTimeMillis())) + "ms  %  " + message)
		fmt.Println("---------------------------------------------")
	}
}

/**
 * 差值打印方式
 * @param :
 * @return:
 * @author: yinjk
 * @time  : 2019/4/20 15:35
 */
func (sw StopWatch) GetDurationPrintFormat() func(sw StopWatch, message string) {
	return func(sw StopWatch, message string) {
		duration := time.Now().Sub(sw.lastPrintTime).Nanoseconds() / 1000000 //计算与上一次打印之间的时间差
		fmt.Println("--------------------------------------------")
		fmt.Println("duration: " + strconv.Itoa(int(duration)) + "ms   %   " + message)
		fmt.Println("--------------------------------------------")
	}
}

func (sw StopWatch) LastTime() time.Time {
	return sw.lastPrintTime
}
func (sw StopWatch) getDuration() time.Duration {
	var duration time.Duration
	if sw.stopped {
		duration = sw.lastTime.Sub(sw.startTime)
	} else {
		duration = time.Now().Sub(sw.startTime)
	}
	return duration
}
