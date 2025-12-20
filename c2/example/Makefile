all:
	go build -o client_only client_only.go
	go build -o server_only server_only.go

linux:
	GOOS=linux GOARCH=amd64 go build -o c2_server server_only.go

clean:
	rm -f client_only server_only c2_server
