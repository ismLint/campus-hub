package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Инициализация бота (замените "YOUR_BOT_TOKEN" на ваш токен)
	bot, err := tgbotapi.NewBotAPI("8619648172:AAGwsZi-jGXB2urTxTHdVzdRBoP4ypzNvQk")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизован как аккаунт %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// 1. Обработка нажатий на инлайн-кнопки (Callback Query)
		if update.CallbackQuery != nil {
			// Отвечаем на callback, чтобы убрать часики загрузки на кнопке
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Вы нажали кнопку!")
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
			}

			// Отправляем сообщение в ответ на нажатие
			text := "Вы выбрали: " + update.CallbackQuery.Data
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
			bot.Send(msg)
			continue
		}

		// 2. Обработка обычных текстовых сообщений (вызов меню)
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, выберите пункт меню:")

			// Создание инлайн-клавиатуры
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				// Первый ряд кнопок
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Add Task", "addTask"),
					tgbotapi.NewInlineKeyboardButtonData("Task list", "TaskList"),
				),
			)

			// Привязываем клавиатуру к сообщению
			msg.ReplyMarkup = numericKeyboard

			// Отправляем сообщение пользователю
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}
}
