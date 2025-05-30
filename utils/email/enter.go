package email

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"tgwp/global"
	"tgwp/log/zlog"
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
