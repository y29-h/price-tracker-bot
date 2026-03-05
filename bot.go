package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Зберігаємо товари кожного користувача
// map[chatID] = список посилань
var userProducts = make(map[int64][]string)

func main() {
	bot, err := tgbotapi.NewBotAPI("ТВІЙ_ТОКЕН")
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

		// Команда /list — показати всі збережені товари
		if msg == "/list" {
			products := userProducts[chatID]
			if len(products) == 0 {
				bot.Send(tgbotapi.NewMessage(chatID, "Список порожній"))
				continue
			}
			bot.Send(tgbotapi.NewMessage(chatID, strings.Join(products, "\n")))
			continue
		}

		// Якщо посилання — парсимо і зберігаємо
		if strings.HasPrefix(msg, "https://rozetka.com.ua") {
			price, err := fetchPrice(msg)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Помилка"))
				continue
			}

			// Додаємо в список цього користувача
			userProducts[chatID] = append(userProducts[chatID], msg)

			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Збережено! Ціна: %s", price)))
			continue
		}

		bot.Send(tgbotapi.NewMessage(chatID, "Надішли посилання з rozetka.com.ua"))
	}
}
