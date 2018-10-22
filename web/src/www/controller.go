package www

import "github.com/kataras/iris"


func Routing(web *iris.Application, logLevel string) {
	log := web.Logger().SetLevel(logLevel)

	index := iris.HTML("./template", ".html")
	web.RegisterView(index)

	/* 首页展示 */
	web.Get("/", func (ctx iris.Context) {

		ctx.View("index.html")

		log.Infof("请求/")
	})

}
