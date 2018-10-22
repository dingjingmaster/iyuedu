package dcrawl

import (
	"container/list"
	"library/mgo.v2"
	"library/mgo.v2/bson"
	"strconv"
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
	id          string               `bson:"_id,omitempty"`
	NovelNum    int                  `bson:"novel_num"`
	CategoryMap map[string]string    `bson:"category_mapping"`
	MainPage    map[string]list.List `bson:"main_page"`
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

func InserDoc(mi SMongoInfo, doc *NovelBean) bool {
	flag := true
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		return flag
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	err = cinfo.Insert(doc.Info)
	if nil != err {
		flag = false
		Log.Errorf("mongo插入info数据失败: %s", err)
		return flag
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	for _, info := range doc.Data {
		err = cdata.Insert(info)
		if nil != err {
			flag = false
			Log.Errorf("mongo插入data数据失败: %s", err)
			return flag
		}
	}

	return flag
}

func UpdateDoc(mi SMongoInfo, id string, doc *NovelBean) bool {
	flag := true
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		return flag
	}

	session.SetMode(mgo.Monotonic, true)

	selector := bson.M{"_id": id}
	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	err = cinfo.Update(selector, doc.Info)
	if nil != err {
		flag = false
		Log.Errorf("mongo更新基本信息失败: %s", err)
		return flag
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	for _, sdata := range doc.Data {
		selectort := bson.M{"_id": sdata.Id}
		err = cdata.Update(selectort, sdata)
		if nil != err {
			flag = false
			Log.Errorf("mongo更新内容失败: %s", err)
			return flag
		}
	}

	defer session.Close()

	return flag
}

func FindDocByField (mi SMongoInfo, field *bson.M, doc *NovelBean) bool {
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

	err = cinfo.Find(bson.M{"_id": id}).One(&ninfo)
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

/**
 *  获取数据库配置信息
 *  1. 目前书籍数量
 *  2. 首页模块
 *   ////////// 2. 分类映射
 */
func GetConfig(mi SMongoInfo, doc *SMongoConfig) bool {

	conf := SMongoConfig{}
	conf.CategoryMap = map[string]string{}
	conf.id = "conf"
	conf.MainPage = map[string]list.List{}
	conf.NovelNum = 0

	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		Log.Errorf("mongo获取session失败: %s", err)
		return false
	}

	session.SetMode(mgo.Monotonic, true)
	mconf := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_conf")
	err = mconf.Find(bson.M{"_id": "conf"}).One(doc)
	if nil != err {
		ret := mconf.Insert(conf)
		if nil == ret {
			doc = & conf
			return true
		}
	}
	defer session.Close()

	return true
}

/* 更新配置信息 */
func UpdateConfig(mi SMongoInfo, doc *SMongoConfig) bool {

	conf := SMongoConfig{}
	conf.id = "conf"

	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		Log.Errorf("mongo获取session失败: %s", err)
		return false
	}

	session.SetMode(mgo.Monotonic, true)
	mconf := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_conf")
	selector := bson.M{"_id": "conf"}
	err = mconf.Update(selector, doc)
	if nil != err {
		Log.Errorf("mongo更新基本信息失败: %s", err)
		return false
	}

	defer session.Close()

	return true
}
