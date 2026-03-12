package service

import (
	"bytes"
	"fmt"
	"html/template"
	"mime"
	"net/mail"
	"os"
	"path/filepath"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/pkg/email"
	"regexp"
	"strings"
	"time"
)

type VerificationEmailData struct {
	SiteName  string
	Username  string
	VerifyUrl string
}

type TestEmailData struct {
	SiteName string
	Time     string
}

type EmailChangeData struct {
	SiteName  string
	Username  string
	OldEmail  string
	NewEmail  string
	VerifyUrl string
}

type PasswordResetData struct {
	SiteName string
	Username string
	ResetUrl string
}

var strictEmailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z0-9]+`)

func (s *EmailService) EmailEnabled() bool {
	if !s.dbConfig.GetBool(consts.ConfigEnableSMTP) {
		return false
	}
	cfg := s.staticConfig
	host := strings.TrimSpace(cfg.SMTP.Host)
	if host == "" {
		return false
	}
	from := strings.TrimSpace(cfg.SMTP.From)
	if from == "" {
		return false
	}
	if _, err := mail.ParseAddress(from); err != nil {
		return false
	}
	return true
}

func (s *EmailService) ShouldSendRegistrationVerificationEmail() bool {
	return s.EmailEnabled() && s.dbConfig.GetBool(consts.ConfigSendRegistrationVerificationEmail)
}

// SendVerificationEmail 发送验证邮件
func (s *EmailService) SendVerificationEmail(toEmail, username, verifyUrl string) error {
	// 检查是否开启 SMTP
	if !s.dbConfig.GetBool(consts.ConfigEnableSMTP) {
		return fmt.Errorf("请先开启SMTP功能")
	}

	cfg := s.staticConfig
	if cfg.SMTP.Host == "" {
		return fmt.Errorf("请设置SMTP服务器地址")
	}

	siteName := s.dbConfig.GetString(consts.ConfigSiteName)
	if siteName == "" {
		siteName = "Perfect Pic"
	}

	// 邮件主题
	subject := fmt.Sprintf("欢迎注册 %s - 请验证您的邮箱", siteName)

	// 读取模板文件
	templatePath := filepath.Join(config.GetConfigDir(), "verification-mail.html")
	contentBytes, err := os.ReadFile(templatePath)
	var bodyTpl string
	if err != nil {
		// 如果模板读取失败，使用默认简单模板
		bodyTpl = `
			<h1>欢迎加入 {{.SiteName}}</h1>
			<p>请点击链接验证邮箱: <a href="{{.VerifyUrl}}">{{.VerifyUrl}}</a></p>
		`
	} else {
		bodyTpl = string(contentBytes)
	}

	data := VerificationEmailData{
		SiteName:  siteName,
		Username:  username,
		VerifyUrl: verifyUrl,
	}

	body, err := renderTemplate(bodyTpl, data)
	if err != nil {
		return err
	}

	_, fromAddr, err := formatAddressHeader(cfg.SMTP.From)
	if err != nil {
		return err
	}
	_, toAddr, err := formatAddressHeader(toEmail)
	if err != nil {
		return err
	}

	smtpConfig := email.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		SSL:      cfg.SMTP.SSL,
	}
	emailInfo := email.Email{
		From:    fromAddr,
		To:      []string{toAddr},
		Subject: subject,
		Body:    body,
	}
	return s.mailer.SendWithSMTP(smtpConfig, emailInfo)
}

// SendTestEmail 发送测试邮件
func (s *EmailService) SendTestEmail(toEmail string) error {
	if !s.dbConfig.GetBool(consts.ConfigEnableSMTP) {
		return fmt.Errorf("请先开启SMTP功能")
	}

	cfg := s.staticConfig
	if cfg.SMTP.Host == "" {
		return fmt.Errorf("请设置SMTP服务器地址")
	}

	siteName := s.dbConfig.GetString(consts.ConfigSiteName)
	if siteName == "" {
		siteName = "Perfect Pic"
	}

	subject := fmt.Sprintf("%s SMTP 测试邮件", siteName)
	bodyTpl := `
<!DOCTYPE html>
<html>
<body>
    <h3>SMTP 配置测试成功</h3>
    <p>这是一封来自 <strong>{{.SiteName}}</strong> 的测试邮件。</p>
    <p>如果您收到此邮件，说明您的 SMTP 服务配置正确。</p>
    <p>时间: {{.Time}}</p>
</body>
</html>
`
	data := TestEmailData{
		SiteName: siteName,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
	}

	body, err := renderTemplate(bodyTpl, data)
	if err != nil {
		return err
	}

	_, fromAddr, err := formatAddressHeader(cfg.SMTP.From)
	if err != nil {
		return err
	}
	_, toAddr, err := formatAddressHeader(toEmail)
	if err != nil {
		return err
	}

	smtpConfig := email.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		SSL:      cfg.SMTP.SSL,
	}
	emailInfo := email.Email{
		From:    fromAddr,
		To:      []string{toAddr},
		Subject: subject,
		Body:    body,
	}
	return s.mailer.SendWithSMTP(smtpConfig, emailInfo)
}

// SendEmailChangeVerification 发送修改邮箱验证邮件
func (s *EmailService) SendEmailChangeVerification(toEmail, username, oldEmail, newEmail, verifyUrl string) error {
	if !s.dbConfig.GetBool(consts.ConfigEnableSMTP) {
		return fmt.Errorf("请先开启SMTP功能")
	}

	cfg := s.staticConfig
	if cfg.SMTP.Host == "" {
		return fmt.Errorf("请设置SMTP服务器地址")
	}

	siteName := s.dbConfig.GetString(consts.ConfigSiteName)
	if siteName == "" {
		siteName = "Perfect Pic"
	}

	// 邮件主题
	subject := fmt.Sprintf("%s - 请确认修改邮箱", siteName)

	// 读取模板文件
	templatePath := filepath.Join(config.GetConfigDir(), "email-change-mail.html")
	contentBytes, err := os.ReadFile(templatePath)
	var bodyTpl string
	if err != nil {
		bodyTpl = `
			<h1>修改邮箱确认 - {{.SiteName}}</h1>
			<p>您请求将邮箱从 {{.OldEmail}} 修改为 {{.NewEmail}}。</p>
			<p>请点击链接确认: <a href="{{.VerifyUrl}}">{{.VerifyUrl}}</a></p>
		`
	} else {
		bodyTpl = string(contentBytes)
	}

	data := EmailChangeData{
		SiteName:  siteName,
		Username:  username,
		OldEmail:  oldEmail,
		NewEmail:  newEmail,
		VerifyUrl: verifyUrl,
	}

	body, err := renderTemplate(bodyTpl, data)
	if err != nil {
		return err
	}

	_, fromAddr, err := formatAddressHeader(cfg.SMTP.From)
	if err != nil {
		return err
	}
	_, toAddr, err := formatAddressHeader(toEmail)
	if err != nil {
		return err
	}

	smtpConfig := email.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		SSL:      cfg.SMTP.SSL,
	}
	emailInfo := email.Email{
		From:    fromAddr,
		To:      []string{toAddr},
		Subject: subject,
		Body:    body,
	}
	return s.mailer.SendWithSMTP(smtpConfig, emailInfo)
}

// SendPasswordResetEmail 发送重置密码邮件
func (s *EmailService) SendPasswordResetEmail(toEmail, username, resetUrl string) error {
	if !s.dbConfig.GetBool(consts.ConfigEnableSMTP) {
		return fmt.Errorf("请先开启SMTP功能")
	}

	cfg := s.staticConfig
	if cfg.SMTP.Host == "" {
		return fmt.Errorf("请设置SMTP服务器地址")
	}

	siteName := s.dbConfig.GetString(consts.ConfigSiteName)
	if siteName == "" {
		siteName = "Perfect Pic"
	}

	// 邮件主题
	subject := fmt.Sprintf("%s - 重置密码请求", siteName)

	// 读取模板文件
	templatePath := filepath.Join(config.GetConfigDir(), "reset-password-mail.html")
	contentBytes, err := os.ReadFile(templatePath)
	var bodyTpl string
	if err != nil {
		bodyTpl = `
			<h1>重置密码 - {{.SiteName}}</h1>
			<p>请点击链接重置密码: <a href="{{.ResetUrl}}">{{.ResetUrl}}</a></p>
			<p>有效期15分钟。</p>
		`
	} else {
		bodyTpl = string(contentBytes)
	}

	data := PasswordResetData{
		SiteName: siteName,
		Username: username,
		ResetUrl: resetUrl,
	}

	body, err := renderTemplate(bodyTpl, data)
	if err != nil {
		return err
	}

	_, fromAddr, err := formatAddressHeader(cfg.SMTP.From)
	if err != nil {
		return err
	}
	_, toAddr, err := formatAddressHeader(toEmail)
	if err != nil {
		return err
	}

	smtpConfig := email.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		SSL:      cfg.SMTP.SSL,
	}
	emailInfo := email.Email{
		From:    fromAddr,
		To:      []string{toAddr},
		Subject: subject,
		Body:    body,
	}
	return s.mailer.SendWithSMTP(smtpConfig, emailInfo)
}

func renderTemplate(tpl string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// formatAddressHeader 规范化邮箱头部地址并返回 header 展示值与裸邮箱地址。
func formatAddressHeader(input string) (string, string, error) {
	// 解析地址 (如果格式不对，这里直接报错，起到 Validation 作用)
	addr, err := mail.ParseAddress(input)
	if err != nil {
		return "", "", err
	}
	cleanAddr := strictEmailRegex.FindString(addr.Address)
	if cleanAddr == "" {
		return "", "", fmt.Errorf("invalid email format detected")
	}

	var finalHeader string
	if addr.Name != "" {
		cleanName := strings.ReplaceAll(addr.Name, "\r", "")
		cleanName = strings.ReplaceAll(cleanName, "\n", "")
		encodedName := mime.BEncoding.Encode("UTF-8", cleanName)
		finalHeader = fmt.Sprintf("%s <%s>", encodedName, cleanAddr)
	} else {
		// 没有名字，只返回清洗后的地址
		finalHeader = cleanAddr
	}

	return finalHeader, cleanAddr, nil
}
