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
	SeedUrl					map[string]int						// 种子 url
}

var SpiderGroup = sync.WaitGroup{}









// 根据 url 获取页面 html 字符串
func GetHTMLByUrl(url *string) (bool, string) {
	charset := "utf8"
	ret := true

	RET: if !ret {
			return ret, ""
	}

	if (nil == url) || ("" == *url) {
		goto RET
	}
	resp, err := http.Get(*url)
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

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}



