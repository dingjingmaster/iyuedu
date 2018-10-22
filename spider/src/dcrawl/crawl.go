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
	"crypto/tls"
	"io"
	"io/ioutil"
	"library/goquery"
	"library/mahonia"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
)

//

type SpiderRun func(np *SpiderContent)

type SpiderContent struct {
	BaseUrl    string         // 首页url
	SpiderName string         // 爬虫名字
	SeedUrl    map[string]int // 种子 url
	MI         SMongoInfo     // mongodb 连接信息
	Exit       bool           // 检测是否可以退出
}

var SpiderGroup = sync.WaitGroup{}


/* 获取 url get 请求 */
func GetHTMLByUrl(url string, head *map[string]string) (bool, string) {

	ret := false
	html := ""

	// 获取请求类型
	rtp, _ := CheckHTTP(url)
	if "https" == rtp {
		ret, html = GetHTTPSRequest(url, head)
	} else {
		ret, html = GetHTTPRequest(url, head)
	}

	if !ret {
		Log.Errorf("url: %s 请求错误!", url)
	}

	return ret, html
}




// 根据 url 获取页面 html 字符串
func Post(url *string, head *map[string]string, para *map[string]string) (bool, string) {
	charset := "utf8"
	ret := true

RET:
	if !ret {
		return ret, ""
	}

	if (nil == url) || ("" == *url) || (nil == para) {
		Log.Errorf("错误的输入参数: %s", url)
		goto RET
	}

	paras := []string{}
	for k, v := range *para {
		paras = append(paras, k+"="+v)
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
	if (nil != err) || (200 != resp.StatusCode) {
		ret = false
		Log.Errorf("无法访问url: %s|%s\n", url, err)
		goto RET
	}
	defer resp.Body.Close()
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










/* 请求 http 返回 html */
func GetHTTPRequest (url string, head *map[string]string) (bool, string) {
	ret := false
	body := ""

	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if nil != err {
		Log.Errorf("错误的 get 请求: %s|%s", url, err)
	}
	if nil != head {
		for k, v := range *head {
			request.Header.Add(k, v)
		}
	}
	resp, err := client.Do(request)
	if (nil != err) || (200 != resp.StatusCode) {
		ret = false
		Log.Errorf("url访问错误: %s|%s\n", url, err)
	} else {
		ret = true
	}
	defer resp.Body.Close()

	if ret {
		err, html := ReadByteToString(&(resp.Body))
		if err {
			body = ConvertToString(string(html), GetHTMLCharset(&html), "utf8")
		} else {
			ret = false
		}
	}

	return ret, body
}

/* 请求 https 返回 html */
func GetHTTPSRequest (url string, head *map[string]string) (bool, string) {
	ret := false
	body := ""
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	request, err := http.NewRequest("GET", url, nil)
	if nil != err {
		Log.Errorf("错误的 get 请求: %s|%s", url, err)
	}
	if nil != head {
		for k, v := range *head {
			request.Header.Add(k, v)
		}
	}
	resp, err := client.Do(request)
	if (nil != err) || (200 != resp.StatusCode) {
		ret = false
		Log.Errorf("url访问错误: %s|%s\n", url, err)
	} else {
		ret = true
	}
	defer resp.Body.Close()

	if ret {
		tret, html := ReadByteToString(&(resp.Body))
		if tret {
			body = ConvertToString(string(html), GetHTMLCharset(&html), "utf8")
		} else {
			ret = false
		}
	}

	return ret, body
}


/* 没有任何依赖的工具函数 */

/* 从 []byte 读取数据并转为 strig */
func ReadByteToString(bt *io.ReadCloser) (bool, string) {
	ret := false
	body := ""

	tmp, err := ioutil.ReadAll(*bt)
	if nil != err {
		ret = false
		Log.Errorf("io读取错误！%s", err)
	} else {
		ret = true
		body = string(tmp)
	}

	return ret, body
}

/* 获取页面的字符编码，返回字符编码 */
func GetHTMLCharset (html *string) string {
	ret := false
	cs := "utf8"

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(*html)))
	if nil == err {
		ret = true
	}

	if ret {
		ic, errs := doc.Find("head").Find("meta[http-equiv]").Attr("content")
		if errs {
			ic = strings.TrimSpace(ic)
			ict := strings.Split(ic, "=")
			ic = ict[len(ict)-1]
			if "" != ic {
				cs = ic
			}
		}
	}

	return cs
}

/* 检查请求类型，是 http 还是 https，返回请求类型和 url */
func CheckHTTP (url string) (string, string) {
	reqType := "http"
	arr := strings.SplitN(url, "://", -1)
	if (len(arr) >= 2) && ("https" == arr[0]) {
		reqType = "https"
	}

	return reqType, url
}

/* 字符串编码转换 */
func ConvertToString(src string, srcCode string, tagCode string) string {
	result := ""
	// 修正网站上编码字符串错误问题
	codes := map[string]string {
		"gb2312": "gbk",
	}

	if v, exits := codes[srcCode]; exits {
		srcCode = v
	}

	if v, exits := codes[tagCode]; exits {
		tagCode = v
	}

	srcResult := mahonia.NewDecoder(srcCode).ConvertString(src)
	_, cdata, err := mahonia.NewDecoder(tagCode).Translate([]byte(srcResult), true)
	if nil == err {
		result = string(cdata)
	}

	return result
}
