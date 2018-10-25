package main

import (
	"dcrawl"
	"dspider"
	"sync"
)

/**
 * 主函数实现保存操作，其它地方都是解析
 * 一个爬虫，一个线程
 */

var MI = dcrawl.SMongoInfo{"127.0.0.1", 27017, "", "", "novel_online", "online"}

var SC = map[*dcrawl.SpiderContent]dcrawl.SpiderRun{}
var saveToMongo = make(chan dcrawl.NovelField, 10)
var MWAIT = sync.WaitGroup{}

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
	booktxt := dcrawl.SpiderContent{}
	booktxt.SpiderName = "booktxt"
	booktxt.BaseUrl = "https://www.booktxt.net/"
	booktxt.ToMongo = &saveToMongo
	booktxt.Exit = &MWAIT
	booktxt.SeedUrl = map[string]int{
		"https://www.booktxt.net/xiaoshuodaquan/": 0,
	}

	/* 添加爬虫 */
	//SC[&mzhu8] = dspider.Mzhu8Run
	SC[&booktxt] = dspider.BookTxtRun

	/* 开始运行爬虫 */
	for spp, spf := range SC {
		MWAIT.Add(1)
		go spf(spp)
	}

	/* 保存到 mongodb */
	go func() {
		MWAIT.Add(1)
		for novelField := range saveToMongo {
			dcrawl.SaveToMongo(MI, novelField)
		}
		MWAIT.Done()
	}()

	MWAIT.Wait()
}
