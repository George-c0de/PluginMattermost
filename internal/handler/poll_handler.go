package handler

import (
	"PluginMattermost/internal/storage"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PollHandler struct {
	store *storage.TarantoolStorage
}

func NewPollHandler(store *storage.TarantoolStorage) *PollHandler {
	return &PollHandler{store: store}
}

// Пример структуры запроса на создание опроса
type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	// Логируем
	log.Println("Received CreatePoll request")

	// Предположим, что данные могут приходить либо JSON-ом, либо формой (slash-команда)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}

	question := r.FormValue("question")
	options := r.Form["options"]

	// Если нет question, пробуем считать JSON
	if question == "" && len(options) == 0 {
		var req CreatePollRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		question = req.Question
		options = req.Options
	}

	if question == "" || len(options) == 0 {
		http.Error(w, "invalid poll data", http.StatusBadRequest)
		return
	}

	// Генерируем случайный ID (в реальном проекте лучше использовать атомарный счётчик в Tarantool)
	rand.Seed(time.Now().UnixNano())
	pollID := uint64(rand.Intn(9999999))

	// Создаём структуру Poll
	poll := &storage.Poll{
		ID:       pollID,
		Question: question,
		Options:  options,
		Votes:    make(map[string]uint64),
	}

	if err := h.store.CreatePoll(poll); err != nil {
		http.Error(w, "failed to create poll", http.StatusInternalServerError)
		return
	}

	// Возвращаем ID опроса и варианты
	resp := map[string]interface{}{
		"poll_id":  pollID,
		"question": question,
		"options":  options,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Vote request")

	if err := json.NewDecoder(r.Body).Decode(&r); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	pollIDStr := r.FormValue("poll_id")
	option := r.FormValue("option")

	// Для slash-команд, где всё может быть одной строкой, например: /poll vote 123 "Go"
	// Можно парсить r.FormValue("text") и разбирать вручную. Примерно так:
	if pollIDStr == "" || option == "" {
		text := r.FormValue("text")
		if text != "" {
			parts := strings.SplitN(text, " ", 2)
			if len(parts) == 2 {
				pollIDStr = parts[0]
				option = strings.Trim(parts[1], `" `)
			}
		}
	}

	if pollIDStr == "" || option == "" {
		http.Error(w, "missing poll_id or option", http.StatusBadRequest)
		return
	}

	pollID, err := strconv.ParseUint(pollIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid poll_id", http.StatusBadRequest)
		return
	}

	poll, err := h.store.GetPoll(pollID)
	if err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}

	// Проверяем, есть ли такой вариант в poll.Options
	validOption := false
	for _, o := range poll.Options {
		if o == option {
			validOption = true
			break
		}
	}
	if !validOption {
		http.Error(w, "invalid option", http.StatusBadRequest)
		return
	}

	// Увеличиваем счётчик голосов
	poll.Votes[option] = poll.Votes[option] + 1

	// Сохраняем
	if err := h.store.UpdatePoll(poll); err != nil {
		http.Error(w, "failed to update poll", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vote accepted"))
}

func (h *PollHandler) Results(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Results request")

	pollIDStr := r.URL.Query().Get("poll_id")
	if pollIDStr == "" {
		// Или парсим как slash-команду
		if err := r.ParseForm(); err == nil {
			pollIDStr = r.FormValue("poll_id")
			if pollIDStr == "" {
				pollIDStr = r.FormValue("text") // /poll results 123
			}
		}
	}

	if pollIDStr == "" {
		http.Error(w, "missing poll_id", http.StatusBadRequest)
		return
	}

	pollID, err := strconv.ParseUint(pollIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid poll_id", http.StatusBadRequest)
		return
	}

	poll, err := h.store.GetPoll(pollID)
	if err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"poll_id":  poll.ID,
		"question": poll.Question,
		"options":  poll.Options,
		"votes":    poll.Votes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
