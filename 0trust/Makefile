# 0trust Makefile

# 可执行文件名称
SERVER_BINARY := 0trust-server
CLIENT_BINARY := 0trust-client

LDFLAGS := -w -s

# 输出目录
OUTPUT_DIR := bin

# 源文件
SERVER_SRCS := server.go utils.go types.go main_server.go
CLIENT_SRCS := client.go utils.go types.go main_client.go

# 默认目标
all: server client

# 编译服务端（仅Linux）
server: $(SERVER_SRCS)
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(SERVER_BINARY) $(SERVER_SRCS)

# 编译客户端（多平台）
client: client-linux client-windows client-darwin

client-linux: $(CLIENT_SRCS)
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(CLIENT_BINARY)-linux $(CLIENT_SRCS)

client-windows: $(CLIENT_SRCS)
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(CLIENT_BINARY)-windows.exe $(CLIENT_SRCS)

client-darwin: $(CLIENT_SRCS)
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(CLIENT_BINARY)-darwin $(CLIENT_SRCS)

# 清理
clean:
	rm -rf $(OUTPUT_DIR)

# 安装服务端（Linux）
install: server
	cp $(OUTPUT_DIR)/$(SERVER_BINARY) /usr/local/bin/

.PHONY: all server client client-linux client-windows client-darwin clean install