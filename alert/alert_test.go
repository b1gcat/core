package alert

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestWechatWebhookMessageSerialization 测试企业微信消息结构的JSON序列化
func TestWechatWebhookMessageSerialization(t *testing.T) {
	// 测试文本消息
	textMsg := &WechatWebhookMessage{
		MsgType: "text",
		Text: &TextMessage{
			Content:             "测试文本消息",
			MentionedList:       []string{"user1", "user2"},
			MentionedMobileList: []string{"13800138000"},
		},
	}

	data, err := json.Marshal(textMsg)
	if err != nil {
		t.Fatalf("Failed to marshal text message: %v", err)
	}

	// 验证JSON包含正确的字段
	expectedFields := []string{"msgtype", "text", "content", "mentioned_list", "mentioned_mobile_list"}
	for _, field := range expectedFields {
		if !containsField(string(data), field) {
			t.Errorf("Text message JSON should contain field '%s'", field)
		}
	}

	// 测试Markdown消息
	markdownMsg := &WechatWebhookMessage{
		MsgType: "markdown",
		Markdown: &MarkdownMessage{
			Content: "# 测试Markdown消息\n这是一条**Markdown**格式的消息",
		},
	}

	data, err = json.Marshal(markdownMsg)
	if err != nil {
		t.Fatalf("Failed to marshal markdown message: %v", err)
	}

	// 验证JSON包含正确的字段
	expectedFields = []string{"msgtype", "markdown", "content"}
	for _, field := range expectedFields {
		if !containsField(string(data), field) {
			t.Errorf("Markdown message JSON should contain field '%s'", field)
		}
	}
}

// TestNewAlertClient 测试创建告警客户端
func TestNewAlertClient(t *testing.T) {
	// 测试创建企业微信告警客户端
	client, err := NewAlertClient(WithWechatWebhookURL("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test-key"))
	if err != nil {
		t.Fatalf("Failed to create alert client: %v", err)
	}

	if client == nil {
		t.Fatal("Alert client should not be nil")
	}

	// 测试创建无选项的告警客户端
	_, err = NewAlertClient()
	if err == nil {
		t.Fatal("Expected error when creating alert client with no options")
	}
}

// TestAlertLevel 测试告警级别
func TestAlertLevel(t *testing.T) {
	levels := []AlertLevel{
		AlertLevelEmergency,
		AlertLevelCritical,
		AlertLevelWarning,
		AlertLevelInfo,
		AlertLevelDebug,
	}

	expectedNames := []string{"emergency", "critical", "warning", "info", "debug"}

	for i, level := range levels {
		if string(level) != expectedNames[i] {
			t.Errorf("AlertLevel %d should be '%s', got '%s'", i, expectedNames[i], level)
		}
	}
}

// TestWechatAlertAdapter 测试企业微信告警适配器
func TestWechatAlertAdapter(t *testing.T) {
	// 创建测试客户端
	client, err := NewAlertClient(WithWechatWebhookURL("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test-key"))
	if err != nil {
		t.Fatalf("Failed to create alert client: %v", err)
	}

	// 这里不实际发送HTTP请求，只测试客户端接口的可用性
	// 实际测试需要mock HTTP客户端

	// 测试接口方法是否存在
	_, ok := client.(*WechatAlertAdapter)
	if !ok {
		t.Fatal("Alert client should be of type WechatAlertAdapter")
	}
}

// containsField 检查JSON字符串是否包含指定字段
func containsField(jsonStr, field string) bool {
	// 简单的字符串包含检查
	return len(jsonStr) > 0 && (jsonStr[0] == '{' || jsonStr[0] == '[') &&
		(strings.Contains(jsonStr, `"`+field+`"`))
}
