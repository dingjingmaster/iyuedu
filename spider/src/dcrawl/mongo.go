package dcrawl

import (
	"library/mgo.v2"
	"library/mgo.v2/bson"
	"strconv"
)

type SMongoInfo struct {
	IP 								string
	Port 							int
	Usr								string
	Pwd								string
	DatabaseName 					string
	PrefixCollect					string
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

	defer session.Close()

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

	return flag
}


func FindDocById (mi SMongoInfo, id string, doc *NovelBean) bool {
	flag := true
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		flag = false
		Log.Errorf("mongo获取session失败: %s", err)
		return flag
	}

	defer session.Close()

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

	for _, did := range ninfo.Blocks{
		tmp := NovelData{}
		err = cdata.Find(bson.M{"_id":did}).One(&tmp)
		if nil == err{
			ndata = append(ndata, tmp)
		} else {
			flag = false
		}
	}

	doc.Info = ninfo
	doc.Data = ndata

	return flag
}
