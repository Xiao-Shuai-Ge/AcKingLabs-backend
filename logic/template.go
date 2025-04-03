package logic

import (
	"context"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"tgwp/utils/snowflake"
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
	//雪花id的生成格式
	node, err := snowflake.NewNode(global.DEFAULT_NODE_ID)
	if err != nil {
		zlog.CtxErrorf(ctx, "NewNode err: %v", err)
		return resp, response.ErrResp(err, response.COMMON_FAIL)
	}
	//一般是生成12位的int64id，也可以生成string的，看snowflakes包
	id := snowflake.GetInt12Id(node)
	//logic使用repo层，调用mysql
	err = repo.NewTemplateRepo(global.DB).InsertData(id)
	if err != nil {
		zlog.CtxErrorf(ctx, "InsertData err: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	//logic使用redis
	//将id放入redis1分钟
	err = global.Rdb.Set(ctx, REDIS_SNOW_ID, id, time.Second*60).Err()
	if err != nil {
		zlog.CtxErrorf(ctx, "Store the verification code err: %v", err)
		return resp, response.ErrResp(err, response.COMMON_FAIL)
	}
	return types.TemplateResp{
		Body: req.Body,
	}, nil
}
