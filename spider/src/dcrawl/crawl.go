/**
 *    主要的爬取数据流程，具体包含以下模块
 *		1. url 发现 及 url 正则匹配
 *		2. 网页 HTML 下载
 *		3. 网页 文件 下载
 * 		4. 网站解析
 *		5. mongodb 存储
 */
package dcrawl

import (
	"io/ioutil"
	"library/goquery"
	"library/mahonia"
	"net/http"
	"strings"
	"sync"
)

//

type SpiderRun func(np *SpiderContent)

type SpiderContent struct {
	BaseUrl					string								// 首页url
	SpiderName				string								// 爬虫名字
	SeedUrl					map[string]int						// 种子 url
	MI						SMongoInfo							// mongodb 连接信息
}

var SpiderGroup = sync.WaitGroup{}



// 根据 url 获取页面 html 字符串
func GetHTMLByUrl(url *string, head *map[string]string) (bool, string) {
	charset := "utf8"
	ret := true

	RET: if !ret {
			return ret, ""
	}

	if (nil == url) || ("" == *url) {
		goto RET
	}

	client := &http.Client{}
	request, err := http.NewRequest("GET", *url, nil)
	if nil != err {
		Log.Errorf("错误的request get: %s|%s", url, err)
		ret = false
		goto RET
	}
	if nil != head {
		for k, v := range *head {
			request.Header.Add(k, v)
		}
	}

	resp, err := client.Do(request)
	defer resp.Body.Close()
	if (nil != err) || (200 != resp.StatusCode) {
		ret = false
		Log.Errorf("无法访问url: %s|%s\n", url, err)
		resp.Body.Close()
		goto RET
	}
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		ret = false
		Log.Errorf("byte转string失败:%s\n", url)
		goto RET
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if nil != err {
		ret = false
		Log.Errorf("HTML转string失败%s\n", url)
		goto RET
	}

	// 获取页面编码
	ic, ret := doc.Find("head").Find("meta[http-equiv]").Attr("content")
	if ret {
		ic = strings.TrimSpace(ic)
		ict := strings.Split(ic, "=")
		ic = ict[len(ict)-1]
		if "" != ic {
			charset = ic
		}
	}

	return ret, ConvertToString(string(body), charset, "utf8")
}


// 根据 url 获取页面 html 字符串
func Post(url *string, head *map[string]string, para *map[string]string) (bool, string) {
	charset := "utf8"
	ret := true

RET: if !ret {
	return ret, ""
}

	if (nil == url) || ("" == *url) || (nil == para) {
		Log.Errorf("错误的输入参数: %s", url)
		goto RET
	}

	paras := []string{}
	for k, v := range *para{
		paras = append(paras,  k + "=" + v)
	}
	client := &http.Client{}
	request, err := http.NewRequest("POST", *url, strings.NewReader(strings.Join(paras, "&")))
	if nil != err {
		Log.Errorf("错误的request post: %s|%s", url, err)
		ret = false
		goto RET
	}
	if nil != head {
		for k, v := range *head {
			request.Header.Add(k, v)
		}
	}

	resp, err := client.Do(request)
	defer resp.Body.Close()
	if (nil != err) || (200 != resp.StatusCode) {
		ret = false
		Log.Errorf("无法访问url: %s|%s\n", url, err)
		resp.Body.Close()
		goto RET
	}
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		ret = false
		Log.Errorf("byte转string失败:%s\n", url)
		goto RET
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if nil != err {
		ret = false
		Log.Errorf("HTML转string失败%s\n", url)
		goto RET
	}

	// 获取页面编码
	ic, errs := doc.Find("head").Find("meta[http-equiv]").Attr("content")
	if errs {
		ic = strings.TrimSpace(ic)
		ict := strings.Split(ic, "=")
		ic = ict[len(ict)-1]
		if "" != ic {
			charset = ic
		}
	}

	return ret, ConvertToString(string(body), charset, "utf8")
}



func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}



