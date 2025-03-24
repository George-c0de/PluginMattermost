package handler

import (
	"PluginMattermost/internal/dto"
	"PluginMattermost/internal/storage"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

type PollHandler struct {
	store *storage.TarantoolStorage
}

func NewPollHandler(store *storage.TarantoolStorage) *PollHandler {
	return &PollHandler{store: store}
}

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	// Логируем
	log.Println("Received CreatePoll request")

	var req dto.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	question := req.Question
	options := req.Options

	if question == "" || len(options) == 0 {
		http.Error(w, "invalid poll data", http.StatusBadRequest)
		return
	}

	// Генерируем случайный ID (в реальном проекте лучше использовать атомарный счётчик в Tarantool)
	pollID := rand.Uint64()

	// Создаём структуру Poll
	poll := &dto.Poll{
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
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Vote request")
	var data dto.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Println(err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if data.PoolID == 0 {
		http.Error(w, "poll_id is required", http.StatusBadRequest)
		return
	}
	if data.Option == "" {
		http.Error(w, "option is required", http.StatusBadRequest)
		return
	}

	pollID := data.PoolID
	option := data.Option

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
	_, err = w.Write([]byte("vote accepted"))
	if err != nil {
		return
	}
}

func (h *PollHandler) Results(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Results request")

	pollIDStr := r.URL.Query().Get("poll_id")

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
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}
