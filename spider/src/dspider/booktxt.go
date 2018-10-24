package dspider

import (
	"dcrawl"
	"io/ioutil"
	"library/goquery"
	"library/mgo.v2/bson"
	"net/http"
	"norm"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func booktxtGetUrl(baseUrl string, seedUrl *map[string]int, bookUrl *map[string]bool) {

	for url, _ := range *seedUrl {
		head := map[string]string{
			"Referer": "https://www.booktxt.net/",
			"Cookie":  "__jsluid=adec22b5996125f9c7014a20d6a84d1c; Hm_lvt_3a0ea2f51f8d9b11a51868e48314bf4d=1538577486,1538577516; Hm_lpvt_3a0ea2f51f8d9b11a51868e48314bf4d=1540189295",
		}
		ok, html := dcrawl.GetHTMLByUrl(url, &head)
		if !ok {
			dcrawl.Log.Errorf("请求url:%s 错误!", url)
			continue
		}

		// 获取书籍 url
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if nil != err {
			dcrawl.Log.Errorf("解析html失败:%s", url)
			continue
		}

		doc.Find("a").Each(func(i int, selection *goquery.Selection) {
			url, ret := selection.Attr("href")
			if ret && strings.HasPrefix(url, baseUrl) && (strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")) {
				// 过滤
				if (url == baseUrl) || (url == (baseUrl + "/")) {
					return
				}
				(*bookUrl)[url] = false
			}
		})
	}
}

func booktxtParserInfo(baseUrl string, bookUrls *map[string]bool, novelInfoChan *chan dcrawl.NovelField) {
	for url, _ := range *bookUrls {
		novelInfo := dcrawl.NovelField{}
		novelInfo.ChapterUrl = map[string]string{}
		novelInfo.ImgContent = []byte{}
		novelInfo.ErrorChapterUrl = map[string]string{}
		novelInfo.ChapterContent = map[string]string{}
		/* 下载 url */
		head := map[string]string{
			"Referer": "https://www.booktxt.net/",
			"Cookie":  "__jsluid=adec22b5996125f9c7014a20d6a84d1c; Hm_lvt_3a0ea2f51f8d9b11a51868e48314bf4d=1538577486,1538577516; Hm_lpvt_3a0ea2f51f8d9b11a51868e48314bf4d=1540189295",
		}
		ok, html := dcrawl.GetHTMLByUrl(url, &head)
		if !ok {
			dcrawl.Log.Errorf("请求url:%s 错误!", url)
			time.Sleep(time.Second)
			continue
		}

		/* 解析 url */
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if nil != err {
			dcrawl.Log.Errorf("解析html失败:%s", url)
			continue
		}
		// 所有信息
		mainInfo := doc.Find("#maininfo")

		/* 小说 url */
		novelInfo.NovelUrl = url

		/* 解析器名字 */
		novelInfo.NovelParse = "booktxt"

		/* 书名 */
		name := mainInfo.Find("#info>h1").Text()
		novelInfo.Name = norm.NormName(name)
		novelInfo.Status = "连载"

		/* 作者名 */
		author := ""
		mainInfo.Find("#info>p").Each(func(i int, selection *goquery.Selection) {
			part := regexp.MustCompile(`^作\S+者`)
			if len(part.FindAllString(selection.Text(), -1)) > 0 {
				author = part.ReplaceAllString(selection.Text(), "")
			}
		})
		novelInfo.Author = norm.NormAuthor(author)

		/* 图片 url */
		dcrawl.Log.Debugf("开始获取 %s 图片url!", novelInfo.Name)
		imgUrl, _ := doc.Find("#fmimg>img").Attr("src")
		if !strings.HasPrefix(imgUrl, baseUrl) {
			imgUrl = baseUrl + imgUrl
		}
		novelInfo.ImgUrl = imgUrl
		tpt := strings.Split(imgUrl, ".")
		novelInfo.ImgType = tpt[len(tpt)-1]

		/* 简介 */
		desc := mainInfo.Find("#intro>p").Text()
		novelInfo.Desc = norm.NormDesc(desc)

		/* 类别 */
		category := ""
		doc.Find(".box_con>.con_top>a").Each(func(i int, selection *goquery.Selection) {
			if (!strings.Contains(selection.Text(), "顶点小说")) && (!strings.Contains(selection.Text(), name)) {
				category = selection.Text()
			}
		})
		novelInfo.Tags = norm.NormCategory(category)

		/* 章节 及 url */
		flg := false
		cp := 0
		dcrawl.Log.Debugf("开始获取 %s 章节信息!", novelInfo.Name)
		doc.Find(".box_con>#list>dl").Children().Each(func(i int, selection *goquery.Selection) {
			if strings.Contains(selection.Text(), "正文") {
				flg = true
			}

			if flg {
				if selection.Is("dd") {
					cp++
					chapter := norm.NormChapterName(selection.Text())
					if href, ok := selection.Find("a").Attr("href"); ok {
						novelInfo.ChapterUrl[novelInfo.NovelUrl + href] = strconv.Itoa(cp) + "{]" + chapter
						dcrawl.Log.Debugf("获取书籍 %s|%s 章节 %s  -->  %s...", novelInfo.Name, novelInfo.Author, chapter, novelInfo.NovelUrl + href)
					}
				}
			}
		})

		*novelInfoChan <- novelInfo
		dcrawl.Log.Infof("%s|%s基本信息提取完成!", novelInfo.Name, novelInfo.Author)
	}
}

/* 下载小说 */
func booktxtDownload(mongo dcrawl.SMongoInfo, wait *sync.WaitGroup, novelInfo *chan dcrawl.NovelField, toSave *chan dcrawl.NovelField) {
	if (nil == novelInfo) || (nil == toSave) {
		dcrawl.Log.Errorf("输入参数错误")
		return
	}

	for {
		if ninfo, ok := <-*novelInfo; ok {
			novelTemp := dcrawl.NovelBean{}
			// 找到数据库中的 书籍，看看章节信息相差大不大
			name := ninfo.Name
			author := ninfo.Author
			spider := ninfo.NovelParse

			field := bson.M{"name": name, "author": author, "novel_parse": spider}
			if dcrawl.FindDocByField(mongo, &field, &novelTemp) {
				// 比较章节数量是否相同
				if (novelTemp.Info.ChapterNum == len(ninfo.ChapterUrl)) && (len(novelTemp.Info.ErrorChapter) <= 0) {
					dcrawl.Log.Debugf("数据库中已存在:%s|%s", name, author)
					continue
				}
			}

			/* 没有找到 或者 有错误章节 或者 有新章节 ---> 开始下载 */
			// 下载图片
			dcrawl.Log.Debugf("开始下载 %s|%s 图片!", name, author)
			img := ninfo.ImgUrl
			resp, err := http.Get(img)
			if nil != err {
				dcrawl.Log.Errorf("下载图片: %s 失败", img)
			} else {
				imgContent, err := ioutil.ReadAll(resp.Body)
				if nil != err {
					dcrawl.Log.Errorf("读取图片: %s 失败", img)
				}
				defer resp.Body.Close()
				ninfo.ImgContent = imgContent
			}

			// 下载章节
			dcrawl.Log.Debugf("开始下载 %s|%s 章节!", name, author)
			for url, cname := range ninfo.ChapterUrl {
				head := map[string]string{
					"Referer": ninfo.NovelUrl,
					"Cookie":  "__jsluid=adec22b5996125f9c7014a20d6a84d1c; Hm_lvt_3a0ea2f51f8d9b11a51868e48314bf4d=1538577486,1538577516; Hm_lpvt_3a0ea2f51f8d9b11a51868e48314bf4d=1540189295",
				}
				ok, html := dcrawl.GetHTMLByUrl(url, &head)
				if !ok {
					dcrawl.Log.Errorf("请求url:%s 错误!", url)
					ninfo.ErrorChapterUrl[url] = cname
					time.Sleep(time.Second * 3)
					continue
				}

				/* 解析 url */
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
				if nil != err {
					dcrawl.Log.Errorf("解析html失败:%s", url)
					ninfo.ErrorChapterUrl[url] = cname
					continue
				}

				/* 获取文章内容 */
				content := doc.Find("#content").Text()
				ninfo.ChapterContent[cname] = norm.NormContent(content)
				time.Sleep(time.Millisecond * 3)
				dcrawl.Log.Debugf("书籍 %s|%s|%s 章节下载完成!", ninfo.Name, ninfo.Author, cname)
			}
			*toSave <- ninfo
			dcrawl.Log.Infof("书籍 %s|%s 信息获取完成!", ninfo.Name, ninfo.Author)
		} else {
			break
		}
	}
	wait.Done()
}

func BookTxtRun(np *dcrawl.SpiderContent) {

	bookUrl := map[string]bool{}
	novelChan := make(chan dcrawl.NovelField, 10)
	wait := sync.WaitGroup{}
	downloadNum := 3

	dcrawl.Log.Infof("bookTxt 开始执行,base url: %s", np.BaseUrl)

	/* 获取 url */
	booktxtGetUrl(np.BaseUrl, &(np.SeedUrl), &bookUrl)
	dcrawl.Log.Debugf("bookTxt 获取书籍url成功!")

	/* 解析小说 */
	go booktxtParserInfo(np.BaseUrl, &bookUrl, &novelChan)

	/**
	 *  1. 检测数据库中是否存在 ---- 书名 + 作者名 + 解析器名
	 *  2. 存在则匹配是否需要下载
	 *  3. 下载 && 保存
	 */
	 for i := 0; i < downloadNum; i++ {
	 	wait.Add(1)
	 	go booktxtDownload(np.MI, &wait, &novelChan, np.ToMongo)
	 }

	 wait.Wait()
}
