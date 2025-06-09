package telegram

import (
	"strconv"
	"strings"

	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/config"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/logger"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/maibot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var botInstance *tgbotapi.BotAPI

func StartBot() {
	bot, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		panic(err)
	}

	botInstance = bot
	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many  types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		if !MessageFilter(*update.Message) {
			logger.Info("Message filtered out: %s", update.Message.Text)
			continue
		}

		go HandleMessage(*update.Message)
	}
}

// SendMessageToTelegram sends a MessageBase message to Telegram
func SendMessageToTelegram(messageBase *maibot.MessageBase) error {
	if botInstance == nil {
		logger.Error("Telegram bot not initialized")
		return nil
	}

	// Parse group/chat ID
	var chatID int64
	var err error

	if messageBase.IsGroupMessage() {
		chatID, err = strconv.ParseInt(messageBase.GetGroupID(), 10, 64)
	} else {
		// For private messages, we need to use the user ID as chat ID
		chatID, err = strconv.ParseInt(messageBase.GetSenderID(), 10, 64)
	}

	if err != nil {
		logger.Error("Invalid chat ID: %v", err)
		return err
	}

	// Convert MessageBase to Telegram message using enhanced conversion
	return convertAndSendMessage(messageBase, chatID)
}

// convertAndSendMessage converts MessageBase segments to Telegram message format and sends it
func convertAndSendMessage(messageBase *maibot.MessageBase, chatID int64) error {
	var replyToMessageID int
	var messageText strings.Builder

	// Handle single segment (like the example message)
	if messageBase.MessageSegment.Type != "seglist" {
		processSegment(messageBase.MessageSegment, &messageText, &replyToMessageID)
	} else {
		// Handle seglist
		segments, err := messageBase.GetSegments()
		if err != nil {
			logger.Error("Failed to get message segments: %v", err)
			return err
		}

		for _, segment := range segments {
			processSegment(segment, &messageText, &replyToMessageID)
		}
	}

	// Handle empty message text
	if messageText.Len() == 0 {
		messageText.WriteString("[空消息]")
	}

	// Create and send message
	msg := tgbotapi.NewMessage(chatID, messageText.String())
	if replyToMessageID != 0 {
		msg.ReplyToMessageID = replyToMessageID
	}

	_, err := botInstance.Send(msg)
	if err != nil {
		logger.Error("Failed to send message to Telegram: %v", err)
		return err
	}

	logger.Info("Message sent to Telegram successfully")
	return nil
}

// processSegment processes a single message segment
func processSegment(segment maibot.MessageSegment, messageText *strings.Builder, replyToMessageID *int) {
	switch segment.Type {
	case "text":
		if text, ok := segment.Data.(string); ok {
			messageText.WriteString(text)
		}
	case "reply":
		if replyID, ok := segment.Data.(string); ok {
			if id, err := strconv.Atoi(replyID); err == nil {
				*replyToMessageID = id
			}
		}
	case "at":
		if userID, ok := segment.Data.(string); ok {
			messageText.WriteString("@" + userID)
		}
	case "emoji":
		if emoji, ok := segment.Data.(string); ok {
			messageText.WriteString(emoji)
		}
	case "image":
		messageText.WriteString("[图片]")
	default:
		logger.Info("Unsupported segment type: %s", segment.Type)
	}
}
