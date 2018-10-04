package dcrawl

import (
	"fmt"
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
	fmt.Println(url)
	return url
}

func InserDoc(mi SMongoInfo, doc *NovelBean) bool {
	flag := true
	session, error := mgo.Dial(getStandaloneUrl(mi))
	if nil != error {
		panic(error)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	error = cinfo.Insert(doc.Info)
	if nil != error {
		flag = false
		panic(error)
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	for _, info := range doc.Data {
		error = cdata.Insert(info)
		if nil != error {
			flag = false
			panic(error)
		}
	}
	return flag
}

func UpdateDoc(mi SMongoInfo, id bson.ObjectId, doc *NovelBean) bool {
	flag := true
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	err = cinfo.Update(id, doc.Info)
	if nil != err {
		flag = false
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	err = cdata.Update(id, doc.Data)
	if nil != err {
		flag = false
	}
	return flag
}


func FindDocById (mi SMongoInfo, id string, doc *NovelBean) bool {
	flag := true
	session, err := mgo.Dial(getStandaloneUrl(mi))
	if nil != err {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	ninfo := NovelInfo{}
	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")

	err = cinfo.Find(bson.M{"_id": bson.ObjectId(id)}).One(&ninfo)
	if nil != err {
		flag = false
		return flag
	}

	ndata := []NovelData{}
	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")

	for _, did := range ninfo.Blocks{
		tmp := NovelData{}
		err = cdata.Find(bson.M{"_id":bson.ObjectId(did)}).One(&tmp)
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
