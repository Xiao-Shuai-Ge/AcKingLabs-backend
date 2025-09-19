package email

import (
	"crypto/tls"
	"fmt"
	"tgwp/global"
	"tgwp/log/zlog"

	"gopkg.in/gomail.v2"
)

// Send 发送邮件
func Send(to []string, subject string, message string) error {
	// 1. 连接SMTP服务器
	host := global.Config.Email.Host
	port := global.Config.Email.Port
	userName := global.Config.Email.UserName
	password := global.Config.Email.Password

	// 2. 构建邮件对象
	m := gomail.NewMessage()
	m.SetHeader("From", userName)   // 发件人
	m.SetHeader("To", to...)        // 收件人
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", message) // 正文

	d := gomail.NewDialer(
		host,
		port,
		userName,
		password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		zlog.Errorf("邮件发送失败：%v", err)
		return err
	}
	return nil
}

// SendWithImage 发送带图片的邮件
func SendWithImage(to []string, subject string, message string, imagePath string) error {
	// 1. 连接SMTP服务器
	host := global.Config.Email.Host
	port := global.Config.Email.Port
	userName := global.Config.Email.UserName
	password := global.Config.Email.Password

	// 2. 构建邮件对象
	m := gomail.NewMessage()
	m.SetHeader("From", userName)   // 发件人
	m.SetHeader("To", to...)        // 收件人
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", message) // 正文
	m.Embed(imagePath)

	d := gomail.NewDialer(
		host,
		port,
		userName,
		password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		zlog.Errorf("邮件发送失败：%v", err)
		return err
	}
	return nil
}

// SendCode 发送验证码
func SendCode(to string, code int64) error {
	message := `
	<p style="text-indent:2em;">你的邮箱验证码为: %06d </p> 
	<p style="text-indent:2em;">此验证码的有效期为5分钟，请尽快使用。</p>
	`
	return Send([]string{to}, "[AcKing学习分享平台] [邮箱验证码]", fmt.Sprintf(message, code))
}

func BookingContest(to []string, url string, title string, time string) error {
	message := `
<div>
    <span>你订阅的比赛</span>
    <a href="%s" style="margin: 2px;">%s</a>
    <span>将在 %s 分钟后开始，请注意准备。</span>
</div>
	`
	return Send(to, "[AcKing学习分享平台] [比赛预约]", fmt.Sprintf(message, url, title, time))
}

func AutoSignin(to string, name string) error {
	message := `
<div>
    <span>检测到你的签到： %s ，已为你自动签到成功。请注意。</span>
</div>
	`
	return Send([]string{to}, "[AcKing学习分享平台] [自动签到]", fmt.Sprintf(message, name))
}

// SendInvitationCodeEmail 发送邀请码邮件
func SendInvitationCodeEmail(to string, code string) error {
	message := `
	<div>
		<p style="text-indent:2em;">恭喜！您的简历已通过审核。</p>
		<p style="text-indent:2em;">您的邀请码为: <strong style="color: #007bff; font-size: 18px;">%s</strong></p>
		<p style="text-indent:2em;">请使用此邀请码注册账号，邀请码仅限该邮箱使用。</p>
		<br>
		<p style="text-indent:2em;">请扫描下方二维码加入群聊：</p>
		<img src="cid:qr-code.png" alt="群聊二维码" style="width: 100px; height: 100px; display: block; margin: 0 auto;">
		<br>
		<p style="text-indent:2em;">如有疑问，请联系管理员。</p>
	</div>
	`
	return SendWithImage([]string{to}, "[AcKing学习分享平台] [简历通过通知]", fmt.Sprintf(message, code), "static/images/qr-code.png")
}

// SendRejectionEmail 发送简历不通过通知邮件
func SendRejectionEmail(to string) error {
	message := `
	<div>
		<p style="text-indent:2em;">经过实验室评审团队的综合评估，很遗憾地通知您，您的简历未通过审核。</p>
		<p style="text-indent:2em;">非常感谢你对AcKing算法竞赛实验室的关注与认可，以及在面试过程中展现出的热情和准备🌹🌹🌹</p>
		<p style="text-indent:2em;">欢迎您继续关注我们的其他活动，祝你在算法学习之路上收获更多进步！👍🏻👍🏻👍🏻</p>
		<br>
		<p style="text-indent:2em;">如有疑问，请联系管理员。</p>
	</div>
	`
	return Send([]string{to}, "[AcKing学习分享平台] [简历审核结果通知]", message)
}

// SendPendingResumeEmail 发送待考核通知邮件
func SendPendingResumeEmail(to string) error {
	message := `
	<div>
		<p style="text-indent:2em;">恭喜！您的简历已通过审核！等待进入下一轮考核！</p>
		<br>
		<p style="text-indent:2em;">请尽快扫描下方二维码加入群聊，等待考核通知：</p>
		<img src="cid:qr-code.png" alt="群聊二维码" style="width: 100px; height: 100px; display: block; margin: 0 auto;">
		<br>
		<p style="text-indent:2em;">如有疑问，请联系管理员。</p>
	</div>
	`
	return SendWithImage([]string{to}, "[AcKing学习分享平台] [简历待考核通知]", message, "static/images/qr-code.png")
}
