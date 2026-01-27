# 0trust Makefile

# 可执行文件名称
SERVER_BINARY := 0trust-server
CLIENT_BINARY := 0trust-client

LDFLAGS := -w -s

# 输出目录
OUTPUT_DIR := bin

# 默认目标
all: server client

# 编译服务端（仅Linux）
server:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(SERVER_BINARY) ./cmd/server

# 编译客户端（多平台）
client: client-linux client-windows client-darwin

client-linux:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(CLIENT_BINARY)-linux ./cmd/client

client-windows:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(CLIENT_BINARY)-windows.exe ./cmd/client

client-darwin:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(CLIENT_BINARY)-darwin ./cmd/client

# 清理
clean:
	rm -rf $(OUTPUT_DIR)

# 安装服务端（Linux）

.PHONY: all server client client-linux client-windows client-darwin clean install