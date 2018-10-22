package main

import (
	"github.com/kataras/iris"
	"os"
	"www"
)

/**
 *	服务器主入口
 */
func main()  {
	/* 全局配置 */
	//workPath := "E:/GitHub/iyuedu/www"
	workPath := "E:/GitHub/iyuedu/web"
	logLevel := "debug"

	/* 切换路径 */
	os.Chdir(workPath)

	/* 开始 www 服务器 */
	web := iris.New()

	/* 配置 html 路径 */
	web.StaticWeb("/css", "./css")
	web.StaticWeb("/js", "./js")
	web.StaticWeb("/img", "./img")

	/* 配置 iris */
	web.Configure(iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:false,
		DisableInterruptHandler:true,
		DisablePathCorrection:false,
		EnablePathEscape:false,
		FireMethodNotAllowed:true,
		DisableBodyConsumptionOnUnmarshal:false,
		DisableAutoFireStatusCode:false,
		TimeFormat:"2006-01-02 15:04:05",
		Charset:"UTF-8",
	}))

	/* 处理请求 */
	www.Routing(web, logLevel)

	web.Run(iris.Addr(":8080"))
}
