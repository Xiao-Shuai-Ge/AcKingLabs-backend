package contest

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/levigross/grequests"
	"log"
	"strconv"
	"strings"
	"tgwp/model"
	"time"
)

type CodeforcesAPIResponse struct {
	Status string `json:"status"`
	Result []struct {
		ID               int    `json:"id"`
		Name             string `json:"name"`
		Type             string `json:"type"`
		Phase            string `json:"phase"`
		StartTimeSeconds int64  `json:"startTimeSeconds"`
		DurationSeconds  int64  `json:"durationSeconds"`
	} `json:"result"`
}

func GetCodeForcesContest() (contests []model.Contest) {
	resp, err := grequests.Get("https://codeforces.com/api/contest.list?gym=false", nil)
	if err != nil {
		log.Printf("API请求失败: %v", err)
		return
	}

	var apiResponse CodeforcesAPIResponse
	if err := json.Unmarshal(resp.Bytes(), &apiResponse); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return
	}

	if apiResponse.Status != "OK" {
		log.Printf("API返回异常状态: %s", apiResponse.Status)
		return
	}

	for _, cfContest := range apiResponse.Result {
		if cfContest.Phase != "BEFORE" || cfContest.StartTimeSeconds*1000 > time.Now().Unix()*1000+7*24*3600*1000 {
			continue
		}

		contests = append(contests, model.Contest{
			Platform:  "Codeforces",
			Title:     cfContest.Name,
			StartTime: cfContest.StartTimeSeconds * 1000,
			EndTime:   cfContest.StartTimeSeconds*1000 + cfContest.DurationSeconds*1000,
			Duration:  cfContest.DurationSeconds,
			Url:       "https://codeforces.com/contests/" + strconv.Itoa(cfContest.ID),
		})
	}
	return
}

type AtcoderAPIResponse []struct {
	ID               string `json:"id"`
	StartEpochSecond int64  `json:"start_epoch_second"`
	DurationSecond   int64  `json:"duration_second"`
	Title            string `json:"title"`
}

func GetAtCoderContest() (contests []model.Contest) {
	resp, err := grequests.Get("https://atcoder.jp/contests/", nil)
	if err != nil {
		log.Printf("AtCoder页面请求失败: %v", err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp)
	if err != nil {
		log.Printf("HTML解析失败: %v", err)
		return
	}

	// 解析比赛表格（参考网页1、9的表格结构）
	doc.Find("div.table-responsive table tbody tr").Each(func(i int, s *goquery.Selection) {
		cols := s.Find("td")
		if cols.Length() < 4 {
			return
		}

		// 解析时间（示例格式：2025-05-17(Sat) 20:00）
		startTimeStr := cols.Eq(0).Find("a").First().Text()
		startTime, err := time.Parse("2006-01-02 15:04", startTimeStr[:16])
		if err != nil {
			log.Printf("时间解析失败: %v", err)
			return
		}

		// 转换为UTC时间戳（网页9提到时差问题）
		startTime = startTime.Add(-time.Hour) // 日本时区转UTC+8
		startTimestamp := startTime.Unix() * 1000
		startTimestamp -= 8 * 3600 * 1000 // 日本时区转UTC+8

		// 解析持续时间（示例格式：01:40）
		durationStr := cols.Eq(2).Text()
		var h, m int
		_, err = fmt.Sscanf(durationStr, "%d:%d", &h, &m)
		if err != nil {
			log.Printf("持续时间解析失败: %v", err)
			return
		}
		duration := int64(h*3600 + m*60)

		fmt.Println(cols.Eq(1).Find("a").Text(), time.Now().Unix()*1000, startTimestamp, (time.Now().Unix()*1000)+7*24*3600*1000)
		if startTimestamp < time.Now().Unix()*1000 || startTimestamp > (time.Now().Unix()*1000)+7*24*3600*1000 {
			return // 已经结束的比赛不显示
		}

		// 仅筛选 AtCoder Beginner Contest 和 AtCoder Regular Contest
		title := cols.Eq(1).Find("a").First().Text()
		if !(strings.HasPrefix(title, "AtCoder Beginner Contest") || strings.HasPrefix(title, "AtCoder Regular Contest")) {
			return
		}

		contests = append(contests, model.Contest{
			Platform:  "AtCoder",
			Title:     cols.Eq(1).Find("a").Text(),
			StartTime: startTimestamp,
			EndTime:   startTimestamp + duration*1000,
			Duration:  duration,
			Url:       "https://atcoder.jp" + cols.Eq(1).Find("a").AttrOr("href", ""),
		})
	})
	return
}

// 牛客网API响应结构体
type NowcoderAPIResponse struct {
	Code int `json:"code"`
	Data []struct {
		ContestId   int    `json:"contestId"`
		ContestName string `json:"contestName"`
		StartTime   int64  `json:"startTime"`   // 毫秒级时间戳
		EndTime     int64  `json:"endTime"`     // 毫秒级时间戳
		Type        int    `json:"contestType"` // 比赛类型
		IsRegister  bool   `json:"isRegister"`  // 是否报名
		Link        string `json:"link"`        // 链接
	} `json:"data"`
}

func GetNowcoderContest() (contests []model.Contest) {
	currentTime := time.Now()
	params := map[string]string{
		"token": "",
		"month": fmt.Sprintf("%d-%d", currentTime.Year(), currentTime.Month()),
		"_":     fmt.Sprintf("%.3f", float64(time.Now().UnixNano())/1e9), // 三位小数时间戳
	}

	// 发送带参数的GET请求
	resp, err := grequests.Get("https://ac.nowcoder.com/acm/calendar/contest",
		&grequests.RequestOptions{
			Params: params,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
		})

	if err != nil {
		log.Printf("牛客网API请求失败: %v", err)
		return
	}

	// 解析JSON响应
	var apiResponse NowcoderAPIResponse
	if err := json.Unmarshal(resp.Bytes(), &apiResponse); err != nil {
		log.Printf("牛客网JSON解析失败: %v", err)
		return
	}

	// 转换数据格式
	for _, ncContest := range apiResponse.Data {
		// 过滤已结束的比赛（参考网页6的接口数据特征）
		// zlog.Debugf("Nowcoder contest: %v", ncContest)
		if time.Now().UnixMilli() >= ncContest.StartTime || time.Now().UnixMilli()+7*24*3600*1000 <= ncContest.StartTime {
			continue
		}

		//fmt.Println(ncContest.ContestId)
		contests = append(contests, model.Contest{
			Platform:  "Nowcoder",
			Title:     ncContest.ContestName,
			StartTime: ncContest.StartTime,
			EndTime:   ncContest.EndTime,
			Duration:  (ncContest.EndTime - ncContest.StartTime) / 1000,
			Url:       ncContest.Link,
		})
	}
	return
}
