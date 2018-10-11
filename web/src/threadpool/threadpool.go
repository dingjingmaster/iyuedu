package threadpool

import "sync"

type ThreadpoolFunc func(interface{})

type CThreadpool struct {
	/* 互斥锁 */
	mutex      sync.Mutex                        	// 互斥锁
	threadNum  int                               	// 线程数
	threadFunc *ThreadpoolFunc						// 执行函数
	threadPara 	[]interface{} 						// 执行函数参数
}

var ct CThreadpool

/* 设置线程数 */
func SetThreeadNum(num int) {
	if num > 0 {
		ct.threadNum = num
	}
}

/* 添加 */

/* 开始运行线程池 */
func Run() {
	for i := 0; i < ct.threadNum; i++ {
		go (*ct.threadFunc)(ct.threadPara)
	}
}
