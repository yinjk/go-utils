/**
 *
 * @author yinjk
 * @create 2019-04-19 15:00
 */
package times

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestStopWatch_GetTimeMillis(t *testing.T) {
	fmt.Println("一个非常耗时的程序执行开始...")
	stopWatch := NewStopWatch(true) //创建计时器，默认传true,如果传false，秒表会处于未激活状态，所有的方法都会直接返回
	//开始计时
	stopWatch.Start()
	time.Sleep(1 * time.Second) //睡眠一秒
	stopWatch.PrettyPrint("循环开始之前")
	for i := 0; i < 10000000000; i++ { //写一个超大的循环来模拟程序耗时
		if i == 1000000000 {
			stopWatch.PrettyPrint("循环执行了1/10")
		}
		if i == 5000000000 {
			stopWatch.PrettyPrint("循环执行了一半")
		}
		if i == 9000000000 {
			stopWatch.PrettyPrint("循环执行了9/10")
		}
	}
	stopWatch.PrettyPrint("循环结束")
	stopWatch.PrettyPrint("打印次数1")
	stopWatch.PrettyPrint("打印次数2")
	stopWatch.PrettyPrint("打印次数3")
	stopWatch.PrettyPrint("打印次数4")
	stopWatch.PrettyPrint("打印次数5，第5次打印的时间和第一次相同，所以对性能的损耗几乎可忽略") //可以看到连续打印五次耗时不到1ms，
	stopWatch.Stop()                                            //停止计时
	time.Sleep(1 * time.Second)                                 //停止计时后睡眠一秒
	stopWatch.PrettyPrint("停止秒表，睡眠一秒之后时间也不会再流动")
	stopWatch.Start() //重新开始计时
	stopWatch.PrettyPrint("重新计时之后")
	time.Sleep(1 * time.Second) //重新计时后睡眠一秒
	stopWatch.PrettyPrint("睡眠1秒")
	//觉得默认的打印格式不喜欢?通过SetPrintFormat函数自定义打印格式
	stopWatch.SetPrintFormat(func(sw StopWatch, message string) { //这里我们不打印ms单位，打印s单位，并且将消息上下的 ---- 替换成 ****
		fmt.Println("*****************************************")
		fmt.Println(strconv.Itoa(int(sw.GetTimeSeconds())) + "s   %   " + message)
		fmt.Println("*****************************************")
	})
	stopWatch.PrettyPrint("自定义的新打印格式")
	stopWatch.PrettyPrint("再打印一次")
	//想在每次打印的时候自动计算与上一次打印使的时间差?stopWatch提供了GetDurationPrintFormat可以打印差值
	stopWatch.SetPrintFormat(stopWatch.GetDurationPrintFormat()) //设置：以 自动计算与上一次打印的时间差 的方式打印
	time.Sleep(100 * time.Millisecond)                           //睡眠一秒
	stopWatch.PrettyPrint("设置差值打印之后，睡眠100毫秒")
	time.Sleep(200 * time.Millisecond) //睡眠两秒
	stopWatch.PrettyPrint("设置差值打印之后，睡眠200毫秒")
	//设置回默认打印格式
	stopWatch.SetPrintFormat(stopWatch.GetDefaultPrintFormat())
	stopWatch.PrettyPrint("重新设置回默认打印方式")
	fmt.Println("程序执行结束！")
}
