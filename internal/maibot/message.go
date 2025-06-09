package maibot

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageSegment represents a message segment
type MessageSegment struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// UserInfo represents user information
type UserInfo struct {
	Platform     string `json:"platform"`
	UserID       string `json:"user_id"`
	UserNickname string `json:"user_nickname,omitempty"`
	UserCardname string `json:"user_cardname,omitempty"`
}

// GroupInfo represents group information
type GroupInfo struct {
	Platform  string `json:"platform"`
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name,omitempty"`
}

// FormatInfo represents supported message formats
type FormatInfo struct {
	ContentFormat []string `json:"content_format,omitempty"`
	AcceptFormat  []string `json:"accept_format,omitempty"`
}

// TemplateInfo represents template information
type TemplateInfo struct {
	TemplateItems   map[string]interface{} `json:"template_items,omitempty"`
	TemplateName    string                 `json:"template_name,omitempty"`
	TemplateDefault bool                   `json:"template_default,omitempty"`
}

// MessageInfo represents message metadata
type MessageInfo struct {
	Platform         string                 `json:"platform"`
	MessageID        string                 `json:"message_id,omitempty"`
	Time             float64                `json:"time,omitempty"`
	UserInfo         *UserInfo              `json:"user_info,omitempty"`
	GroupInfo        *GroupInfo             `json:"group_info,omitempty"`
	FormatInfo       *FormatInfo            `json:"format_info,omitempty"`
	TemplateInfo     *TemplateInfo          `json:"template_info,omitempty"`
	AdditionalConfig map[string]interface{} `json:"additional_config,omitempty"`
}

// MessageBase represents the complete message structure
type MessageBase struct {
	MessageInfo    MessageInfo    `json:"message_info"`
	MessageSegment MessageSegment `json:"message_segment"`
	RawMessage     string         `json:"raw_message,omitempty"`
}

// NewTextSegment creates a text message segment
func NewTextSegment(text string) MessageSegment {
	return MessageSegment{
		Type: "text",
		Data: text,
	}
}

// NewVoiceSegment creates a wav voice message segment
func NewVoiceSegment(base64Data string) MessageSegment {
	return MessageSegment{
		Type: "voice",
		Data: base64Data,
	}
}

// NewImageSegment creates an image message segment
func NewImageSegment(base64Data string) MessageSegment {
	return MessageSegment{
		Type: "image",
		Data: base64Data,
	}
}

// NewEmojiSegment creates an emoji message segment
func NewEmojiSegment(emoji string) MessageSegment {
	return MessageSegment{
		Type: "emoji",
		Data: emoji,
	}
}

// NewAtSegment creates an @ mention message segment
func NewAtSegment(userID string) MessageSegment {
	return MessageSegment{
		Type: "at",
		Data: userID,
	}
}

// NewReplySegment creates a reply message segment
func NewReplySegment(messageID string) MessageSegment {
	return MessageSegment{
		Type: "reply",
		Data: messageID,
	}
}

// NewSegList creates a segment list containing multiple segments
func NewSegList(segments []MessageSegment) MessageSegment {
	return MessageSegment{
		Type: "seglist",
		Data: segments,
	}
}

// NewMessageBase creates a new MessageBase with the given parameters
func NewMessageBase(platform, messageID string, userInfo *UserInfo, groupInfo *GroupInfo, segments []MessageSegment) *MessageBase {
	messageInfo := MessageInfo{
		Platform:  platform,
		MessageID: messageID,
		Time:      float64(time.Now().Unix()) + float64(time.Now().Nanosecond())/1e9,
		UserInfo:  userInfo,
		GroupInfo: groupInfo,
	}

	messageSegment := NewSegList(segments)

	return &MessageBase{
		MessageInfo:    messageInfo,
		MessageSegment: messageSegment,
	}
}

// NewSimpleTextMessage creates a simple text message
func NewSimpleTextMessage(platform, messageID, userID, text string) *MessageBase {
	userInfo := &UserInfo{
		Platform: platform,
		UserID:   userID,
	}

	segments := []MessageSegment{NewTextSegment(text)}

	return NewMessageBase(platform, messageID, userInfo, nil, segments)
}

// NewGroupTextMessage creates a group text message
func NewGroupTextMessage(platform, messageID, userID, groupID, text string) *MessageBase {
	userInfo := &UserInfo{
		Platform: platform,
		UserID:   userID,
	}

	groupInfo := &GroupInfo{
		Platform: platform,
		GroupID:  groupID,
	}

	segments := []MessageSegment{NewTextSegment(text)}

	return NewMessageBase(platform, messageID, userInfo, groupInfo, segments)
}

// ToJSON converts MessageBase to JSON string
func (mb *MessageBase) ToJSON() (string, error) {
	data, err := json.Marshal(mb)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON creates MessageBase from JSON string
func FromJSON(jsonStr string) (*MessageBase, error) {
	var mb MessageBase
	err := json.Unmarshal([]byte(jsonStr), &mb)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to message: %w", err)
	}
	return &mb, nil
}

// GetSegments returns the segments from a seglist message segment
func (mb *MessageBase) GetSegments() ([]MessageSegment, error) {
	if mb.MessageSegment.Type != "seglist" {
		return nil, fmt.Errorf("message segment is not a seglist")
	}

	segments, ok := mb.MessageSegment.Data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid seglist data format")
	}

	result := make([]MessageSegment, 0, len(segments))
	for _, seg := range segments {
		segData, err := json.Marshal(seg)
		if err != nil {
			continue
		}

		var msgSeg MessageSegment
		if err := json.Unmarshal(segData, &msgSeg); err != nil {
			continue
		}

		result = append(result, msgSeg)
	}

	return result, nil
}

// GetTextContent extracts all text content from the message
func (mb *MessageBase) GetTextContent() string {
	// Handle single segment (like the example message)
	if mb.MessageSegment.Type == "text" {
		if text, ok := mb.MessageSegment.Data.(string); ok {
			return text
		}
		return ""
	}

	// Handle seglist
	segments, err := mb.GetSegments()
	if err != nil {
		return ""
	}

	var textContent string
	for _, seg := range segments {
		if seg.Type == "text" {
			if text, ok := seg.Data.(string); ok {
				textContent += text
			}
		}
	}

	return textContent
}

// IsGroupMessage checks if this is a group message
func (mb *MessageBase) IsGroupMessage() bool {
	return mb.MessageInfo.GroupInfo != nil
}

// GetSenderID returns the sender's user ID
func (mb *MessageBase) GetSenderID() string {
	if mb.MessageInfo.UserInfo != nil {
		return mb.MessageInfo.UserInfo.UserID
	}
	return ""
}

// GetGroupID returns the group ID if this is a group message
func (mb *MessageBase) GetGroupID() string {
	if mb.MessageInfo.GroupInfo != nil {
		return mb.MessageInfo.GroupInfo.GroupID
	}
	return ""
}
