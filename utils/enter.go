package utils

import (
	"encoding/json"
	"math/rand"
	"path/filepath"
	"regexp"
	"runtime"
	"tgwp/log/zlog"
	"time"
)

/*
GetRootPath 搜索项目的文件根目录, 并和 myPath 拼接起来
*/
func GetRootPath(myPath string) string {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		panic("Something wrong with getting root path")
	}
	absPath, err := filepath.Abs(fileName)
	rootPath := filepath.Dir(filepath.Dir(absPath))
	if err != nil {
		panic(any(err))
	}
	return filepath.Join(rootPath, myPath)
}

// StructToMap
//
//	@Description: struct to map
//	@param value
//	@return map[string]interface{}
func StructToMap(value interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	resJson, err := json.Marshal(value)
	if err != nil {
		zlog.Errorf("Json Marshal failed ,msg: %s", err.Error())
		return nil
	}
	err = json.Unmarshal(resJson, &m)
	if err != nil {
		zlog.Errorf("Json Unmarshal failed,msg : %s", err.Error())
		return nil
	}
	return m
}

// StuctToJson
//
//	@Description: struct to json
//	@param value
//	@return string
//	@return error
func StuctToJson(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), err
}

// JsonToStruct
//
//	@Description: json to struct
//	@param str
//	@param value
//	@return error
func JsonToStruct(str string, value interface{}) error {
	return json.Unmarshal([]byte(str), value)
}

// RandomCode
//
//	@Description: 生成随机码
//	@return string
func RandomCode() string {
	const charset = "0123456789abcdefghijklmnopqrlstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 6)
	rand.Seed(time.Now().UnixNano())
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// IdentifyPhone
//
//	@Description: 判定是否为中国手机号
//	@param phone
//	@return bool
func IdentifyPhone(phone string) bool {
	var phoneRegex = regexp.MustCompile(`^1(3[0-9]|4[57]|5[0-35-9]|7[0-9]|8[0-9]|9[8])\d{8}$`)
	return phoneRegex.MatchString(phone)
}

// RecordTime a tool to record time
// e.g [defer utils.RecordTime(time.Now())()]
func RecordTime(start time.Time) func() {
	return func() {
		end := time.Now()
		zlog.Debugf("use time:%d", end.Unix()-start.Unix())
	}
}
