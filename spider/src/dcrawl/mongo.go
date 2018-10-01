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

func UpdateDoc(mi SMongoInfo, id bson.ObjectId, doc NovelBean) bool {
	flag := true
	session, error := mgo.Dial(getStandaloneUrl(mi))
	if nil != error {
		panic(error)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	cinfo := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_info")
	error = cinfo.Update(id, doc.Info)
	if nil != error {
		flag = false
	}

	cdata := session.DB(mi.DatabaseName).C(mi.PrefixCollect + "_data")
	error = cdata.Update(id, doc.Data)
	if nil != error {
		flag = false
	}
	return flag
}

