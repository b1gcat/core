package alert

// WechatWebhookMessage 企业微信webhook消息统一结构
type WechatWebhookMessage struct {
	MsgType      string               `json:"msgtype"`
	Text         *TextMessage         `json:"text,omitempty"`
	Markdown     *MarkdownMessage     `json:"markdown,omitempty"`
	MarkdownV2   *MarkdownV2Message   `json:"markdown_v2,omitempty"`
	Image        *ImageMessage        `json:"image,omitempty"`
	News         *NewsMessage         `json:"news,omitempty"`
	File         *FileMessage         `json:"file,omitempty"`
	Voice        *VoiceMessage        `json:"voice,omitempty"`
	TemplateCard *TemplateCardMessage `json:"template_card,omitempty"`
}

// TextMessage 文本消息
type TextMessage struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// MarkdownMessage Markdown消息
type MarkdownMessage struct {
	Content string `json:"content"`
}

// MarkdownV2Message MarkdownV2消息
type MarkdownV2Message struct {
	Content string `json:"content"`
}

// ImageMessage 图片消息
type ImageMessage struct {
	Base64 string `json:"base64"`
	MD5    string `json:"md5"`
}

// NewsMessage 图文消息
type NewsMessage struct {
	Articles []NewsArticle `json:"articles"`
}

// NewsArticle 图文消息中的文章
type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl,omitempty"`
}

// FileMessage 文件消息
type FileMessage struct {
	MediaID string `json:"media_id"`
}

// VoiceMessage 语音消息
type VoiceMessage struct {
	MediaID  string `json:"media_id"`
	Duration int    `json:"duration,omitempty"`
}

// TemplateCardMessage 模板卡片消息
type TemplateCardMessage struct {
	CardType string `json:"card_type"`
	// 根据不同card_type包含不同的结构
	// 这里只定义常用的文本通知模板卡片
	TextNoticeTemplateCard *TextNoticeTemplateCard `json:"text_notice_template_card,omitempty"`
}

// TextNoticeTemplateCard 文本通知模板卡片
type TextNoticeTemplateCard struct {
	MainTitle             TemplateCardTitle   `json:"main_title"`
	SubTitleText          string              `json:"sub_title_text,omitempty"`
	HorizontalContentList []HorizontalContent `json:"horizontal_content_list,omitempty"`
	JumpList              []JumpInfo          `json:"jump_list,omitempty"`
	CardAction            CardAction          `json:"card_action,omitempty"`
}

// TemplateCardTitle 模板卡片标题
type TemplateCardTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// HorizontalContent 模板卡片水平内容
type HorizontalContent struct {
	KeyName   string `json:"keyname"`
	ValueText string `json:"value_text,omitempty"`
	ValueURL  string `json:"value_url,omitempty"`
}

// JumpInfo 模板卡片跳转信息
type JumpInfo struct {
	Type     int    `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// CardAction 模板卡片动作
type CardAction struct {
	Type     int    `json:"type"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	WechatWebhookURL string `json:"wechat_webhook_url"`
}

// Options 告警客户端选项
type Options struct {
	WechatWebhookURL string
	// 可以添加更多选项
}

// Option 选项函数类型
type Option func(*Options)

// WithWechatWebhookURL 设置企业微信webhook URL
func WithWechatWebhookURL(url string) Option {
	return func(opts *Options) {
		opts.WechatWebhookURL = url
	}
}
