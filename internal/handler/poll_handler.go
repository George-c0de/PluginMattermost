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

// respondWithJSON отправляет JSON-ответ с указанным статусом
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

// respondWithError отправляет JSON-ответ с ошибкой
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	log.Println("Received CreatePoll request")

	var req dto.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}
	question := req.Question
	options := req.Options

	if question == "" || len(options) == 0 {
		respondWithError(w, http.StatusBadRequest, "invalid poll data")
		return
	}

	// Генерируем случайный ID (в реальном проекте лучше использовать атомарный счётчик в Tarantool)
	pollID := rand.Uint64()

	// Создаём структуру Poll
	poll := &dto.Poll{
		ID:        pollID,
		Question:  question,
		Options:   options,
		Votes:     make(map[string]uint64),
		Closed:    false,
		UserVotes: make(map[string]string),
	}

	if err := h.store.CreatePoll(poll); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create poll")
		return
	}

	// Возвращаем ID опроса и варианты
	resp := map[string]interface{}{
		"poll_id":  pollID,
		"question": question,
		"options":  options,
	}
	respondWithJSON(w, http.StatusCreated, resp)
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Vote request")
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		respondWithError(w, http.StatusUnauthorized, "user_id not found")
		return
	}
	log.Println(userID)
	var data dto.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Println(err)
		respondWithError(w, http.StatusUnprocessableEntity, "invalid json")
		return
	}
	if data.PoolID == 0 {
		respondWithError(w, http.StatusUnprocessableEntity, "poll_id is required")
		return
	}
	if data.Option == "" {
		respondWithError(w, http.StatusUnprocessableEntity, "option is required")
		return
	}

	pollID := data.PoolID
	option := data.Option

	poll, err := h.store.GetPoll(pollID)
	if err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}
	if poll.Closed {
		http.Error(w, "poll is already closed", http.StatusBadRequest)
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

	// Если пользователь уже голосовал, отклоняем повторное голосование
	if poll.UserVotes != nil {
		if _, exists := poll.UserVotes[userID]; exists {
			respondWithError(w, http.StatusBadRequest, "user already voted")
			return
		}
	} else {
		poll.UserVotes = make(map[string]string)
	}
	// Увеличиваем счётчик голосов
	poll.Votes[option] = poll.Votes[option] + 1
	poll.UserVotes[userID] = data.Option
	// Сохраняем
	if err := h.store.ReplacePoll(poll); err != nil {
		http.Error(w, "failed to update poll", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "vote accepted"})
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
		"poll_id":    poll.ID,
		"question":   poll.Question,
		"options":    poll.Options,
		"votes":      poll.Votes,
		"closed":     poll.Closed,
		"user_votes": poll.UserVotes,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// EndPoll завершает голосование (ставит флаг Closed = true)
func (h *PollHandler) EndPoll(w http.ResponseWriter, r *http.Request) {
	log.Println("Received EndPoll request")

	data := struct {
		PollID uint64 `json:"poll_id"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Println(err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if data.PollID == 0 {
		http.Error(w, "poll_id is required", http.StatusBadRequest)
		return
	}
	err := h.store.UpdatePoll(data.PollID)
	if err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "poll closed"})
}

func (h *PollHandler) DeleteVote(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Vote request")
	// Получаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		respondWithError(w, http.StatusUnauthorized, "user_id not found")
		return
	}
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
	if poll.Closed {
		http.Error(w, "poll is already closed", http.StatusBadRequest)
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
	// Проверяем, что голос принадлежит пользователю
	votedOption, exists := poll.UserVotes[userID]
	if !exists || votedOption != data.Option {
		respondWithError(w, http.StatusBadRequest, "vote for this user not found")
		return
	}

	// Уменьшаем счётчик голосов, если он больше нуля
	if poll.Votes[data.Option] > 0 {
		poll.Votes[data.Option]--
	}
	// Удаляем запись о голосе пользователя
	delete(poll.UserVotes, userID)

	// Сохраняем
	if err := h.store.ReplacePoll(poll); err != nil {
		http.Error(w, "failed to update poll", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "vote removed"})
}
