package list

import (
	"fmt"
	"gorm.io/gorm"
	"tgwp/global"
)

/*                             通用的列表查询                              */
type PageInfo struct {
	Limit int    `form:"limit"`
	Page  int    `form:"page"`
	Key   string `form:"key"` //分页关键词
	Order string `form:"order"`
}

func (p *PageInfo) GetPage() int {
	if p.Page > 20 || p.Page <= 0 {
		p.Page = 1
	}
	return p.Page
}
func (p *PageInfo) GetLimit() int {
	if p.Limit > 100 || p.Limit <= 0 {
		p.Limit = 10
	}
	return p.Limit
}
func (p *PageInfo) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

type Options struct {
	PageInfo PageInfo
	Likes    []string //模糊匹配
	Preloads []string //预加载
	Where    *gorm.DB //自定义条件
	Debug    bool     //是否开启debug
	Order    string   //排序
}

func ListQuery[T any](model T, option Options) (list []T, count int, err error) {
	// 自己的基础查询
	query := global.DB.Model(&model).Where(model)
	//显示日志
	if option.Debug {
		query = query.Debug()
	}
	//预加载
	if len(option.Preloads) > 0 {
		for _, v := range option.Preloads {
			query = query.Preload(v)
		}
	}
	//模糊匹配
	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		likes := global.DB.Where("")
		for _, v := range option.Likes {
			likes.Or(fmt.Sprintf("%s like ?", v), fmt.Sprintf("%%%s%%", option.PageInfo.Key))
		}
		query = query.Where(likes)
	}
	//定制化查询
	if option.Where != nil {
		query = query.Where(option.Where)
	}
	//查总数,这里的总数会经过前面的处理
	var cnt int64
	query.Count(&cnt)
	count = int(cnt)
	//分页查询
	limit := option.PageInfo.GetLimit()
	offset := option.PageInfo.GetOffset()
	//排序
	if option.PageInfo.Order != "" {
		//前端定义的排序规则
		query = query.Order(option.PageInfo.Order)
	} else {
		//自定义的排序规则
		if option.Order != "" {
			query = query.Order(option.Order)
		}
	}
	err = query.Offset(offset).Limit(limit).Find(&list).Error
	return
}

// 用于绑定的id列表
type RemoveReq struct {
	Ids []int `json:"ids"`
}
