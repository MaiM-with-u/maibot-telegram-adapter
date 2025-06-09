# MaiBot WebSocket Client

基于 `maibot_message_base.md` 规范实现的 WebSocket 消息收发客户端。

## 功能特性

- 支持认证头部（platform、authorization）
- 完整的消息结构体定义（MessageBase、MessageInfo、MessageSegment等）
- 自动重连机制
- 心跳保活
- 多种消息类型支持（文本、图片、表情、@提及、回复）
- 群组和私聊消息支持

## 快速开始

### 1. 创建客户端

```go
import "github.com/MaiM-with-u/maibot-telegram-adapter/internal/maibot"

// 创建客户端
client := maibot.NewClient("ws://localhost:8090/ws", "telegram", "your_auth_token")

// 设置消息处理函数
client.SetMessageHandler(func(msg *maibot.MessageBase) {
    logger.Info("收到消息: %s", msg.GetTextContent())
})
```

### 2. 连接和监听

```go
// 连接到WebSocket服务器
err := client.Connect()
if err != nil {
    logger.Error("连接失败: %v", err)
    return
}
defer client.Close()

// 开始监听消息（阻塞）
client.Listen()
```

### 3. 发送消息

#### 发送简单文本消息

```go
// 私聊消息
err := client.SendTextMessage("msg_001", "user123", "你好！")

// 群组消息
err := client.SendTextMessage("msg_002", "user123", "大家好！", "group456")
```

#### 发送复杂消息

```go
// 创建用户信息
userInfo := &maibot.UserInfo{
    Platform:     "telegram",
    UserID:       "user123",
    UserNickname: "张三",
}

// 创建群组信息（可选）
groupInfo := &maibot.GroupInfo{
    Platform:  "telegram",
    GroupID:   "group456",
    GroupName: "开发群",
}

// 创建消息片段
segments := []maibot.MessageSegment{
    maibot.NewTextSegment("Hello "),
    maibot.NewAtSegment("user789"),
    maibot.NewTextSegment("，请查看："),
    maibot.NewImageSegment("base64_image_data"),
    maibot.NewEmojiSegment("😀"),
}

// 创建并发送消息
msg := maibot.NewMessageBase("telegram", "msg_003", userInfo, groupInfo, segments)
err := client.SendMessageBase(msg)
```

## 消息类型

### 支持的消息片段类型

- **文本片段**: `NewTextSegment(text)`
- **图片片段**: `NewImageSegment(base64Data)`
- **表情片段**: `NewEmojiSegment(emoji)`
- **@提及片段**: `NewAtSegment(userID)`
- **回复片段**: `NewReplySegment(messageID)`

### 便捷函数

```go
// 创建简单文本消息
msg := maibot.NewSimpleTextMessage("telegram", "msg_001", "user123", "Hello")

// 创建群组文本消息
msg := maibot.NewGroupTextMessage("telegram", "msg_002", "user123", "group456", "Hello Group")
```

## 消息处理

### 解析接收到的消息

```go
func handleMessage(msg *maibot.MessageBase) {
    // 检查是否为群组消息
    if msg.IsGroupMessage() {
        logger.Info("群组消息，群组ID: %s", msg.GetGroupID())
    }
    
    // 获取发送者ID
    senderID := msg.GetSenderID()
    
    // 获取文本内容
    textContent := msg.GetTextContent()
    
    // 解析所有消息片段
    segments, err := msg.GetSegments()
    if err == nil {
        for _, segment := range segments {
            switch segment.Type {
            case "text":
                // 处理文本
            case "image":
                // 处理图片
            case "at":
                // 处理@提及
            case "reply":
                // 处理回复
            }
        }
    }
}
```

## 配置要求

客户端需要在配置文件中设置MaiBot URL：

```toml
[maibot]
url = "ws://localhost:8090/ws"
```

## 错误处理

客户端具有自动重连功能，当连接断开时会自动尝试重连。所有的网络错误都会记录在日志中。

认证失败时（错误码1008），客户端会停止重连尝试。

## 注意事项

1. 所有字符串使用UTF-8编码
2. 图片数据必须进行Base64编码
3. 消息ID应在平台内保持唯一
4. 大消息（>1MB）可能会被服务器拒绝