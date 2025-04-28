
## ssh隧道
```bash
ssh -i key_path -p 19043 \
-L 3320:localhost:3320 \
-L 9020:localhost:9020 \
username@ip_address
```

## 设置
在./backend/.env中配置自己的设置

## 运行服务器
在./backend/中
```go
go run main.go
```
