package telegram

import (
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
)

func chatIDFilter(filter config.MessageFilter, id int64) bool {
	isWhitelist := filter.Mode == "whitelist"
	for _, chatID := range filter.List {
		if chatID == id {
			return isWhitelist
		}
	}
	return !isWhitelist
}

func MessageFilter(message tgbotapi.Message) bool {
	for _, id := range config.Get().MessageFilter.BannedUsers {
		if message.From.ID == id {
			return false
		}
	}

	isPrivate := message.Chat.IsPrivate()
	filters := config.Get().MessageFilter
	return chatIDFilter(lo.Ternary(isPrivate, filters.Private, filters.Groups), message.Chat.ID)
}
