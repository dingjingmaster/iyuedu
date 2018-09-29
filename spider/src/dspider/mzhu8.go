package dspider

import (
	"dcrawl"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"norm"
	"strconv"
	"strings"
)


var mzhu8 = dcrawl.SpiderContent{}

func init()  {
	// 名著8
	mzhu8.BaseUrl = "http://www.mzhu8.com"
	mzhu8.BookUrl = map[string]bool{}
	mzhu8.SeedUrl = map[string]int{
		"http://www.mzhu8.com/mulu/1/":8,
		//"http://www.mzhu8.com/mulu/2/":8,
		//"http://www.mzhu8.com/mulu/3/":3,
		//"http://www.mzhu8.com/mulu/5/":5,
		//"http://www.mzhu8.com/mulu/7/":13,
		//"http://www.mzhu8.com/mulu/6/":51,
		//"http://www.mzhu8.com/mulu/16/":36,
	}
}

// 获取书籍 url
func Mzhu8GetUrl() bool {

	mulu := []string{}

	for k, v := range mzhu8.SeedUrl {
		for i := 1; i <= v; i++ {
			mulu = append(mulu, k + strconv.Itoa(i) + ".html")
		}
	}

	// 获取书籍 bookUrl
	for _, v := range mulu {
		ret, html :=dcrawl.GetHTMLByUrl(&v)
		if !ret {
			dcrawl.Log.Errorf("获取html失败:%s", v)
			continue
		}

		doc, err:= goquery.NewDocumentFromReader(strings.NewReader(html))
		if nil != err {
			dcrawl.Log.Errorf("解析html失败:%s",v)
		}

		doc.Find("a").Each(func(i int, selection *goquery.Selection) {
			url, ret := selection.Attr("href")
			if ret && strings.HasPrefix(url, "http://"){
				// 过滤掉目录url + baseUrl
				if strings.HasPrefix(url, mzhu8.BaseUrl + "/mulu/") || (url == mzhu8.BaseUrl) || (url == mzhu8.BaseUrl + "/") {return}
				mzhu8.BookUrl[url] = false
			}
		})
	}

	return true
}

// 解析书籍 url
func Mzhu8ParseBook() bool {

	if (nil == mzhu8.BookUrl) || (len(mzhu8.BookUrl) <= 0) {
		return false
	}

	for url, _ := range mzhu8.BookUrl {
		ret, html :=dcrawl.GetHTMLByUrl(&url)
		if !ret {
			dcrawl.Log.Errorf("获取html失败:%s", url)
			continue
		}

		doc, err:= goquery.NewDocumentFromReader(strings.NewReader(html))
		if nil != err {
			dcrawl.Log.Errorf("解析html失败:%s",url)
			continue
		}

		/* 获取书籍信息 */
		bookInfo := doc.Find(".book_info")
		if 0 == bookInfo.Size() {
			dcrawl.Log.Errorf("未获取到书籍信息:%s", url)
			continue
		}

		//imgUrl, _ := bookInfo.Find("#fmimg").Find("img").Attr("src")
		name := bookInfo.Find("h1").Text()
		name = norm.NormName(name)
		author := bookInfo.Find(".infos").Find(".i_author").Text()
		author = norm.NormAuthor(author)
		category := bookInfo.Find(".infos").Find(".i_sort").Text()
		category = norm.NormCategory(category)
		status := bookInfo.Find(".infos").Find(".i_lz").Text()
		status = norm.NormStatus(status)
		desc := bookInfo.Find("p").Text()
		desc = norm.NormDesc(desc)

		/* 获取书籍章节信息 */
		chapterInfo := doc.Find(".box_con")
		if 0 == bookInfo.Size() {
			chapterInfo = chapterInfo.Find("#list")
			if 0 == bookInfo.Size() {
				dcrawl.Log.Errorf("未获取到章节信息:%s", url)
				continue
			}
		}

		chapterInfo.Find("dl").Children().Each(func(i int, selection *goquery.Selection) {
			//dt := ""
			if selection.Is("dt") {
				//dt = selection.Text()
				//fmt.Println(dt)
				return
			}
			fmt.Println(selection.Text() + "------------" +  norm.NormChapterName(selection.Text()))
		})

		return true
	}

	return true
}
