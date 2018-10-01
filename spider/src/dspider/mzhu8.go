package dspider

import (
	"dcrawl"
	"fmt"
	"io/ioutil"
	. "library/goquery"
	"net/http"
	"norm"
	"strconv"
	"strings"
	"sync"
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
			}
		})

		data <- novelInfo
	}
	close(data)
}

/* 获取内容、下载图片 和 章节 */
func downloadData(baseUrl string, dn sync.WaitGroup, nd chan dcrawl.NovelField, toMongo chan dcrawl.NovelBean) {
	for info := range nd {

		// 章节缓冲
		//chapterContent := map[string]string{}

		novelBean := dcrawl.NovelBean{}


		// id 要重新做 n + a + yuan

		novelBean.Info.Name = info.Name
		novelBean.Info.Author = info.Author
		novelBean.Info.NovelUrl = info.NovelUrl
		novelBean.Info.NovelParse = "mzhu8"

		novelBean.Info.Tags = info.Tags
		novelBean.Info.Status = info.Status
		novelBean.Info.Desc = info.Desc

		novelBean.Info.ChapterUrl = info.ChapterUrl


		/* 下载图片 */
		img := info.ImgUrl
		it := strings.Split(img, ".")
		imgType := it[len(it) -1 ]

		resp, err := http.Get(img)
		if nil != err {
			dcrawl.Log.Errorf("下载图片: %s 失败", img)
		}
		defer resp.Body.Close()

		imgContent, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			dcrawl.Log.Errorf("读取图片: %s 失败", img)
		}
		novelBean.Info.ImgUrl = img
		novelBean.Info.ImgType = imgType
		novelBean.Info.ImgContent = imgContent

		/* 下载章节 */
		for url, cname := range info.ChapterUrl {
			url = baseUrl + url

			client := &http.Client{}
			request, _ := http.NewRequest("Get", url, nil)
			request.Header.Add("Set-Cookie", "UM_distinctid=1660e90cb9f7c-05bcac472ed42a-8383268-1fa400-1660e90cba05fe; PHPSESSID=85cdb747bc14135ba65b2983576e1ae4; jieqiUserInfo=jieqiUserId%3D5723%2CjieqiUserName%3Ddingjing%2CjieqiUserGroup%3D3%2CjieqiUserName_un%3Ddingjing%2CjieqiUserLogin%3D1538402456%2CjieqiUserVip%3D0%2CjieqiUserPassword%3D25d55ad283aa400af464c76d713c07ad%2CjieqiUserHonor_un%3D%26%23x79C0%3B%26%23x624D%3B%2CjieqiUserGroupName_un%3D%26%23x666E%3B%26%23x901A%3B%26%23x4F1A%3B%26%23x5458%3B; jieqiVisitInfo=jieqiUserLogin%3D1538402456%2CjieqiUserId%3D5723; CNZZDATA4695146=cnzz_eid%3D764393492-1537838282-%26ntime%3D1538409214")
			request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
			request.Header.Add("Referer", info.NovelUrl)

			resp, _ := client.Do(request)
			hc, _ := ioutil.ReadAll(resp.Body)

			//ret, html := dcrawl.GetHTMLByUrl(&url)
			//if !ret {
			//	dcrawl.Log.Errorf("章节(%s): %s 获取失败", cname, url)
			//}

			fmt.Println(cname)
			fmt.Println(dcrawl.ConvertToString(string(hc), "gbk", "utf8"))
		}
		return
		fmt.Println(imgContent)
		fmt.Println(img)
		fmt.Println(imgType)


	}
	dn.Done()
}


/* 存储 这里会进行检查，相对复杂 */
func saveToMongo()  {
	
}



func Mzhu8Run(np *dcrawl.SpiderContent) {
	dcrawl.Log.Infof("mzhu8开始执行,base url: %s", np.BaseUrl)

	downloadNum := 10
	bookUrl := map[string]bool{}
	novelData := make(chan dcrawl.NovelField, 1000)
	downloadGroup := sync.WaitGroup{}
	saveData := make(chan dcrawl.NovelBean, 100)


	/* 获取url */
	mzhu8GetUrl(np.BaseUrl, &np.SeedUrl, &bookUrl)

	/* 解析小说 */
	go mzhu8ParseBook(np.BaseUrl, &bookUrl, novelData)

	/* 获取小说基本信息 */
	for i := 0; i < downloadNum; i ++ {
		downloadGroup.Add(1)
		go downloadData(np.BaseUrl, downloadGroup, novelData, saveData)
	}
	downloadGroup.Wait()

	/* 保存小说 */

	dcrawl.SpiderGroup.Done()
}








