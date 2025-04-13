package logic

import (
	"context"
	"tgwp/types"
	"tgwp/utils"
	"time"
)

type TemplateLogic struct {
}

func NewTemplateLogic() *TemplateLogic {
	return &TemplateLogic{}
}

// 这个包内的常量
const (
	REDIS_SNOW_ID = "island:test.code:string"
)

func (l *TemplateLogic) Way(ctx context.Context, req types.TemplateReq) (resp types.TemplateResp, err error) {
	defer utils.RecordTime(time.Now())()

	return
}
