package main

import (
	"dcrawl"
	"fmt"
	"sort"
)

/**
 * 主函数实现保存操作，其它地方都是解析
 *
 *
 */

var MI = dcrawl.SMongoInfo {"127.0.0.1", 27017, "", "", "novel_online", "online"}

var SC = map[*dcrawl.SpiderContent]dcrawl.SpiderRun{}

func main() {

	///* mzhu8 爬虫 */
	//mzhu8 := dcrawl.SpiderContent{}
	//mzhu8.SpiderName = "mzhu8"
	//mzhu8.BaseUrl = "http://www.mzhu8.com"
	//mzhu8.MI = MI
	//mzhu8.SeedUrl = map[string]int{
	//	"http://www.mzhu8.com/mulu/1/":8,
	//	//"http://www.mzhu8.com/mulu/2/":8,
	//	//"http://www.mzhu8.com/mulu/3/":3,
	//	//"http://www.mzhu8.com/mulu/5/":5,
	//	//"http://www.mzhu8.com/mulu/7/":13,
	//	//"http://www.mzhu8.com/mulu/6/":51,
	//	//"http://www.mzhu8.com/mulu/16/":36,
	//}
	//
	///* 添加爬虫 */
	//SC[&mzhu8] = dspider.Mzhu8Run
	//
	///* 开始运行爬虫 */
	//dcrawl.SpiderGroup.Add(len(SC))
	//for spp, spf := range SC {
	//	go spf(spp)
	//}
	//
	///* 等待所有爬虫执行完毕 */
	//dcrawl.SpiderGroup.Wait()
	//

	m := make(map[int]string)
	m[1] = "a"
	m[2] = "c"
	m[0] = "b"
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Println("Key:", k, "Value:", m[k])
	}


	//novel := dcrawl.NovelBean{}
	//ninfo := dcrawl.NovelInfo{}
	//ndata := dcrawl.NovelData{}
	//novel.Info = ninfo
	//novel.Data = append(novel.Data, ndata)
	//mi := dcrawl.SMongoInfo{"127.0.0.1", 27017, "", "", "test_mongo", "testCollection"}
	//
	////fmt.Println(dcrawl.InserDoc(mi, &novel))
	//
	//fmt.Println(dcrawl.FindDocById(mi, "ffffffffffff", &novel))
	//fmt.Println(novel.Info.Name)
}
