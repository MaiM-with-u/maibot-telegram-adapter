package maibot

import (
	"fmt"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/logger"
	"time"
)

// ExampleMessageHandler demonstrates how to handle incoming messages
func ExampleMessageHandler(msg *MessageBase) {
	logger.Info("Processing message from platform: %s", msg.MessageInfo.Platform)
	
	if msg.IsGroupMessage() {
		logger.Info("Group message from group: %s", msg.GetGroupID())
	} else {
		logger.Info("Private message")
	}
	
	logger.Info("Sender: %s", msg.GetSenderID())
	logger.Info("Text content: %s", msg.GetTextContent())
	
	// Process different types of message segments
	segments, err := msg.GetSegments()
	if err != nil {
		logger.Error("Failed to get segments: %v", err)
		return
	}
	
	for i, segment := range segments {
		switch segment.Type {
		case "text":
			logger.Info("Segment %d [text]: %v", i, segment.Data)
		case "image":
			logger.Info("Segment %d [image]: image data received", i)
		case "at":
			logger.Info("Segment %d [at]: mentioned user %v", i, segment.Data)
		case "reply":
			logger.Info("Segment %d [reply]: replying to message %v", i, segment.Data)
		case "emoji":
			logger.Info("Segment %d [emoji]: %v", i, segment.Data)
		default:
			logger.Info("Segment %d [%s]: %v", i, segment.Type, segment.Data)
		}
	}
}

// ExampleUsage demonstrates how to use the maibot WebSocket client
func ExampleUsage() {
	// Create a new client with platform and auth token
	client := NewClient("ws://localhost:8090/ws", "telegram", "your_auth_token")
	
	// Set message handler
	client.SetMessageHandler(ExampleMessageHandler)
	
	// Connect to the WebSocket server
	err := client.Connect()
	if err != nil {
		logger.Error("Failed to connect: %v", err)
		return
	}
	defer client.Close()
	
	// Send a simple text message
	err = client.SendTextMessage("msg_001", "user123", "Hello from telegram adapter!")
	if err != nil {
		logger.Error("Failed to send text message: %v", err)
	}
	
	// Send a group text message
	err = client.SendTextMessage("msg_002", "user123", "Hello group!", "group456")
	if err != nil {
		logger.Error("Failed to send group text message: %v", err)
	}
	
	// Create and send a complex message with multiple segments
	userInfo := &UserInfo{
		Platform:     "telegram",
		UserID:       "user123",
		UserNickname: "Test User",
	}
	
	groupInfo := &GroupInfo{
		Platform:  "telegram",
		GroupID:   "group456",
		GroupName: "Test Group",
	}
	
	segments := []MessageSegment{
		NewTextSegment("Hello "),
		NewAtSegment("user789"),
		NewTextSegment(", please check this image: "),
		NewImageSegment("base64_encoded_image_data_here"),
		NewEmojiSegment("😀"),
	}
	
	complexMsg := NewMessageBase("telegram", "msg_003", userInfo, groupInfo, segments)
	err = client.SendMessageBase(complexMsg)
	if err != nil {
		logger.Error("Failed to send complex message: %v", err)
	}
	
	// Start listening for messages (this will block)
	logger.Info("Starting to listen for messages...")
	client.Listen()
}

// SendReplyMessage demonstrates how to send a reply message
func SendReplyMessage(client *Client, originalMessageID, userID, groupID, replyText string) error {
	userInfo := &UserInfo{
		Platform: "telegram",
		UserID:   userID,
	}
	
	var groupInfo *GroupInfo
	if groupID != "" {
		groupInfo = &GroupInfo{
			Platform: "telegram",
			GroupID:  groupID,
		}
	}
	
	segments := []MessageSegment{
		NewReplySegment(originalMessageID),
		NewTextSegment(replyText),
	}
	
	messageID := fmt.Sprintf("reply_%d", time.Now().Unix())
	msg := NewMessageBase("telegram", messageID, userInfo, groupInfo, segments)
	
	return client.SendMessageBase(msg)
}

// SendImageMessage demonstrates how to send an image message
func SendImageMessage(client *Client, userID, groupID, imageBase64 string) error {
	userInfo := &UserInfo{
		Platform: "telegram",
		UserID:   userID,
	}
	
	var groupInfo *GroupInfo
	if groupID != "" {
		groupInfo = &GroupInfo{
			Platform: "telegram",
			GroupID:  groupID,
		}
	}
	
	segments := []MessageSegment{
		NewTextSegment("Here's an image: "),
		NewImageSegment(imageBase64),
	}
	
	messageID := fmt.Sprintf("img_%d", time.Now().Unix())
	msg := NewMessageBase("telegram", messageID, userInfo, groupInfo, segments)
	
	return client.SendMessageBase(msg)
}