package dcrawl

import (
	"container/list"
	"library/mgo.v2"
	"library/mgo.v2/bson"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SMongoInfo struct {
	IP            string
	Port          int
	Usr           string
	Pwd           string
	DatabaseName  string
	PrefixCollect string
}

type SMongoConfig struct {
	Id          string               `bson:"_id,omitempty"`
	NovelNum    int                  `bson:"novel_num"`
	CategoryMap map[string]string    `bson:"category_mapping"`
	MainPage    map[string]list.List `bson:"main_page"`
}

/* mongodb 锁 */
var mongoLock = sync.Mutex{}

func InserDoc(mi SMongoInfo, doc *NovelBean) bool {
	flag := true
	mongoLock.Lock()
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		mongoLock.Unlock()
		return flag
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	config := SMongoConfig{}
	if ok := GetConfig(mi, &config); !ok {
		flag = false
		mongoLock.Unlock()
		return flag
	}

	doc.Info.Id = strconv.Itoa(config.NovelNum)

	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	err = cinfo.Insert(doc.Info)
	if nil != err {
		flag = false
		Log.Errorf("mongo插入info数据失败: %s", err)
		mongoLock.Unlock()
		return flag
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	for _, info := range doc.Data {
		err = cdata.Insert(info)
		if nil != err {
			flag = false
			Log.Errorf("mongo插入data数据失败: %s", err)
			mongoLock.Unlock()
			return flag
		}
	}

	/* 更新配置 */
	config.NovelNum++
	if ok := UpdateConfig(mi, &config); !ok {
		flag = false
		Log.Errorf("mongo插入config 数据失败")
		mongoLock.Unlock()
		return flag
	}

	mongoLock.Unlock()
	return flag
}

func UpdateDoc(mi SMongoInfo, id string, doc *NovelBean) bool {
	flag := true
	mongoLock.Lock()
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		mongoLock.Unlock()
		return flag
	}

	session.SetMode(mgo.Monotonic, true)

	selector := bson.M{"_id": id}
	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	err = cinfo.Update(selector, doc.Info)
	if nil != err {
		flag = false
		Log.Errorf("mongo更新基本信息失败: %s", err)
		mongoLock.Unlock()
		return flag
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	for _, sdata := range doc.Data {
		selectort := bson.M{"_id": sdata.Id}
		err = cdata.Update(selectort, sdata)
		if nil != err {
			flag = false
			Log.Errorf("mongo更新内容失败: %s", err)
			mongoLock.Unlock()
			return flag
		}
	}

	defer session.Close()

	return flag
}

func FindDocByField(mi SMongoInfo, field *bson.M, doc *NovelBean) bool {
	flag := true
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		return flag
	}

	session.SetMode(mgo.Monotonic, true)

	ninfo := NovelInfo{}
	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")

	err = cinfo.Find(field).One(&ninfo)
	if nil != err {
		flag = false
		Log.Errorf("mongo查找数据info失败: %s", err)
		return flag
	}

	ndata := []NovelData{}
	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")

	for _, did := range ninfo.Blocks {
		tmp := NovelData{}
		err = cdata.Find(bson.M{"_id": did}).One(&tmp)
		if nil == err {
			ndata = append(ndata, tmp)
		} else {
			flag = false
		}
	}

	defer session.Close()

	doc.Info = ninfo
	doc.Data = ndata

	return flag
}

func FindDocById(mi SMongoInfo, id string, doc *NovelBean) bool {
	flag := true

	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		return flag
	}

	session.SetMode(mgo.Monotonic, true)

	ninfo := NovelInfo{}
	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")

	if err = cinfo.Find(bson.M{"_id": id}).One(&ninfo); nil != err {
		flag = false
		Log.Errorf("mongo查找数据info失败: %s", err)
		return flag
	}

	ndata := []NovelData{}
	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")

	for _, did := range ninfo.Blocks {
		tmp := NovelData{}
		if err = cdata.Find(bson.M{"_id": did}).One(&tmp); nil == err {
			ndata = append(ndata, tmp)
		} else {
			flag = false
		}
	}

	defer session.Close()

	doc.Info = ninfo
	doc.Data = ndata

	return flag
}

/**
 *  获取数据库配置信息
 *  1. 目前书籍数量
 *  2. 首页模块
 *   ////////// 2. 分类映射
 */
func GetConfig(mi SMongoInfo, doc *SMongoConfig) bool {

	conf := SMongoConfig{}
	conf.CategoryMap = map[string]string{}
	conf.Id = "conf"
	conf.MainPage = map[string]list.List{}
	conf.NovelNum = 0

	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		Log.Errorf("mongo获取session失败: %s", err)
		return false
	}

	session.SetMode(mgo.Monotonic, true)
	mconf := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_conf")
	if err = mconf.Find(bson.M{"_id": "conf"}).One(doc); nil != err {
		ret := mconf.Insert(conf)
		if nil == ret {
			doc = &conf
			return true
		}
	}
	defer session.Close()

	return true
}

/* 更新配置信息 */
func UpdateConfig(mi SMongoInfo, doc *SMongoConfig) bool {

	conf := SMongoConfig{}
	conf.Id = "conf"

	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		mconf := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_conf")
		selector := bson.M{"_id": "conf"}
		if err = mconf.Update(selector, doc); nil != err {
			Log.Errorf("mongo更新基本信息失败: %s", err)
			return false
		}
	} else {
		Log.Errorf("mongo获取session失败: %s", err)
		return false
	}

	return true
}

func getStandaloneUrl(mi SMongoInfo) string {
	url := "mongodb://"
	if "" != mi.Usr {
		url += mi.Usr + ":" + mi.Pwd + "@" + mi.IP + ":" + strconv.Itoa(mi.Port)
	} else {
		url += mi.IP + ":" + strconv.Itoa(mi.Port)
	}
	return url
}

/**
 *  保存数据到 mongodb
 *  1. 检查数据是否已在数据库存在：name + author + parser
 *  2. 已存在则更新数据库中的数据
 *  3. 未存在则插入数据库
 */
func SaveToMongo(mongo SMongoInfo, info NovelField) {
	Log.Infof("开始保存书籍: %s|%s|%s !!!", info.NovelParse, info.Name, info.Author)
	times := time.Now().Format("20060102150405")

	// 1. 是否已有该书籍
	novelTmp := NovelBean{}
	fild := bson.M{"name": info.Name, "author": info.Author, "novel_parse": info.NovelParse}

	if ok := FindDocByField(mongo, &fild, &novelTmp); ok {
		/* 已有这一信息 */
		if novelTmp.Info.MaskLevel > 0 {
			Log.Infof("不需要更新的书籍: %s|%s", novelTmp.Info.Name, novelTmp.Info.Author)
			return
		}

		/* 更新主要字段 */
		// 图片
		if ("" == novelTmp.Info.ImgType) || (len(novelTmp.Info.ImgContent) <= 10) {
			novelTmp.Info.ImgType = info.ImgType
			novelTmp.Info.ImgContent = info.ImgContent
			novelTmp.Info.ImgUrl = info.ImgUrl
		}
		if "" != info.Tags {
			novelTmp.Info.Tags = info.Tags
		} // 类别
		if "" != info.Status {
			novelTmp.Info.Status = info.Status
		} // 状态
		if "" != info.Desc {
			novelTmp.Info.Desc = info.Desc
		} // 简介

		// 更新时间
		novelTmp.Info.Uptime = times

		// 检查章节数
		chapterAll := map[string]string{}
		for _, data := range novelTmp.Data {
			for k, v := range data.ChapterContent {
				chapterAll[k] = v
			}
		}
		if len(info.ChapterContent) >= len(chapterAll) {
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

			tmp := map[string]bool{}
			for _, iv := range novelTmp.Info.ErrorChapter {
				for mk, _ := range chapterAll {
					cnameE := iv
					cnameR := strings.Split(mk, "{]")[1]
					if cnameE == cnameR {
						tmp[mk] = true
						break
					}
				}
			}
			// 删除已解决的错误章节
			for ik, _ := range tmp {
				delete(novelTmp.Info.ErrorChapter, ik)
			}
		}

		novelTmp.Info.ChapterNum = len(info.ChapterUrl)
		novelTmp.Info.BlockVolume = NovelContentNum
		/* 重新封装章节内容 */
		blockIds := []string{}
		data := []NovelData{}
		GeneratorChapterContent(info.Name, info.Author, novelTmp.Info.NovelParse, chapterAll, &blockIds, &data)
		novelTmp.Info.Blocks = blockIds
		novelTmp.Data = data
		if UpdateDoc(mongo, novelTmp.Info.Id, &novelTmp) {
			Log.Infof("更新 %s|%s|%s 成功!!!", info.NovelParse, info.Name, info.Author)
		}
	} else {
		novelTmp.Info.Name = info.Name
		novelTmp.Info.Author = info.Author
		novelTmp.Info.NovelUrl = info.NovelUrl
		novelTmp.Info.NovelParse = info.NovelParse

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
		novelTmp.Info.BlockVolume = NovelContentNum
		novelTmp.Info.ChapterNum = len(info.ChapterUrl)

		blockIds := []string{}
		data := []NovelData{}
		GeneratorChapterContent(info.Name, info.Author, info.NovelParse, info.ChapterContent, &blockIds, &data)
		novelTmp.Info.Blocks = blockIds
		novelTmp.Data = data

		if InserDoc(mongo, &novelTmp) {
			Log.Infof("插入 %s|%s|%s 成功!!!", info.NovelParse, info.Name, info.Author)
		}
	}
}
