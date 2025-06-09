# MaiBot Telegram Adapter

一个连接MaiBot服务器和Telegram的适配器，通过WebSocket实现消息双向转发。

## 功能

- 通过WebSocket连接MaiBot服务器
- 转发Telegram消息到MaiBot
- 转发MaiBot消息到Telegram
- 支持消息过滤
- 自动重连

## 配置

创建 `config.toml` 文件：

```toml
Platform = "telegram"
TelegramBotToken = "YOUR_BOT_TOKEN"

[MaiBot]
URL = "ws://localhost:8080"

[MessageFilter]
BannedUsers = []

[MessageFilter.Groups]
Mode = "whitelist"
List = []

[MessageFilter.Private]
Mode = "blacklist"
List = []
```

## 运行

```bash
go run cmd/maibot-telegram-adapter.go
```

或者：

```bash
go build -o maibot-telegram-adapter cmd/maibot-telegram-adapter.go
./maibot-telegram-adapter
```

## 依赖

- Go 1.24.2+
- Telegram Bot Token
- MaiBot
