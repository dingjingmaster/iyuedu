package dspider

import (
	"dcrawl"
	"fmt"
	. "library/goquery"
	"norm"
	"strconv"
	"strings"
)

// 获取书籍 url
func mzhu8GetUrl(baseUrl string, seedUrl *map[string]int, bookUrl *map[string]bool) {

	mulu := []string{}

	for k, v := range *seedUrl {
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

		doc, err:= NewDocumentFromReader(strings.NewReader(html))
		if nil != err {
			dcrawl.Log.Errorf("解析html失败:%s",v)
		}

		doc.Find("a").Each(func(i int, selection *Selection) {
			url, ret := selection.Attr("href")
			if ret && strings.HasPrefix(url, "http://"){
				// 过滤掉目录url + baseUrl
				if strings.HasPrefix(url, baseUrl + "/mulu/") || (url == baseUrl) || (url == baseUrl + "/") {return}
				(*bookUrl)[url] = false
			}
		})
	}
}

// 解析书籍 url
func mzhu8ParseBook(baseUrl string, bookUrl *map[string]bool, data chan dcrawl.NovelField) {
	if ("" == baseUrl || nil == bookUrl) || (len(*bookUrl) <= 0) {
		dcrawl.Log.Errorf("base url 或 book url 错误")
		return
	}

	for url, _ := range *bookUrl {
		novelInfo := dcrawl.NovelField{}
		novelInfo.ChapterUrl = map[string]string{}
		ret, html :=dcrawl.GetHTMLByUrl(&url)
		if !ret {
			dcrawl.Log.Errorf("获取html失败:%s", url)
			continue
		}

		doc, err:= NewDocumentFromReader(strings.NewReader(html))
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

		novelInfo.NovelUrl = url
		novelInfo.ImgUrl, _ = bookInfo.Find("#fmimg").Find("img").Attr("src")
		name := bookInfo.Find("h1").Text()
		novelInfo.Name = norm.NormName(name)
		author := bookInfo.Find(".infos").Find(".i_author").Text()
		novelInfo.Author = norm.NormAuthor(author)
		category := bookInfo.Find(".infos").Find(".i_sort").Text()
		novelInfo.Tags = norm.NormCategory(category)
		status := bookInfo.Find(".infos").Find(".i_lz").Text()
		novelInfo.Status = norm.NormStatus(status)
		desc := bookInfo.Find("p").Text()
		novelInfo.Desc = norm.NormDesc(desc)

		/* 获取书籍章节信息 */
		chapterInfo := doc.Find(".box_con")
		if 0 == bookInfo.Size() {
			chapterInfo = chapterInfo.Find("#list")
			if 0 == bookInfo.Size() {
				dcrawl.Log.Errorf("未获取到章节信息:%s", url)
				continue
			}
		}

		section := ""
		chapterInfo.Find("dl").Children().Each(func(i int, selection *Selection) {
			if selection.Is("dt") {
				dt := selection.Text()
				if norm.CheckSection(dt) {
					section = strings.TrimSpace(dt)
				}
				return
			}
			if "" != selection.Text() {
				h, _ := selection.Find("a").Attr("href")
				novelInfo.ChapterUrl[h] = section + " " + norm.NormChapterName(selection.Text())
				//fmt.Println(section + " " + norm.NormChapterName(selection.Text()) + " --- " + h)
			}
		})

		data <- novelInfo
	}
	close(data)
}










func Mzhu8Run(np *dcrawl.SpiderContent) {
	dcrawl.Log.Infof("mzhu8开始执行,base url: %s", np.BaseUrl)

	bookUrl := map[string]bool{}
	novelData := make(chan dcrawl.NovelField, 100)


	/* 获取url */
	mzhu8GetUrl(np.BaseUrl, &np.SeedUrl, &bookUrl)

	/* 解析小说 */
	go mzhu8ParseBook(np.BaseUrl, &bookUrl, novelData)

	/* 获取小说基本信息 */
	for info := range novelData {
		fmt.Println(info.Name)
	}

	dcrawl.SpiderGroup.Done()
}








