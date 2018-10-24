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
)

//

type SpiderRun func(np *SpiderContent)

type SpiderContent struct {
	BaseUrl    string           // 首页url
	SpiderName string           // 爬虫名字
	SeedUrl    map[string]int   // 种子 url
	MI         SMongoInfo       // mongodb 连接信息
	ToMongo    *chan NovelField // 保存到 mongodb
}

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

	return ret, html
}

func Post(url string, head *map[string]string, para *map[string]string) (bool, string) {
	ret := false
	html := ""
	client := &http.Client{}
	paras := []string{}
	for k, v := range *para {
		paras = append(paras, k + "=" + v)
	}

	rtp, _ := CheckHTTP(url)
	if "https" == rtp {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cookieJar, _ := cookiejar.New(nil)
		client = &http.Client{
			Jar:       cookieJar,
			Transport: tr,
		}
	}
	if request, err := http.NewRequest("POST", url, strings.NewReader(strings.Join(paras, "&"))); nil == err {
		if nil != head {
			for k, v := range *head {
				request.Header.Add(k, v)
			}
		}
		if resp, err := client.Do(request); (nil == err) && (200 == resp.StatusCode) {
			defer resp.Body.Close()
			if err, htmp := ReadByteToString(&(resp.Body)); err {
				ret = true
				html = ConvertToString(string(htmp), GetHTMLCharset(&htmp), "utf8")
			}
		} else {
			Log.Errorf("无法访问url: %s|%s", url, err)
		}
	} else {
		Log.Errorf("错误的 request post: %s|%s", url, err)
	}

	return ret, html
}

/* 请求 http 返回 html */
func GetHTTPRequest(url string, head *map[string]string) (bool, string) {
	ret := false
	body := ""

	client := &http.Client{}
	if request, err := http.NewRequest("GET", url, nil); nil == err {
		if nil != head {
			for k, v := range *head {
				request.Header.Add(k, v)
			}
		}
		if resp, err := client.Do(request); (nil == err) && (200 == resp.StatusCode) {
			defer resp.Body.Close()
			ret = true
			if rt, html := ReadByteToString(&(resp.Body)); rt {
				body = ConvertToString(string(html), GetHTMLCharset(&html), "utf8")
			}
		} else {
			Log.Errorf("url访问错误: %s | %s", url, err)
		}

	} else {
		Log.Errorf("错误的 get 请求: %s | %s", url, err)
	}

	return ret, body
}

/* 请求 https 返回 html */
func GetHTTPSRequest(url string, head *map[string]string) (bool, string) {
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

	if request, err := http.NewRequest("GET", url, nil); nil == err {
		if nil != head {
			for k, v := range *head {
				request.Header.Add(k, v)
			}
		}
		if resp, err := client.Do(request); (nil == err) && (200 == resp.StatusCode) {
			if tret, html := ReadByteToString(&(resp.Body)); tret {
				body = ConvertToString(string(html), GetHTMLCharset(&html), "utf8")
				ret = true
			}
		} else {
			Log.Errorf("url访问错误: %s|%s\n", url, err)
		}

	} else {
		Log.Errorf("错误的 get 请求: %s|%s", url, err)
	}

	return ret, body
}

/* 没有任何依赖的工具函数 */

/* 从 []byte 读取数据并转为 strig */
func ReadByteToString(bt *io.ReadCloser) (bool, string) {
	ret := false
	body := ""

	if tmp, err := ioutil.ReadAll(*bt); nil == err {
		ret = true
		body = string(tmp)
	} else {
		Log.Errorf("io读取错误！%s", err)
	}

	return ret, body
}

/* 获取页面的字符编码，返回字符编码 */
func GetHTMLCharset(html *string) string {
	cs := "utf8"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(*html)))
	if nil == err {
		if ic, ok := doc.Find("head").Find("meta[http-equiv]").Attr("content"); ok{
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
func CheckHTTP(url string) (string, string) {
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
	codes := map[string]string{
		"gb2312": "gbk",
	}

	if v, exits := codes[srcCode]; exits {
		srcCode = v
	}

	if v, exits := codes[tagCode]; exits {
		tagCode = v
	}

	srcResult := mahonia.NewDecoder(srcCode).ConvertString(src)
	if _, cdata, err := mahonia.NewDecoder(tagCode).Translate([]byte(srcResult), true); nil == err {
		result = string(cdata)
	}
	return result
}
