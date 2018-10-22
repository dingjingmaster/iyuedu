package main

import (
	"dcrawl"
	"fmt"
)

/**
 * 主函数实现保存操作，其它地方都是解析
 * 一个爬虫，一个线程
 */

//var MI = dcrawl.SMongoInfo{"127.0.0.1", 27017, "", "", "novel_online", "online"}

//var SC = map[*dcrawl.SpiderContent]dcrawl.SpiderRun{}

func main() {
	/* mzhu8 爬虫 */
	//mzhu8 := dcrawl.SpiderContent{}
	//mzhu8.SpiderName = "mzhu8"
	//mzhu8.BaseUrl = "http://www.mzhu8.com"
	//mzhu8.MI = MI
	//mzhu8.SeedUrl = map[string]int{
	//	//"http://www.mzhu8.com/mulu/18/":1,
	//	"http://www.mzhu8.com/mulu/1/":  8,
	//	"http://www.mzhu8.com/mulu/2/":  8,
	//	"http://www.mzhu8.com/mulu/3/":  3,
	//	"http://www.mzhu8.com/mulu/5/":  5,
	//	"http://www.mzhu8.com/mulu/7/":  13,
	//	"http://www.mzhu8.com/mulu/6/":  51,
	//	"http://www.mzhu8.com/mulu/16/": 36,
	//}

	/* 顶点小说 */
	//booktxt := dcrawl.SpiderContent{}
	//booktxt.SpiderName = "booktxt"
	//booktxt.BaseUrl = "http://www.mzhu8.com"
	//booktxt.MI = MI
	//booktxt.SeedUrl = map[string]int{
	//}

	/* 添加爬虫 */
	//SC[&mzhu8] = dspider.Mzhu8Run
	//SC[&booktxt] = dspider.Mzhu8Run

	/* 开始运行爬虫 */
	//for spp, spf := range SC {
	//	dcrawl.SpiderGroup.Add(1)
	//	go spf(spp)
	//}

	/* 等待所有爬虫执行完毕 */
	//dcrawl.SpiderGroup.Wait()



	ret, html := dcrawl.GetHTMLByUrl("http://www.baiddu.com", nil)
	fmt.Println(ret)
	fmt.Println(html)


}
