package notifyService

import (
	"context"
	"fmt"
	"tgwp/global"
	"tgwp/internal/utils/messageService"
	"tgwp/log/zlog"
	"tgwp/repo"
	"tgwp/utils/email"
)

// NotifyType 通知类型
type NotifyType string

const (
	NotifyTypeSystemMessage NotifyType = "system_message" // 系统消息
	NotifyTypeLike          NotifyType = "like"           // 点赞通知
	NotifyTypeReply         NotifyType = "reply"          // 回复通知
	NotifyTypeHelpPost      NotifyType = "help_post"      // 求助帖通知
)

// NotifyData 通知数据
type NotifyData struct {
	Type     NotifyType // 通知类型
	UserID   int64      // 接收者ID
	SenderID int64      // 发送者ID（系统消息为0）
	Content  string     // 通知内容
	URL      string     // 跳转链接
}

// SendNotify 发送通知（站内消息 + 邮件通知）
func SendNotify(ctx context.Context, data NotifyData) {
	// 1. 发送站内消息（始终发送）
	sendInternalMessage(data)

	// 2. 检查用户设置，决定是否发送邮件
	go checkAndSendEmail(ctx, data)
}

// sendInternalMessage 发送站内消息
func sendInternalMessage(data NotifyData) {
	switch data.Type {
	case NotifyTypeSystemMessage:
		messageService.SendSystemMessage(data.UserID, data.Content, data.URL)
	case NotifyTypeLike:
		messageService.SendLikeMessageIfNotSelf(data.UserID, data.SenderID, data.Content, data.URL)
	case NotifyTypeReply:
		messageService.SendCommentMessageIfNotSelf(data.UserID, data.SenderID, data.Content, data.URL)
	case NotifyTypeHelpPost:
		messageService.SendSystemMessage(data.UserID, data.Content, data.URL)
	}
}

// checkAndSendEmail 检查用户设置并发送邮件
func checkAndSendEmail(ctx context.Context, data NotifyData) {
	// 获取用户设置
	settingRepo := repo.NewSettingRepo(global.DB)
	setting, err := settingRepo.GetOrCreate(data.UserID)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取用户设置失败，用户ID: %d, 错误: %v", data.UserID, err)
		return
	}

	// 根据通知类型检查是否需要发送邮件
	shouldSendEmail := false
	switch data.Type {
	case NotifyTypeSystemMessage:
		shouldSendEmail = setting.Settings.SystemMessageEmailNotify
	case NotifyTypeLike:
		shouldSendEmail = setting.Settings.LikeNotify
	case NotifyTypeReply:
		shouldSendEmail = setting.Settings.ReplyNotify
	case NotifyTypeHelpPost:
		shouldSendEmail = setting.HelpPostNotify // 使用冗余字段
	}

	if !shouldSendEmail {
		zlog.CtxDebugf(ctx, "用户 %d 未开启 %s 邮件通知", data.UserID, data.Type)
		return
	}

	// 获取用户邮箱
	userRepo := repo.NewUserRepo(global.DB)
	user, err := userRepo.GetUserProfileByID(data.UserID)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取用户信息失败，用户ID: %d, 错误: %v", data.UserID, err)
		return
	}

	if user.Email == "" {
		zlog.CtxWarnf(ctx, "用户 %d 未设置邮箱，无法发送邮件通知", data.UserID)
		return
	}

	// 发送邮件
	subject := getEmailSubject(data.Type)
	emailContent := formatEmailContent(data.Content, data.URL)
	err = email.Send([]string{user.Email}, subject, emailContent)
	if err != nil {
		zlog.CtxErrorf(ctx, "发送邮件失败，用户ID: %d, 邮箱: %s, 错误: %v", data.UserID, user.Email, err)
		return
	}

	zlog.CtxInfof(ctx, "邮件通知发送成功，用户ID: %d, 邮箱: %s, 类型: %s", data.UserID, user.Email, data.Type)
}

// getEmailSubject 获取邮件主题
func getEmailSubject(notifyType NotifyType) string {
	switch notifyType {
	case NotifyTypeSystemMessage:
		return "[AcKing学习分享平台] [系统消息]"
	case NotifyTypeLike:
		return "[AcKing学习分享平台] [点赞通知]"
	case NotifyTypeReply:
		return "[AcKing学习分享平台] [回复通知]"
	case NotifyTypeHelpPost:
		return "[AcKing学习分享平台] [求助帖通知]"
	default:
		return "[AcKing学习分享平台] [通知]"
	}
}

// formatEmailContent 格式化邮件内容
func formatEmailContent(content string, url string) string {
	baseURL := "http://120.79.250.47" // 从配置中读取
	fullURL := baseURL + url

	html := `
	<div style="font-family: Arial, sans-serif; padding: 20px; background-color: #f5f5f5;">
		<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
			<h2 style="color: #333; margin-bottom: 20px;">通知详情</h2>
			<div style="background-color: #f8f9fa; padding: 15px; border-radius: 4px; margin-bottom: 20px;">
				<p style="color: #555; line-height: 1.6; margin: 0;">%s</p>
			</div>
			<div style="margin-top: 20px;">
				<a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px;">查看详情</a>
			</div>
			<hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
			<p style="color: #999; font-size: 12px; text-align: center;">
				如不想接收此类邮件，请在<a href="%s/settings" style="color: #007bff;">个人设置</a>中关闭邮件通知
			</p>
		</div>
	</div>
	`
	return fmt.Sprintf(html, content, fullURL, baseURL)
}

// SendSystemMessageNotify 发送系统消息通知
func SendSystemMessageNotify(ctx context.Context, userID int64, content, url string) {
	SendNotify(ctx, NotifyData{
		Type:     NotifyTypeSystemMessage,
		UserID:   userID,
		SenderID: 0,
		Content:  content,
		URL:      url,
	})
}

// SendLikeNotify 发送点赞通知
func SendLikeNotify(ctx context.Context, userID, senderID int64, content, url string) {
	if userID == senderID {
		return // 不给自己发通知
	}
	SendNotify(ctx, NotifyData{
		Type:     NotifyTypeLike,
		UserID:   userID,
		SenderID: senderID,
		Content:  content,
		URL:      url,
	})
}

// SendReplyNotify 发送回复通知
func SendReplyNotify(ctx context.Context, userID, senderID int64, content, url string) {
	if userID == senderID {
		return // 不给自己发通知
	}
	SendNotify(ctx, NotifyData{
		Type:     NotifyTypeReply,
		UserID:   userID,
		SenderID: senderID,
		Content:  content,
		URL:      url,
	})
}

// SendHelpPostNotify 批量发送求助帖通知
func SendHelpPostNotify(ctx context.Context, postTitle, postURL string, excludeUserID int64) {
	// 获取所有开启了求助帖通知的用户ID
	settingRepo := repo.NewSettingRepo(global.DB)
	userIDs, err := settingRepo.GetUserIDsWithHelpPostNotify()
	if err != nil {
		zlog.CtxErrorf(ctx, "获取开启求助帖通知的用户列表失败: %v", err)
		return
	}

	content := fmt.Sprintf("有新的求助帖：《%s》", postTitle)

	// 异步批量发送通知
	for _, userID := range userIDs {
		// 跳过发帖人自己
		if userID == excludeUserID {
			continue
		}

		// 发送通知
		go SendNotify(ctx, NotifyData{
			Type:     NotifyTypeHelpPost,
			UserID:   userID,
			SenderID: excludeUserID,
			Content:  content,
			URL:      postURL,
		})
	}

	zlog.CtxInfof(ctx, "批量发送求助帖通知完成，通知用户数: %d", len(userIDs))
}
