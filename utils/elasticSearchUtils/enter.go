package elasticSearchUtils

import (
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"strings"
	"tgwp/log/zlog"
)

func Get(client *elasticsearch.Client, index string, id string) (map[string]interface{}, error) {
	// 获取数据
	getResp, err := client.Get(index, id)
	if err != nil {
		zlog.Errorf("获取数据失败: %v", err)
		return nil, err
	}
	defer getResp.Body.Close()

	if getResp.StatusCode == 404 {
		// 返回空，不报错
		return nil, nil
	}

	if getResp.IsError() {
		zlog.Errorf("获取数据失败，状态码: %d", getResp.StatusCode)
		return nil, err
	}

	var r map[string]interface{}
	if err := json.NewDecoder(getResp.Body).Decode(&r); err != nil {
		zlog.Errorf("解析响应体失败: %v", err)
		return nil, err
	}

	return r["_source"].(map[string]interface{}), nil
}

func Update(client *elasticsearch.Client, index string, id string, bodyJson map[string]interface{}) (err error) {
	// 创建数据
	body, err := json.Marshal(
		map[string]interface{}{
			"doc":           bodyJson,
			"doc_as_upsert": true,
		})
	if err != nil {
		zlog.Errorf("创建数据失败: %v", err)
		return
	}
	updateResp, err := client.Update(index, id, bytes.NewReader(body))
	if err != nil {
		zlog.Errorf("创建数据失败: %v", err)
		return
	}
	defer updateResp.Body.Close()

	if updateResp.IsError() {
		zlog.Errorf("创建数据失败，状态码: %d", updateResp.StatusCode)
		zlog.Errorf("响应体: %s", updateResp.String())
		return
	}

	var r map[string]interface{}
	if err = json.NewDecoder(updateResp.Body).Decode(&r); err != nil {
		zlog.Errorf("解析响应体失败: %v", err)
		return
	}

	return
}

func Search(client *elasticsearch.Client, index string, query string) (map[string]interface{}, error) {
	search, err := client.Search(
		client.Search.WithIndex(index),
		client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		zlog.Errorf("搜索失败: %v", err)
		return nil, err
	}
	defer search.Body.Close()

	if search.IsError() {
		zlog.Errorf("搜索失败，状态码: %d", search.StatusCode)
		zlog.Errorf("响应体: %s", search.String())
		return nil, err
	}

	var r map[string]interface{}
	if err = json.NewDecoder(search.Body).Decode(&r); err != nil {
		zlog.Errorf("解析响应体失败: %v", err)
		return nil, err
	}
	return r, nil
}
