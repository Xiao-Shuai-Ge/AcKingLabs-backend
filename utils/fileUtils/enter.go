package fileUtils

import (
	"errors"
	"strings"
	"tgwp/utils/listUtils"
)

var whiteList = []string{"jpg", "png", "jpeg", "gif", "webp"}

func ImageSuffixJudge(fileName string) (suffix string, err error) {
	_list := strings.Split(fileName, ".")
	length := len(_list)
	if length == 1 {
		err = errors.New("文件名错误")
		return
	}
	suffix = _list[length-1]
	if !listUtils.InList(whiteList, suffix) {
		err = errors.New("文件格式错误")
		return
	}
	return
}
