package alert

import (
	"fmt"
)

// AlertClient 告警客户端接口
type AlertClient interface {
	// SendAlert 发送告警消息
	SendAlert(level, title, content string) error

	// SendText 发送文本告警
	SendText(content string, mentionedList, mentionedMobileList []string) error

	// SendMarkdown 发送Markdown格式告警
	SendMarkdown(content string) error

	// SendMarkdownV2 发送MarkdownV2格式告警
	SendMarkdownV2(content string) error
}

// AlertLevel 告警级别
type AlertLevel string

const (
	// AlertLevelEmergency 紧急告警
	AlertLevelEmergency AlertLevel = "emergency"
	// AlertLevelCritical 严重告警
	AlertLevelCritical AlertLevel = "critical"
	// AlertLevelWarning 警告告警
	AlertLevelWarning AlertLevel = "warning"
	// AlertLevelInfo 信息告警
	AlertLevelInfo AlertLevel = "info"
	// AlertLevelDebug 调试告警
	AlertLevelDebug AlertLevel = "debug"
)

// NewAlertClient 创建告警客户端
func NewAlertClient(opts ...Option) (AlertClient, error) {
	// 创建默认选项
	options := &Options{}

	// 应用所有选项
	for _, opt := range opts {
		opt(options)
	}

	// 目前只支持企业微信webhook
	if options.WechatWebhookURL != "" {
		return &WechatAlertAdapter{
			client: NewWechatAlertClient(options.WechatWebhookURL),
		}, nil
	}

	return nil, fmt.Errorf("no valid alert channel configured")
}

// WechatAlertAdapter 企业微信告警适配器
type WechatAlertAdapter struct {
	client *WechatAlertClient
}

// SendAlert 发送告警消息（根据级别格式化）
func (a *WechatAlertAdapter) SendAlert(level, title, content string) error {
	// 根据告警级别设置不同的Markdown格式
	var levelIcon string
	switch AlertLevel(level) {
	case AlertLevelEmergency:
		levelIcon = "🚨"
	case AlertLevelCritical:
		levelIcon = "🔴"
	case AlertLevelWarning:
		levelIcon = "⚠️"
	case AlertLevelInfo:
		levelIcon = "ℹ️"
	case AlertLevelDebug:
		levelIcon = "🐛"
	default:
		levelIcon = "📢"
	}

	// 构造Markdown内容
	markdownContent := fmt.Sprintf("%s **[告警]** %s\n\n**级别**: %s\n**标题**: %s\n**内容**: %s",
		levelIcon, level, level, title, content)

	return a.client.SendMarkdownMessage(markdownContent)
}

// SendText 发送文本告警
func (a *WechatAlertAdapter) SendText(content string, mentionedList, mentionedMobileList []string) error {
	return a.client.SendTextMessage(content, mentionedList, mentionedMobileList)
}

// SendMarkdown 发送Markdown格式告警
func (a *WechatAlertAdapter) SendMarkdown(content string) error {
	return a.client.SendMarkdownMessage(content)
}

// SendMarkdownV2 发送MarkdownV2格式告警
func (a *WechatAlertAdapter) SendMarkdownV2(content string) error {
	return a.client.SendMarkdownV2Message(content)
}

// DefaultAlertClient 默认告警客户端
var DefaultAlertClient AlertClient

// InitDefaultAlertClient 初始化默认告警客户端
func InitDefaultAlertClient(opts ...Option) error {
	client, err := NewAlertClient(opts...)
	if err != nil {
		return err
	}
	DefaultAlertClient = client
	return nil
}

// SendAlert 发送告警（使用默认客户端）
func SendAlert(level, title, content string) error {
	if DefaultAlertClient == nil {
		return fmt.Errorf("default alert client not initialized")
	}
	return DefaultAlertClient.SendAlert(level, title, content)
}

// SendText 发送文本告警（使用默认客户端）
func SendText(content string, mentionedList, mentionedMobileList []string) error {
	if DefaultAlertClient == nil {
		return fmt.Errorf("default alert client not initialized")
	}
	return DefaultAlertClient.SendText(content, mentionedList, mentionedMobileList)
}

// SendMarkdown 发送Markdown告警（使用默认客户端）
func SendMarkdown(content string) error {
	if DefaultAlertClient == nil {
		return fmt.Errorf("default alert client not initialized")
	}
	return DefaultAlertClient.SendMarkdown(content)
}

// SendMarkdownV2 发送MarkdownV2告警（使用默认客户端）
func SendMarkdownV2(content string) error {
	if DefaultAlertClient == nil {
		return fmt.Errorf("default alert client not initialized")
	}
	return DefaultAlertClient.SendMarkdownV2(content)
}
