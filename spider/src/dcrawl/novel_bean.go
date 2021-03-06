package dcrawl

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

var NovelContentNum = 600

type NovelField struct {
	Name     string
	Author   string
	NovelUrl string

	NovelParse string

	ImgUrl     string
	ImgContent []byte
	ImgType    string

	Tags   string
	Status string
	Desc   string

	ChapterUrl      map[string]string
	ErrorChapterUrl map[string]string

	ChapterContent map[string]string
}

type NovelInfo struct {
	Id         string `bson:"_id,omitempty"` // 主键 数字
	Name       string `bson:"name"`          // 书名
	Author     string `bson:"author"`        // 作者名
	NormName   string `bson:"norm_name"`     // 归一书名
	NormAuthor string `bson:"norm_author"`   // 归一作者名
	NovelUrl   string `bson:"novel_url"`     // 书籍url
	NovelParse string `bson:"novel_parse"`   // 解析器

	ImgUrl     string `bson:"img_url"`     // 封面页链接
	ImgContent []byte `bson:"img_content"` // 图片内容
	ImgType    string `bson:"img_type"`    // 图片类型

	Category string `bson:"category"` // 分类 3级分类
	Tags     string `bson:"tags"`     // 标签 抓取类别
	Status   string `bson:"status"`   // 连载/完结状态
	Desc     string `bson:"desc"`     // 描述

	ChapterUrl   map[string]string `bson:"chapter_url"`   // 章节链接
	ErrorChapter map[string]string `bson:"error_chapter"` // 错误章节

	Intime string `bson:"intime"` // 入库时间
	Uptime string `bson:"uptime"` // 更新时间

	MaskLevel   int      `bson:"mask_level"`   // 是否限制 1-表示不再展示 2-表示精修展示
	Blocks      []string `bson:"blocks"`       // 子块 id
	BlockVolume int      `bson:"block_volume"` // 子块容量
	ChapterNum  int      `bson:"chapter_num"`  // 总章节数
}

type NovelData struct {
	Id             string            `bson:"_id,omitempty"`   // 主键
	Name           string            `bson:"name"`            // 书名
	Author         string            `bson:"author"`          // 作者名
	ChapterContent map[string]string `bson:"chapter_content"` // 章节信息
}

type NovelBean struct {
	Info NovelInfo
	Data []NovelData
}

/* 封装章节内容信息 */
func GeneratorChapterContent(name string, author string, spiderName string, chapterContent map[string]string, blockIds *[]string, data *[]NovelData) {

	flag := 0
	count := 0

	tmpContent := map[string]string{}
	for cname, ccontent := range chapterContent {
		flag++
		tmpContent[cname] = ccontent
		if (flag % NovelContentNum) == 0 {
			count++
			tmd5 := md5.New()
			novelDatat := NovelData{}
			tmd5.Write([]byte(name + author + spiderName + strconv.Itoa(count)))
			id := hex.EncodeToString(tmd5.Sum(nil))
			novelDatat.Id = string(id)
			novelDatat.Name = name
			novelDatat.Author = author
			novelDatat.ChapterContent = tmpContent
			tmpContent = map[string]string{}
			*data = append(*data, novelDatat)
			*blockIds = append(*blockIds, string(id))
		}
	}

	if len(tmpContent) > 0 {
		count++
		tmd5 := md5.New()
		novelDatat := NovelData{}
		tmd5.Write([]byte(name + author + spiderName + strconv.Itoa(count)))
		id := hex.EncodeToString(tmd5.Sum(nil))
		novelDatat.Id = string(id)
		novelDatat.Name = name
		novelDatat.Author = author
		novelDatat.ChapterContent = tmpContent
		tmpContent = nil
		*data = append(*data, novelDatat)
		*blockIds = append(*blockIds, string(id))
	}
}
