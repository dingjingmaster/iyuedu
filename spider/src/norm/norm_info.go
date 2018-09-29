package norm

import (
	"regexp"
	"strings"
)

func NormName(mstr string) string {
	// 删除字符开头
	reg := regexp.MustCompile(`^(书名是|书名)`)
	mstr = reg.ReplaceAllString(mstr, "")
	// 删除整个字符
	reg = regexp.MustCompile(`[/.@!~#$%^&*:";?\\+=-_,{}\[\]<>！￥…（）—=、“”：；？。，《》]`)
	mstr = reg.ReplaceAllString(mstr, "")

	return strings.TrimSpace(mstr)
}

func NormAuthor(mstr string) string {

	// 删除字符开头
	reg := regexp.MustCompile(`^(作者名是|作者名|作者)`)
	mstr = reg.ReplaceAllString(mstr, "")
	// 删除整个字符
	reg = regexp.MustCompile(`[/.@!#$%^&*:";_,{}<>！￥…（）“”：；。，《》]`)
	mstr = reg.ReplaceAllString(mstr, "")

	return mstr
}


func NormCategory(mstr string) string {

	// 删除字符开头
	reg := regexp.MustCompile(`^(类型是|类型)`)
	mstr = reg.ReplaceAllString(mstr, "")
	// 删除整个字符
	reg = regexp.MustCompile(`[/.@!~#$%^&*:";?\\+=-_,{}\[\]<>！￥…（）—=、“”：；？。，《》]`)
	mstr = reg.ReplaceAllString(mstr, "")

	return mstr
}

func NormStatus(mstr string) string {

	// 删除字符开头
	reg := regexp.MustCompile(`^(状态是|状态)`)
	mstr = reg.ReplaceAllString(mstr, "")
	// 删除整个字符
	reg = regexp.MustCompile(`[/.@!~#$%^&*:";?\\+=-_,{}\[\]<>！￥…（）—=、“”：；？。，《》]`)
	mstr = reg.ReplaceAllString(mstr, "")

	if strings.Contains(mstr, "完结") {
		mstr = "完结"
	} else {
		mstr = "连载"
	}

	return mstr
}

func NormDesc(mstr string) string {

	// 删除字符开头
	reg := regexp.MustCompile(`^(作品简介|简介|描述)`)
	mstr = reg.ReplaceAllString(mstr, "")
	// 删除整个字符
	reg = regexp.MustCompile(`[\s\v ]+`)
	mstr = reg.ReplaceAllString(mstr, "")

	return mstr
}

func NormChapterName(mstr string) string {

	// 检测括号
	mstrt := mstr
	reg := regexp.MustCompile(`[\[{【]\s*\S+[}\]】]$`)
	mstr = reg.ReplaceAllString(mstr, "")
	if "" == mstr {mstr = mstrt}

	mstrt = mstr
	reg = regexp.MustCompile(`[（(]\s*\S+[)）]$`)
	for _,mmtr := range reg.FindAllString(mstr, -1) {
		subreg := regexp.MustCompile(`月票|感谢|求推荐|求点击|求订阅|求票|第\S更|加更|打赏`)
		if len(subreg.FindAllString(mmtr, -1)) > 0 {
			mstr = reg.ReplaceAllString(mstr, "")
		}
	}
	if "" == mstr {mstr = mstrt}
	// 删除整个字符
	reg = regexp.MustCompile(`[/.@!~#$%^&*:";?\\+=-_,￥—=]`)
	mstr = reg.ReplaceAllString(mstr, "")

	return mstr
}