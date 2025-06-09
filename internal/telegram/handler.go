package telegram

import (
	"encoding/base64"
	"github.com/davecgh/go-spew/spew"
	"io"
	"strconv"

	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/logger"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/maibot"
	"github.com/cshum/vipsgen/vips"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(message tgbotapi.Message) {
	logger.Info("incoming message")
	spew.Dump(message)

	messageBase := ConvertTelegramToMessageBase(message)
	if messageBase != nil {
		SendToMaiBot(messageBase)
	}
}

func ConvertTelegramToMessageBase(tgMsg tgbotapi.Message) *maibot.MessageBase {
	platform := "telegram"
	messageID := strconv.Itoa(tgMsg.MessageID)
	userID := strconv.FormatInt(tgMsg.From.ID, 10)

	userInfo := &maibot.UserInfo{
		Platform: platform,
		UserID:   userID,
	}

	if tgMsg.From.UserName != "" {
		userInfo.UserNickname = tgMsg.From.UserName
	} else if tgMsg.From.FirstName != "" {
		userInfo.UserNickname = tgMsg.From.FirstName
		if tgMsg.From.LastName != "" {
			userInfo.UserNickname += " " + tgMsg.From.LastName
		}
	}

	var groupInfo *maibot.GroupInfo
	if tgMsg.Chat.IsGroup() || tgMsg.Chat.IsSuperGroup() {
		groupInfo = &maibot.GroupInfo{
			Platform: platform,
			GroupID:  strconv.FormatInt(tgMsg.Chat.ID, 10),
		}
		if tgMsg.Chat.Title != "" {
			groupInfo.GroupName = tgMsg.Chat.Title
		}
	}

	var segments []maibot.MessageSegment

	if tgMsg.ReplyToMessage != nil {
		replyID := strconv.Itoa(tgMsg.ReplyToMessage.MessageID)
		segments = append(segments, maibot.NewReplySegment(replyID))
	}

	if tgMsg.Text != "" {
		segments = append(segments, maibot.NewTextSegment(tgMsg.Text))
	}

	if tgMsg.Photo != nil && len(tgMsg.Photo) > 0 {
		largestPhoto := tgMsg.Photo[len(tgMsg.Photo)-1]
		fileContent, err := GetFileContent(largestPhoto.FileID, botInstance)
		if err != nil {
			logger.Error("Failed to get photo content: %v", err)
			segments = append(segments, maibot.NewTextSegment("[图片]"))
		} else {
			base64Data := base64.StdEncoding.EncodeToString(fileContent)
			segments = append(segments, maibot.NewImageSegment(base64Data))
		}
	}

	// temporary disabled 'cuz maibot does not support receive voice yet
	//if tgMsg.Voice != nil {
	//	fileContent, err := GetFileContent(tgMsg.Voice.FileID, botInstance)
	//	if err != nil {
	//		logger.Error("Failed to get voice content: %v", err)
	//		segments = append(segments, maibot.NewTextSegment("[语音]"))
	//	} else {
	//		wavData, err := audio.ToWav(fileContent)
	//		if err != nil {
	//			logger.Error("Failed to convert voice to WAV: %v", err)
	//			segments = append(segments, maibot.NewTextSegment("[语音]"))
	//		} else {
	//			base64Data := base64.StdEncoding.EncodeToString(wavData)
	//			segments = append(segments, maibot.NewVoiceSegment(base64Data))
	//		}
	//	}
	//}

	if tgMsg.Sticker != nil {
		base64Data, err := convertStickerToBase64PNG(tgMsg.Sticker.FileID)
		if err != nil {
			logger.Error("Failed to convert sticker to PNG: %v", err)
			segments = append(segments, maibot.NewTextSegment("[贴纸]"))
		} else {
			segments = append(segments, maibot.NewEmojiSegment(base64Data))
		}
	}

	if len(segments) == 0 {
		segments = append(segments, maibot.NewTextSegment("[未支持的消息类型]"))
	}

	messageInfo := maibot.MessageInfo{
		Platform:  platform,
		MessageID: messageID,
		Time:      float64(tgMsg.Date),
		UserInfo:  userInfo,
		GroupInfo: groupInfo,
	}

	messageBase := &maibot.MessageBase{
		MessageInfo:    messageInfo,
		MessageSegment: maibot.NewSegList(segments),
	}

	if tgMsg.Text != "" {
		messageBase.RawMessage = tgMsg.Text
	} else if tgMsg.Caption != "" {
		messageBase.RawMessage = tgMsg.Caption
	}

	return messageBase
}

func SendToMaiBot(messageBase *maibot.MessageBase) {
	client := maibot.GetDefaultClient()
	if client == nil {
		logger.Error("MaiBot client not initialized")
		return
	}

	err := client.SendMessageBase(messageBase)
	if err != nil {
		logger.Error("Failed to send message to MaiBot: %v", err)
		return
	}

	logger.Info("Message sent to MaiBot successfully")
}

// convertStickerToBase64PNG downloads a Telegram sticker (WebP) and converts it to base64-encoded PNG
func convertStickerToBase64PNG(fileID string) (string, error) {
	// Get bot instance from the global variable in telegram.go
	if botInstance == nil {
		return "", io.EOF // Use a standard error; we'll log the details in the caller
	}

	webpData, err := GetFileContent(fileID, botInstance)
	if err != nil {
		return "", err
	}

	// Initialize vips
	vips.Startup(nil)
	defer vips.Shutdown()

	// Load WebP image from buffer
	image, err := vips.NewWebploadBuffer(webpData, nil)
	if err != nil {
		return "", err
	}
	defer image.Close()

	// Save as PNG to buffer
	pngData, err := image.PngsaveBuffer(nil)
	if err != nil {
		return "", err
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(pngData), nil
}
