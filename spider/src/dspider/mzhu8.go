package dspider

import (
	"crypto/md5"
	"dcrawl"
	"encoding/hex"
	"io/ioutil"
	. "library/goquery"
	"net/http"
	"norm"
	"strconv"
	"strings"
	"sync"
	"time"
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
		ret, html :=dcrawl.GetHTMLByUrl(&v, nil)
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
		novelInfo.ImgContent = []byte{}
		novelInfo.ErrorChapterUrl = map[string]string{}
		novelInfo.ChapterContent = map[string]string{}
		ret, html :=dcrawl.GetHTMLByUrl(&url, nil)
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
		dcrawl.Log.Infof("解析 %s|%s 信息成功!!!", novelInfo.Name, novelInfo.Author)
	}
	close(data)
}


/* 获取内容、下载图片 和 章节 */
func downloadData(baseUrl string, spiderName string, info dcrawl.NovelField, /*nd chan dcrawl.NovelField,*/ download sync.WaitGroup, toMongo chan dcrawl.NovelField) {
	//for info := range nd {
		info.NovelParse = spiderName
		dcrawl.Log.Infof("开始下载 %s|%s 内容", info.Name, info.Author)

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
		info.ImgType = imgType
		info.ImgContent = imgContent

		/* 下载章节 */
		for url, cname := range info.ChapterUrl {
			url = baseUrl + url
			post := "http://www.mzhu8.com/modules/article/show.php"
			head1 := map[string]string{}

			head1["Set-Cookie"] = "UM_distinctid=1660e90cb9f7c-05bcac472ed42a-8383268-1fa400-1660e90cba05fe; PHPSESSID=85cdb747bc14135ba65b2983576e1ae4; jieqiUserInfo=jieqiUserId%3D5723%2CjieqiUserName%3Ddingjing%2CjieqiUserGroup%3D3%2CjieqiUserName_un%3Ddingjing%2CjieqiUserLogin%3D1538402456%2CjieqiUserVip%3D0%2CjieqiUserPassword%3D25d55ad283aa400af464c76d713c07ad%2CjieqiUserHonor_un%3D%26%23x79C0%3B%26%23x624D%3B%2CjieqiUserGroupName_un%3D%26%23x666E%3B%26%23x901A%3B%26%23x4F1A%3B%26%23x5458%3B; jieqiVisitInfo=jieqiUserLogin%3D1538402456%2CjieqiUserId%3D5723; CNZZDATA4695146=cnzz_eid%3D764393492-1537838282-%26ntime%3D1538409214"
			head1["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"
			head1["Referer"] = info.NovelUrl

			ret, body := dcrawl.GetHTMLByUrl(&url, &head1)
			if !ret {
				info.ErrorChapterUrl[url] = cname
				dcrawl.Log.Errorf("请求错误: %s", url)
				continue
			}

			doc, err:= NewDocumentFromReader(strings.NewReader(body))
			if nil != err {
				info.ErrorChapterUrl[url] = cname
				dcrawl.Log.Errorf("解析html失败: %s",url)
				continue
			}

			/* 解析post参数 */
			script := doc.Find("#chapterContent>script").Text()
			arr := strings.Split(script, "\"")
			if len(arr) < 5 {
				info.ErrorChapterUrl[url] = cname
				dcrawl.Log.Errorf("错误的script:%s", url)
				continue
			}

			/* 获取文章内容 */
			head2 := map[string]string{}
			para := map[string]string{}
			head2["Cookie"] = "UM_distinctid=1660e90cb9f7c-05bcac472ed42a-8383268-1fa400-1660e90cba05fe; PHPSESSID=85cdb747bc14135ba65b2983576e1ae4; CNZZDATA4695146=cnzz_eid%3D764393492-1537838282-%26ntime%3D1538526303; jieqiUserInfo=jieqiUserId%3D5723%2CjieqiUserName%3Ddingjing%2CjieqiUserGroup%3D3%2CjieqiUserName_un%3Ddingjing%2CjieqiUserLogin%3D1538528772%2CjieqiUserVip%3D0%2CjieqiUserPassword%3D25d55ad283aa400af464c76d713c07ad%2CjieqiUserHonor_un%3D%26%23x79C0%3B%26%23x624D%3B%2CjieqiUserGroupName_un%3D%26%23x666E%3B%26%23x901A%3B%26%23x4F1A%3B%26%23x5458%3B; jieqiVisitInfo=jieqiUserLogin%3D1538528772%2CjieqiUserId%3D5723"
			head2["Content-Type"] = "application/x-www-form-urlencoded"
			head2["Cache-Control"] = "no-cache"
			head2["Referer"] = url
			head2["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"

			para["aid"] = arr[1]
			para["cid"] = arr[3]
			para["r"] = "0.4814884913447719"

			ret, body = dcrawl.Post(&post, &head2, &para)
			if !ret {
				info.ErrorChapterUrl[url] = cname
				dcrawl.Log.Errorf("错误的响应: %s", url)
				continue
			}

			info.ChapterContent[para["cid"] + "{]" + cname] = norm.NormContent(body)
		}
		toMongo <- info
		dcrawl.Log.Infof("下载 %s|%s 成功!!!", info.Name, info.Author)
		download.Done()
	//}
}

/* 存储 这里会进行检查，相对复杂 */
func saveToMongo(mongo dcrawl.SMongoInfo, spiderName string, dn sync.WaitGroup, data chan dcrawl.NovelField)  {
	for info := range data {
		dcrawl.Log.Infof("开始保存书籍: %s|%s|%s !!!", spiderName, info.Name, info.Author)
		tmd5 := md5.New()
		tmd5.Write([]byte(info.Name + info.Author + spiderName))
		id :=  hex.EncodeToString(tmd5.Sum(nil))
		times := time.Now().Format("20060102150405")

		// 是否已有该书籍
		novelTmp := dcrawl.NovelBean{}
		ret := dcrawl.FindDocById(mongo, id, &novelTmp)
		if ret {
			// 已有这一条信息
			if novelTmp.Info.MaskLevel > 0 {
				dcrawl.Log.Infof("不需要更新的书籍: %s|%s", novelTmp.Info.Name, novelTmp.Info.Author)
				continue
			}
			/* 更新主要字段 */
			// 图片
			if "" == novelTmp.Info.ImgType {
				novelTmp.Info.ImgType = info.ImgType
				novelTmp.Info.ImgContent = info.ImgContent
				novelTmp.Info.ImgUrl = info.ImgUrl
			}

			if "" != info.Tags {novelTmp.Info.Tags = info.Tags}								// 类别
			if "" != info.Status {novelTmp.Info.Status = info.Status}						// 状态
			if "" != info.Desc {novelTmp.Info.Desc = info.Desc}								// 简介
			info.Desc = times

			/* 检查章节数 */
			chapterAll := map[string]string{}
			for _, data := range novelTmp.Data {
				for k, v := range data.ChapterContent {
					chapterAll[k] = v
				}
			}
			if len(info.ChapterContent) > len(chapterAll) {
				// 更新章节内容
				for ik, iv := range info.ChapterContent {
					if _, exist := chapterAll[ik]; !exist {
						chapterAll[ik] = iv
					}
				}
				// 更新章节ChapterUrl
				for ik, iv := range info.ChapterUrl {
					if _, exist := novelTmp.Info.ChapterUrl[ik]; !exist {
						novelTmp.Info.ChapterUrl[ik] = iv
					}
				}
				// 更新错误章节信息
				for ik, iv := range info.ErrorChapterUrl {
					novelTmp.Info.ErrorChapter[ik] = iv
				}
				for ik, iv := range novelTmp.Info.ErrorChapter {
					arr := strings.Split(ik, "/")
					ii := strings.Split(arr[len(arr) - 1], ".")[0]
					if _, exits := chapterAll[ii + "{]" + iv]; exits {
						delete(novelTmp.Info.ErrorChapter, ii + "{]" + iv)
					}
				}
			}

			novelTmp.Info.ChapterNum = len(info.ChapterUrl)
			novelTmp.Info.BlockVolume = dcrawl.NovelContentNum

			/* 重新封装章节内容 */
			blockIds := []string{}
			data := [] dcrawl.NovelData{}
			dcrawl.GeneratorChapterContent(info.Name, info.Author, spiderName, chapterAll, &blockIds, &data)
			novelTmp.Info.Blocks = blockIds
			novelTmp.Data = data
			if dcrawl.UpdateDoc(mongo, id, &novelTmp) {
				dcrawl.Log.Infof("更新 %s|%s|%s 成功!!!", spiderName, info.Name, info.Author)
			}
		} else {
			novelTmp.Info.Id = id
			novelTmp.Info.Name = info.Name
			novelTmp.Info.Author = info.Author
			novelTmp.Info.NovelUrl = info.NovelUrl
			novelTmp.Info.NovelParse = spiderName

			novelTmp.Info.ImgUrl = info.ImgUrl
			novelTmp.Info.ImgContent = info.ImgContent
			novelTmp.Info.ImgType = info.ImgType

			novelTmp.Info.Tags = info.Tags
			novelTmp.Info.Status = info.Status
			novelTmp.Info.Desc = info.Desc

			novelTmp.Info.ChapterUrl = info.ChapterUrl
			novelTmp.Info.ErrorChapter = info.ErrorChapterUrl

			novelTmp.Info.Intime = times
			novelTmp.Info.Uptime = times

			novelTmp.Info.MaskLevel = 0
			novelTmp.Info.BlockVolume = dcrawl.NovelContentNum
			novelTmp.Info.ChapterNum = len(info.ChapterUrl)

			blockIds := []string{}
			data := [] dcrawl.NovelData{}
			dcrawl.GeneratorChapterContent(info.Name, info.Author, spiderName, info.ChapterContent, &blockIds, &data)
			novelTmp.Info.Blocks = blockIds
			novelTmp.Data = data

			if dcrawl.InserDoc(mongo, &novelTmp) {
				dcrawl.Log.Infof("插入 %s|%s|%s 成功!!!", spiderName, info.Name, info.Author)
			}
		}
	}

	dn.Done()
}


func Mzhu8Run(np *dcrawl.SpiderContent) {
	//downloadNum := 10
	bookUrl := map[string]bool{}
	saveGroup := sync.WaitGroup{}
	downloadGroup := sync.WaitGroup{}
	novelData := make(chan dcrawl.NovelField, 10)
	saveData := make(chan dcrawl.NovelField, 10)

	dcrawl.Log.Infof("mzhu8开始执行,base url: %s", np.BaseUrl)

	/* 获取url */
	mzhu8GetUrl(np.BaseUrl, &np.SeedUrl, &bookUrl)
	dcrawl.Log.Infof("mzhu8获取url成功")

	/* 解析小说 */
	go mzhu8ParseBook(np.BaseUrl, &bookUrl, novelData)

	/* 获取小说基本信息 */
	for info := range novelData {
		downloadGroup.Add(1)
		go downloadData(np.BaseUrl, np.SpiderName, info, downloadGroup, saveData)
	}

	/* 保存小说 */
	saveGroup.Add(1)
	go saveToMongo(np.MI, np.SpiderName, saveGroup, saveData)

	downloadGroup.Wait(); close(saveData)
	saveGroup.Wait(); dcrawl.SpiderGroup.Done()
	dcrawl.Log.Infof("mzhu8 爬虫执行完毕!!!")
}
