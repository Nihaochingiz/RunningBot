package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	file, err := os.Open("config.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var token string
	for scanner.Scan() {
		token = scanner.Text()
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	currentDate := time.Now().Format("2006-01-02") // преобразование в формат YYYY-MM-DD
	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		messageTemplate := "Привет, я телеграм бот для бега.\nСегодняшнее число: %s\n%sкм-%sмин"
		message := fmt.Sprintf(messageTemplate, currentDate, "", "")

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
		_, err = bot.Send(msg)
		if err != nil {
			panic(err)
		}

		// Ожидаем ответа от пользователя
		response := waitForResponse(updates, bot, update.Message.Chat.ID)

		km, minutes := parseResponse(response)
		fmt.Printf("Километраж: %s км, Время: %s\n", km, minutes)

		responseMessage := fmt.Sprintf("Введено: %s км - %s", km, minutes)
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
		_, err = bot.Send(msg)
		insert(currentDate, km, minutes)

		if err != nil {
			panic(err)
		}

	}

}

// Функция для ожидания и проверки ответа пользователя
func waitForResponse(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, chatID int64) string {
	for {
		update := <-updates
		if update.Message == nil {
			continue
		}
		response := update.Message.Text
		if isValidFormat(response) {
			return response
		} else {
			msg := tgbotapi.NewMessage(chatID, "Неверный формат. Пожалуйста, введите данные в формате 'км-мин'")
			_, err := bot.Send(msg)
			if err != nil {
				panic(err)
			}
		}
	}
}

// Функция для проверки формата ввода от пользователя
func isValidFormat(input string) bool {
	r := regexp.MustCompile(`^\d+км-\d+мин$`)
	return r.MatchString(input)
}

// Функция для парсинга данных от пользователя
func parseResponse(input string) (string, string) {
	values := strings.SplitN(input, "км-", 2)
	return values[0], values[1]
}

func insert(currentDate, km, minutes string) {
	db, err := sql.Open("", "")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Printf("У нас " + minutes)

	// Добавляем пробел между числом и словом "мин"
	formattedMinutes := strings.ReplaceAll(minutes, "мин", " мин")

	result, err := db.Exec("INSERT INTO running_statistics (date, distance, time) VALUES (?, ?, ?)", currentDate, km+" км", formattedMinutes)
	if err != nil {
		panic(err)
	}

	lastInsertId, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	fmt.Printf("ID добавленного объекта: %d\n", lastInsertId)
	fmt.Printf("Количество затронутых строк: %d\n", rowsAffected)
}
