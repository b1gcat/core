package mqtt

const (
	// Topics for MQTT communication
	TopicPolicy     = "mqtt/policy"
	TopicClientInfo = "mqtt/client/info"
	TopicSystemCmd  = "mqtt/system/cmd"

	// Default broker address
	DefaultBrokerAddress = "tcp://localhost:1883"
)

// Policy represents the configuration policy sent from server to clients
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	Timestamp   int64                  `json:"timestamp"`
	GroupID     string                 `json:"group_id,omitempty"` // 组ID，为空表示全局策略
}

// DefaultGroup is the default group for clients
const DefaultGroup = "default"

// ClientInfo represents the information collected from clients
type ClientInfo struct {
	ClientID   string         `json:"client_id"`
	IPAddress  string         `json:"ip_address"`
	Connected  bool           `json:"connected"`
	LastSeen   int64          `json:"last_seen"`
	Metadata   map[string]any `json:"metadata"`
	PolicyInfo *Policy        `json:"policy_info,omitempty"`
	Group      string         `json:"group,omitempty"` // 客户端所属组
}

// SystemCommand represents commands sent from server to clients
type SystemCommand struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
