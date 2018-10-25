package dcrawl

import (
	"container/list"
	"library/mgo.v2"
	"library/mgo.v2/bson"
	"strconv"
	"strings"
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
//var mongoLock = sync.Mutex{}

func InserDoc(mi SMongoInfo, doc *NovelBean) bool {
	ret := false
	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		config := SMongoConfig{}
		if ok := GetConfig(mi, &config); ok {
			config.NovelNum++
			doc.Info.Id = strconv.Itoa(config.NovelNum)

			cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
			if err = cinfo.Insert(doc.Info); nil == err {

				cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
				for _, info := range doc.Data {

					if err = cdata.Insert(info); nil == err {
						/* 更新配置 */
						if ok := UpdateConfig(mi, &config); ok {
							ret = true
						} else {
							cinfo.RemoveId(doc.Info.Id)
							for _, ei := range doc.Info.Blocks {
								cinfo.RemoveId(ei)
							}
							Log.Errorf("%s|%s|%s 更新 config 失败 ...", doc.Info.NovelParse, doc.Info.Name, doc.Info.Author)
						}
					} else {
						cinfo.RemoveId(doc.Info.Id)
						for _, ei := range doc.Info.Blocks {
							cinfo.RemoveId(ei)
						}
						Log.Errorf("%s|%s|%s 更新 data 失败: %s", doc.Info.NovelParse, doc.Info.Name, doc.Info.Author, err)
						break
					}
				}
			} else {
				cinfo.RemoveId(doc.Info.Id)
				Log.Errorf("%s|%s|%s 更新 info 失败: %s", doc.Info.NovelParse, doc.Info.Name, doc.Info.Author, err)
			}
		} else {
			Log.Errorf("%s|%s|%s 获取 config 失败: %s", doc.Info.NovelParse, doc.Info.Name, doc.Info.Author, err)
		}
	} else {
		Log.Errorf("mongo获取 session 失败: %s", err)
	}

	return ret
}

func UpdateDoc(mi SMongoInfo, id string, doc *NovelBean) bool {
	ret := false
	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		selector := bson.M{"_id": id}
		cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
		if err = cinfo.Update(selector, doc.Info); nil == err {
			cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
			for _, sdata := range doc.Data {
				selectort := bson.M{"_id": sdata.Id}
				if err = cdata.Update(selectort, sdata); nil == err {
					ret = true
				} else {
					Log.Errorf("%s|%s|%s 更新 data 失败: %s", doc.Info.NovelParse, doc.Info.Name, doc.Info.Author, err)
				}
			}
		} else {
			Log.Errorf("mongo更新 info 信息失败: %s", err)
		}
	} else {
		Log.Errorf("mongo获取 session 失败: %s", err)
	}

	return ret
}

func FindDocByField(mi SMongoInfo, field *bson.M, doc *NovelBean) bool {

	ret := false
	ninfo := NovelInfo{}
	ndata := []NovelData{}
	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
		if err = cinfo.Find(*field).One(&ninfo); nil == err {
			ret = true
			cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
			for _, did := range ninfo.Blocks {
				tmp := NovelData{}
				if err = cdata.Find(bson.M{"_id": did}).One(&tmp); nil == err {
					ndata = append(ndata, tmp)
				} else {
					ret = false
					Log.Errorf("%s 获取 data 失败: %s", field, err)
					break
				}
			}
		} else {
			Log.Errorf("%s 获取 info 失败: %s", field, err)
		}
	} else {
		Log.Errorf("%s 获取 session 失败: %s", field, err)
	}

	if ret {
		doc.Info = ninfo
		doc.Data = ndata
	}

	return ret
}

func FindDocById(mi SMongoInfo, id string, doc *NovelBean) bool {

	ret := false
	ninfo := NovelInfo{}
	ndata := []NovelData{}
	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
		if err = cinfo.Find(bson.M{"_id": id}).One(&ninfo); nil == err {
			cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
			for _, did := range ninfo.Blocks {
				tmp := NovelData{}
				if err = cdata.Find(bson.M{"_id": did}).One(&tmp); nil == err {
					ret = true
					ndata = append(ndata, tmp)
				} else {
					ret = false
					Log.Errorf("查找数据 data 失败: %s", err)
					break
				}
			}
		} else {
			Log.Errorf("%s 查找数据 info 失败: %s", id, err)
		}
	} else {
		Log.Errorf("%s 获取 session 失败: %s", id, err)
	}

	if ret {
		doc.Info = ninfo
		doc.Data = ndata
	}

	return ret
}

/**
 *  获取数据库配置信息
 *  1. 目前书籍数量
 *  2. 首页模块
 *   ////////// 2. 分类映射
 */
func GetConfig(mi SMongoInfo, doc *SMongoConfig) bool {

	ret := true
	conf := SMongoConfig{}
	conf.CategoryMap = map[string]string{}
	conf.Id = "conf"
	conf.MainPage = map[string]list.List{}
	conf.NovelNum = 0

	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		mconf := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_conf")
		if err = mconf.Find(bson.M{"_id": "conf"}).One(doc); nil != err {
			if rt := mconf.Insert(conf); nil == rt {
				doc = &conf
			}
		}
	} else {
		ret = false
		Log.Errorf("config 获取 session 失败: %s", err)
	}

	return ret
}

/* 更新配置信息 */
func UpdateConfig(mi SMongoInfo, doc *SMongoConfig) bool {

	ret := false
	conf := SMongoConfig{}
	conf.Id = "conf"

	if session, err := mgo.Dial(getStandaloneUrl(mi)); nil == err {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		mconf := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_conf")
		selector := bson.M{"_id": "conf"}
		if err = mconf.Update(selector, doc); nil == err {
			ret = true
		} else {
			Log.Errorf("config 更新 config 失败: %s", err)
		}
	} else {
		Log.Errorf("config 获取 session 失败: %s", err)
	}

	return ret
}

func getStandaloneUrl(mi SMongoInfo) string {
	url := "mongodb://"
	if "" != mi.Usr {
		url += mi.Usr + ":" + mi.Pwd + "@" + mi.IP + ":" + strconv.Itoa(mi.Port) + "/admin"
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
	/* 1. 是否已有该书籍 */
	novelTmp := NovelBean{}
	fild := bson.M{"name": info.Name, "author": info.Author, "novel_parse": info.NovelParse}
	if ok := FindDocByField(mongo, &fild, &novelTmp); ok {
		/* 已有这一信息 */
		if novelTmp.Info.MaskLevel > 0 {
			Log.Infof("不需要更新的书籍: %s|%s|%s", novelTmp.Info.NovelParse, novelTmp.Info.Name, novelTmp.Info.Author)
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
		} else {
			Log.Errorf("更新 %s|%s|%s 失败!!!", info.NovelParse, info.Name, info.Author)
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
		} else {
			Log.Errorf("插入 %s|%s|%s 失败!!!", info.NovelParse, info.Name, info.Author)
		}
	}
}
