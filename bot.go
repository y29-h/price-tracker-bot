package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("ВАШ_ТОКЕН_ТУТ")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Бот запущен: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message.Text
		chatID := update.Message.Chat.ID

		if msg == "/start" {
			bot.Send(tgbotapi.NewMessage(chatID, "Привіт! Кидай посилання на товар з Rozetka"))
			continue
		}

		// Якщо це посилання — парсимо ціну
		price, err := fetchPrice(msg)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Помилка при отриманні ціни"))
			continue
		}

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ціна: %s", price)))
	}
}
