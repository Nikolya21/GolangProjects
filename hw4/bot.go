package hw4

// сюда писать код

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

var (
	// @BotFather в телеграме даст вам токен. Если захотите потыкать своего бота через телегу - используйте именно его
	BotToken = "XXX"

	// Урл, в который будет стучаться телега при получении сообщения от пользователя.
	// Может быть как айпишником личной виртуалки, так и просто выдан сервисом для деплоя
	WebhookURL = "https://525f2cb5.ngrok.io"
)

var tasks []Task
var mu sync.Mutex

type Task struct {
	ID               int
	Text             string
	AuthorID         int64
	AuthorUsername   string
	AssigneeID       int64
	AssigneeUsername string
	Resolved         bool
}

var countID int = 0

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    "127.0.0.1:8081",
		Handler: mux,
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		query, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Не удалось прочитать тело запроса"))
			return
		}
		var tgBotStruct tgbotapi.Update
		err = json.Unmarshal(query, &tgBotStruct)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Не удалось преобразовать в структуру tgbotapi"))
			return
		}
		if tgBotStruct.Message == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Было отправлено пустое тело запроса"))
			return
		}
		if err := handleCommand(bot, tgBotStruct); err != nil {
			http.Error(w, "Не удалось обработать команду", http.StatusInternalServerError)
			return
		}
	})

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutDownCtx); err != nil {
			return err
		}
		err := <-serverErr
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil

	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	msg := update.Message
	if msg == nil || msg.From == nil {
		return nil
	}

	chatID := msg.Chat.ID
	userID := msg.From.ID
	username := msg.From.UserName
	text := strings.TrimSpace(msg.Text)
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return nil
	}
	command := fields[0]
	switch {
	case command == "/new":
		title := strings.TrimSpace(strings.TrimPrefix(text, "/new"))
		if title == "" {
			return sendMessage(bot, chatID, "Ввели пустой запрос")
		}
		mu.Lock()
		countID++

		task := Task{
			ID:             countID,
			Text:           title,
			AuthorID:       userID,
			AuthorUsername: username,
		}

		tasks = append(tasks, task)
		mu.Unlock()

		return sendMessage(
			bot,
			chatID,
			fmt.Sprintf("Задача %q создана, id=%d", task.Text, task.ID),
		)
	case command == "/tasks":
		var lines []string

		for _, task := range tasks {
			if task.Resolved {
				continue
			}

			lines = append(lines, formatTask(task, userID, true))
		}

		return sendMessage(bot, chatID, joinTasks(lines))
	case command == "/my":
		var lines []string

		for _, task := range tasks {
			if task.AssigneeID == userID && !task.Resolved {
				lines = append(lines, formatTask(task, userID, false))
			}
		}
		return sendMessage(bot, chatID, joinTasks(lines))
	case command == "/owner":
		var lines []string

		for _, task := range tasks {
			if task.AuthorID == userID && !task.Resolved {
				lines = append(lines, formatTask(task, userID, false))
			}
		}

		return sendMessage(bot, chatID, joinTasks(lines))
	case strings.HasPrefix(command, "/assign_"):
		id, err := parseTaskID(command, "/assign_")
		if err != nil {
			return sendMessage(bot, chatID, "Некорректный ID задачи")
		}

		index := findTaskIndex(id)
		if index == -1 {
			return sendMessage(bot, chatID, "Задача не найдена")
		}

		task := &tasks[index]
		oldAssigneeID := task.AssigneeID

		task.AssigneeID = userID
		task.AssigneeUsername = username

		if err := sendMessage(
			bot,
			chatID,
			fmt.Sprintf("Задача %q назначена на вас", task.Text),
		); err != nil {
			return err
		}

		if oldAssigneeID != 0 && oldAssigneeID != userID {
			return sendMessage(
				bot,
				oldAssigneeID,
				fmt.Sprintf(
					"Задача %q назначена на @%s",
					task.Text,
					username,
				),
			)
		}

		if oldAssigneeID == 0 && task.AuthorID != userID {
			return sendMessage(
				bot,
				task.AuthorID,
				fmt.Sprintf(
					"Задача %q назначена на @%s",
					task.Text,
					username,
				),
			)
		}

		return nil

	case strings.HasPrefix(command, "/unassign_"):
		id, err := parseTaskID(command, "/unassign_")
		if err != nil {
			return sendMessage(bot, chatID, "Некорректный ID задачи")
		}

		index := findTaskIndex(id)
		if index == -1 {
			return sendMessage(bot, chatID, "Задача не найдена")
		}

		task := &tasks[index]

		if task.AssigneeID != userID {
			return sendMessage(bot, chatID, "Задача не на вас")
		}

		task.AssigneeID = 0
		task.AssigneeUsername = ""

		if err := sendMessage(bot, chatID, "Принято"); err != nil {
			return err
		}

		if task.AuthorID != userID {
			return sendMessage(
				bot,
				task.AuthorID,
				fmt.Sprintf(
					"Задача %q осталась без исполнителя",
					task.Text,
				),
			)
		}

		return nil
	case strings.HasPrefix(command, "/resolve_"):
		id, err := parseTaskID(command, "/resolve_")
		if err != nil {
			return sendMessage(bot, chatID, "Некорректный ID задачи")
		}

		index := findTaskIndex(id)
		if index == -1 {
			return sendMessage(bot, chatID, "Задача не найдена")
		}

		task := tasks[index]

		if task.AssigneeID != userID {
			return sendMessage(bot, chatID, "Задача не на вас")
		}

		tasks = append(tasks[:index], tasks[index+1:]...)

		if err := sendMessage(
			bot,
			chatID,
			fmt.Sprintf("Задача %q выполнена", task.Text),
		); err != nil {
			return err
		}

		if task.AuthorID != userID {
			return sendMessage(
				bot,
				task.AuthorID,
				fmt.Sprintf(
					"Задача %q выполнена @%s",
					task.Text,
					task.AssigneeUsername,
				),
			)
		}

		return nil
	}

	return nil
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	_, err := bot.Send(tgbotapi.NewMessage(chatID, text))
	return err
}

func parseTaskID(command, prefix string) (int, error) {
	idText := strings.TrimPrefix(command, prefix)
	if idText == "" {
		return 0, fmt.Errorf("пустой ID задачи")
	}

	return strconv.Atoi(idText)
}

func findTaskIndex(id int) int {
	for i, task := range tasks {
		if task.ID == id {
			return i
		}
	}

	return -1
}

func joinTasks(lines []string) string {
	if len(lines) == 0 {
		return "Нет задач"
	}

	return strings.Join(lines, "\n\n")
}

func formatTask(task Task, userID int64, showAssignee bool) string {
	result := fmt.Sprintf(
		"%d. %s by @%s",
		task.ID,
		task.Text,
		task.AuthorUsername,
	)

	if showAssignee && task.AssigneeID != 0 {
		if task.AssigneeID == userID {
			result += "\nassignee: я"
		} else {
			result += "\nassignee: @" + task.AssigneeUsername
		}
	}

	if task.AssigneeID == 0 {
		result += fmt.Sprintf("\n/assign_%d", task.ID)
	} else if task.AssigneeID == userID {
		result += fmt.Sprintf(
			"\n/unassign_%d /resolve_%d",
			task.ID,
			task.ID,
		)
	}

	return result
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
