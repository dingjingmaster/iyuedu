package main

import "dspider"

func main() {



	dspider.Mzhu8GetUrl()
	dspider.Mzhu8ParseBook()

	//dspider.Mzhu8GetUrl(&mzhu8)
	//for _,v := range mzhu8.BookUrl{
	//	fmt.Println(v)
	//}
	//println(mzhu8.BookUrl)

	//dcrawl.SpiderRun(&mzhu8)




	//novel := ditem.NovelBean{}
	//ninfo := ditem.NovelInfo{}
	//ndata := ditem.NovelData{}
	//novel.Info = ninfo
	//novel.Data = append(novel.Data, ndata)
	//mi := ddb.SMongoInfo{"127.0.0.1", 27017, "", "", "test_mongo", "testCollection"}
	//fmt.Println(ddb.InserDoc(mi, &novel))
}
