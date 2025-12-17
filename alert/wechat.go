package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WechatAlertClient 企业微信告警客户端
type WechatAlertClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewWechatAlertClient 创建新的企业微信告警客户端
func NewWechatAlertClient(webhookURL string) *WechatAlertClient {
	return &WechatAlertClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMessage 发送消息
func (c *WechatAlertClient) SendMessage(msg *WechatWebhookMessage) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// 验证消息类型
	if msg.MsgType == "" {
		return fmt.Errorf("msgtype must be specified")
	}

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned non-200 status: %d", resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查企业微信返回的错误
	if result.Errcode != 0 {
		return fmt.Errorf("wechat webhook error: %d - %s", result.Errcode, result.Errmsg)
	}

	return nil
}

// SendTextMessage 发送文本消息
func (c *WechatAlertClient) SendTextMessage(content string, mentionedList, mentionedMobileList []string) error {
	msg := &WechatWebhookMessage{
		MsgType: "text",
		Text: &TextMessage{
			Content:             content,
			MentionedList:       mentionedList,
			MentionedMobileList: mentionedMobileList,
		},
	}

	return c.SendMessage(msg)
}

// SendMarkdownMessage 发送Markdown消息
func (c *WechatAlertClient) SendMarkdownMessage(content string) error {
	msg := &WechatWebhookMessage{
		MsgType: "markdown",
		Markdown: &MarkdownMessage{
			Content: content,
		},
	}

	return c.SendMessage(msg)
}

// SendMarkdownV2Message 发送MarkdownV2消息
func (c *WechatAlertClient) SendMarkdownV2Message(content string) error {
	msg := &WechatWebhookMessage{
		MsgType: "markdown_v2",
		MarkdownV2: &MarkdownV2Message{
			Content: content,
		},
	}

	return c.SendMessage(msg)
}
