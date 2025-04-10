package logic

import (
	"context"
	"fmt"
	"strconv"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"
	"unicode/utf8"
)

type PostLogic struct {
}

func NewPostLogic() *PostLogic {
	return &PostLogic{}
}

// CreatePost 创建帖子
func (l *PostLogic) CreatePost(ctx context.Context, req types.CreatePostReq) (resp types.CreatePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 如果是周记打卡，先检查时间是否正确
	if req.Type == "diary" {
		req.Source = GetWeekCode()
		if len(req.Source) == 0 {
			zlog.CtxErrorf(ctx, "周记打卡时间错误: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		zlog.CtxInfof(ctx, "解析出打卡周数: %s", req.Source)
	}
	// 判断数据范围
	// 1. 标题不能超过 50 个字符
	if len(req.Title) > 50 {
		zlog.CtxErrorf(ctx, "标题不能超过 50 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2. 内容不能超过 5000 个字符
	zlog.CtxInfof(ctx, "内容长度: %d", utf8.RuneCountInString(req.Content))
	if utf8.RuneCountInString(req.Content) > 5000 {
		zlog.CtxErrorf(ctx, "内容不能超过 5000 个字: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 3. 除了周记打卡可以私密，其他类型都不可以私密
	if req.Type != "diary" && req.IsPrivate {
		zlog.CtxErrorf(ctx, "非周记打卡不能私密: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 4. 不允许出现不存在的类型
	if !global.TYPE_SET[req.Type] {
		zlog.CtxErrorf(ctx, "不存在的类型: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 创建帖子
	id := global.SnowflakeNode.Generate().Int64()
	post := model.Post{
		ID:        id,
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		Source:    req.Source,
		IsPrivate: req.IsPrivate,
	}
	err = repo.NewPostRepo(global.DB).CreatePost(post)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.ID = id
	return
}

func (l *PostLogic) GetPostDetail(ctx context.Context, req types.GetPostDetailReq) (resp types.GetPostDetailResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询帖子详情
	post, err := repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 判断是否有权限查看
	if post.IsPrivate {
		// 如果是私密，只有自己和管理员可以查看
		if post.UserID != postID && !(req.OperatorRole >= global.ROLE_ADMIN) {
			zlog.CtxErrorf(ctx, "非作者或管理员无权查看私密帖子: %v", err)
			return resp, response.ErrResp(err, response.PERMISSION_DENIED)
		}
	}
	// 转换为响应结构
	resp.ID = post.ID
	resp.UserID = post.UserID
	resp.Title = post.Title
	resp.Content = post.Content
	resp.Type = post.Type
	resp.Source = post.Source
	resp.Likes = post.Likes
	resp.Comments = post.Comments
	resp.CreatedAt = post.CreatedTime
	resp.UpdatedAt = post.UpdatedTime

	resp.IsAdminLike = post.IsAdminLike
	resp.IsPrivate = post.IsPrivate
	resp.IsFeatured = post.IsFeatured
	return
}

func (l *PostLogic) LikePost(ctx context.Context, req types.LikePostReq) (resp types.LikePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断是否已经点赞
	isLike, err := repo.NewPostRepo(global.DB).IsPostLikeExists(postID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	if isLike {
		// 已经点赞，取消点赞
		err = repo.NewPostRepo(global.DB).CancelPostLike(postID, operatorID)
		if err != nil {
			zlog.CtxErrorf(ctx, "取消点赞失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		resp.IsLike = false
		return
	} else {
		// 先判断是否为管理员点赞
		if req.OperatorRole >= global.ROLE_ADMIN {
			// 判断是否已有管理员点赞，如果第一次有管理员点赞，应该为用户增加经验
			var post model.Post
			post, err = repo.NewPostRepo(global.DB).GetPostDetail(postID)
			if err != nil {
				zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
				return resp, response.ErrResp(err, response.DATABASE_ERROR)
			}
			if !post.IsAdminLike {
				// 第一次有管理员点赞，增加经验 +5，并标记为管理员点赞
				err = repo.NewPostRepo(global.DB).MarkAdminLikePost(postID)
				if err != nil {
					zlog.CtxErrorf(ctx, "标记管理员点赞失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
				// 增加经验
				err = repo.NewUserRepo(global.DB).AddUserXp(post.UserID, 5)
				if err != nil {
					zlog.CtxErrorf(ctx, "增加经验失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
			}
		}
		// 点赞
		id := global.SnowflakeNode.Generate().Int64()
		postLike := model.PostLike{
			PostID: postID,
			UserID: operatorID,
			ID:     id,
		}
		err = repo.NewPostRepo(global.DB).AddPostLike(postLike)
		if err != nil {
			zlog.CtxErrorf(ctx, "点赞失败: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		resp.IsLike = true
	}
	return
}

func (l *PostLogic) GetLikePost(ctx context.Context, req types.GetLikePostReq) (resp types.GetLikePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询是否有点赞记录
	IsLike, err := repo.NewPostRepo(global.DB).IsPostLikeExists(postID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	resp.IsLike = IsLike
	return
}

func (l *PostLogic) CreateComment(ctx context.Context, req types.CreateCommentReq) (resp types.CreateCommentResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断内容长度
	if utf8.RuneCountInString(req.Content) > 1000 {
		zlog.CtxErrorf(ctx, "评论内容不能超过 1000 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 创建评论
	id := global.SnowflakeNode.Generate().Int64()
	comment := model.Comment{
		ID:      id,
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
		Likes:   0,
	}
	err = repo.NewPostRepo(global.DB).CreateComment(comment)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建评论失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.ID = id
	return
}

func (l *PostLogic) GetMoreComments(ctx context.Context, req types.GetMoreCommentsReq) (resp types.GetMoreCommentsResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	beforeID, err := strconv.ParseInt(req.BeforeID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.BeforeID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 从数据库中查询评论
	comments, err := repo.NewPostRepo(global.DB).GetMoreComments(postID, beforeID, req.Count)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询评论失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	zlog.CtxDebugf(ctx, "查询评论成功: %v", comments)
	for _, comment := range comments {
		resp.Comments = append(resp.Comments, types.Comment{
			ID:          comment.ID,
			UserID:      comment.UserID,
			Content:     comment.Content,
			Likes:       comment.Likes,
			CreatedAt:   comment.CreatedTime,
			IsAdminLike: comment.IsAdminLike,
		})
	}
	resp.Length = len(resp.Comments)
	return
}

func (l *PostLogic) LikeComment(ctx context.Context, req types.LikeCommentReq) (resp types.LikeCommentResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	commentID, err := strconv.ParseInt(req.CommentID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.CommentID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	err = repo.NewPostRepo(global.DB).CancelCommentLike(commentID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "hhhh取消点赞失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	return
	// 判断是否已经点赞
	isLike, err := repo.NewPostRepo(global.DB).IsCommentLikeExists(commentID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	if isLike {
		zlog.CtxDebugf(ctx, "取消点赞")
		err = repo.NewPostRepo(global.DB).CancelCommentLike(commentID, operatorID)
		if err != nil {
			zlog.CtxErrorf(ctx, "取消点赞失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		resp.IsLike = false
		return
	} else {
		// 先判断是否为管理员点赞
		if req.OperatorRole >= global.ROLE_ADMIN {
			// 判断是否已有管理员点赞，如果第一次有管理员点赞，应该为用户增加经验
			var comment model.Comment
			comment, err = repo.NewPostRepo(global.DB).GetCommentDetail(commentID)
			if err != nil {
				zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
				return resp, response.ErrResp(err, response.DATABASE_ERROR)
			}
			if !comment.IsAdminLike {
				// 第一次有管理员点赞，增加经验 +2，并标记为管理员点赞
				err = repo.NewPostRepo(global.DB).MarkAdminLikeComment(commentID)
				if err != nil {
					zlog.CtxErrorf(ctx, "标记管理员点赞失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
				// 增加经验
				err = repo.NewUserRepo(global.DB).AddUserXp(comment.UserID, 2)
				if err != nil {
					zlog.CtxErrorf(ctx, "增加经验失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
			}
		}
		// 点赞
		id := global.SnowflakeNode.Generate().Int64()
		commentLike := model.CommentLike{
			CommentID: commentID,
			UserID:    operatorID,
			ID:        id,
		}
		err = repo.NewPostRepo(global.DB).AddCommentLike(commentLike)
		if err != nil {
			zlog.CtxErrorf(ctx, "点赞失败: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		resp.IsLike = true
	}
	return
}

func (l *PostLogic) GetLikeComment(ctx context.Context, req types.GetLikeCommentReq) (resp types.GetLikeCommentResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	commentID, err := strconv.ParseInt(req.CommentID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.CommentID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询是否有点赞记录
	IsLike, err := repo.NewPostRepo(global.DB).IsCommentLikeExists(commentID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	resp.IsLike = IsLike
	return
}

func GetWeekCode() string {
	timeNow := time.Now()
	timeNow = time.UnixMilli(1744081921000)
	timestamp := timeNow.UnixMilli()

	// 打卡时间为每周的周日中午到周二的中午，为了先确定当前周数，先把时间减去 2 天
	timestamp -= 2 * 24 * 60 * 60 * 1000
	// 周一到周日分别为 1 到 7
	weekday := int64((time.UnixMilli(timestamp).Weekday()+6)%7 + 1)

	// 时间返回到周一中午 12:00:00
	timestamp -= (weekday - 1) * 24 * 60 * 60 * 1000
	timestamp -= (timestamp + 8*3600*1000) % (24 * 60 * 60 * 1000) // 取整到天
	timestamp += 12 * 60 * 60 * 1000                               // 加上中午(UTC+8)
	// 如果当前时间不在合法打卡
	if utils.Abs(timeNow.UnixMilli()-(timestamp+7*24*60*60*1000)) > 24*60*60*1000 {
		zlog.Debugf("%v = %v", utils.Abs(timeNow.UnixMilli()-(timestamp+7*24*60*60*1000)), 24*60*60*1000)
		zlog.Warnf("当前时间不在合法打卡时间范围内 %v ~ %v", time.UnixMilli(timestamp+7*24*60*60*1000), timeNow)
		return ""
	}
	// 计算当前周数，确定年份和月份
	week := 0
	year := time.UnixMilli(timestamp).Year()
	month := time.UnixMilli(timestamp).Month()
	for month == time.UnixMilli(timestamp).Month() {
		timestamp -= 7 * 24 * 60 * 60 * 1000
		week++
	}
	// 格式化周数
	weekCode := fmt.Sprintf("%d-%d-%d", year, month, week)
	return weekCode
}
