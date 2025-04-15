package logic

import (
	"context"
	"strconv"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"
)

type MessageLogic struct {
}

func NewMessageLogic() *MessageLogic {
	return &MessageLogic{}
}

func (l *MessageLogic) GetMessageCount(ctx context.Context, req types.GetMessageCountReq) (resp types.GetMessageCountResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 获取消息数量
	resp.SystemCount, resp.LikeCount, resp.CommentCount, err = repo.NewMessageRepo(global.DB).GetMessageCount(userID)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取消息数量失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.Count = resp.SystemCount + resp.LikeCount + resp.CommentCount
	return resp, nil
}

func (l *MessageLogic) GetMessageList(ctx context.Context, req types.GetMessageListReq) (resp types.GetMessageListResp, err error) {
	defer utils.RecordTime(time.Now())()
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 获取消息列表
	var MessageList []model.Message
	MessageList, err = repo.NewMessageRepo(global.DB).GetMessageList(userID, req.Type)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取消息列表失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 转换为 MessageResp
	for i := range MessageList {
		resp.Messages = append(resp.Messages, types.Message{
			ID:        MessageList[i].ID,
			Type:      MessageList[i].Type,
			Content:   MessageList[i].Content,
			CreatedAt: MessageList[i].CreatedTime,
			SenderID:  MessageList[i].SenderID,
			Url:       MessageList[i].Url,
			IsRead:    MessageList[i].IsRead,
		})
	}
	resp.Length = len(MessageList)
	return
}

func (l *MessageLogic) MarkReadMessage(ctx context.Context, req types.MarkReadMessageReq) (err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	messageID, err := strconv.ParseInt(req.MessageID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.MessageID, err)
		return response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 标记已读
	err = repo.NewMessageRepo(global.DB).MarkReadMessage(userID, messageID)
	if err != nil {
		zlog.CtxErrorf(ctx, "标记已读失败: %v", err)
		return response.ErrResp(err, response.DATABASE_ERROR)
	}
	return nil
}
