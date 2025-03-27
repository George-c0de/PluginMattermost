package handler

import (
	"PluginMattermost/internal/dto"
	"PluginMattermost/internal/logger"
	"PluginMattermost/internal/storage"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
)

type PollHandler struct {
	store *storage.TarantoolStorage
	log   *logger.Logger
}

func NewPollHandler(store *storage.TarantoolStorage, log *logger.Logger) *PollHandler {
	return &PollHandler{store: store, log: log}
}

// respondWithJSON отправляет JSON-ответ с указанным статусом
func (h *PollHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.log.Error("Failed to write JSON response:", err)
	}
}

// HealthHandler обеспечивает базовую проверку "живости" сервиса.
func (h *PollHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		h.log.Error("health check write error", err)
	}
}

// respondWithError отправляет JSON-ответ с ошибкой
func (h *PollHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Received CreatePoll request")

	var req dto.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}
	question := req.Question
	options := req.Options

	if question == "" || len(options) == 0 {
		h.respondWithError(w, http.StatusBadRequest, "invalid poll data")
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
		h.respondWithError(w, http.StatusInternalServerError, "failed to create poll")
		return
	}

	// Возвращаем ID опроса и варианты
	resp := map[string]interface{}{
		"poll_id":  pollID,
		"question": question,
		"options":  options,
	}
	h.respondWithJSON(w, http.StatusCreated, resp)
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Received Vote request")

	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "user_id not found")
		return
	}
	var data dto.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.respondWithError(w, http.StatusUnprocessableEntity, "invalid json")
		return
	}
	if data.PoolID == 0 {
		h.respondWithError(w, http.StatusUnprocessableEntity, "poll_id is required")
		return
	}
	if data.Option == "" {
		h.respondWithError(w, http.StatusUnprocessableEntity, "option is required")
		return
	}

	pollID := data.PoolID
	option := data.Option

	poll, err := h.store.GetPoll(pollID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "poll not found")
		return
	}
	if poll.Closed {
		h.respondWithError(w, http.StatusBadRequest, "poll is already closed")
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
		h.respondWithError(w, http.StatusBadRequest, "invalid option")
		return
	}

	// Если пользователь уже голосовал, отклоняем повторное голосование
	if poll.UserVotes != nil {
		if _, exists := poll.UserVotes[userID]; exists {
			h.respondWithError(w, http.StatusBadRequest, "user already voted")
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
		h.respondWithError(w, http.StatusInternalServerError, "failed to update poll")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "vote accepted"})
}

func (h *PollHandler) Results(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Received Results request")

	pollIDStr := r.URL.Query().Get("poll_id")

	if pollIDStr == "" {
		h.respondWithError(w, http.StatusUnprocessableEntity, "missing poll_id")
		return
	}

	pollID, err := strconv.ParseUint(pollIDStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid poll_id")
		return
	}

	poll, err := h.store.GetPoll(pollID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "poll not found")
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

	h.respondWithJSON(w, http.StatusOK, resp)
}

// EndPoll завершает голосование (ставит флаг Closed = true)
func (h *PollHandler) EndPoll(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Received EndPoll request")

	data := struct {
		PollID uint64 `json:"poll_id"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if data.PollID == 0 {
		h.respondWithError(w, http.StatusBadRequest, "poll_id is required")
		return
	}
	err := h.store.UpdatePoll(data.PollID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "internal error")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "poll closed"})
}

func (h *PollHandler) DeleteVote(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Received Vote request")
	// Получаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "user_id not found")
		return
	}
	var data dto.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if data.PoolID == 0 {
		h.respondWithError(w, http.StatusBadRequest, "poll_id is required")
		return
	}
	if data.Option == "" {
		h.respondWithError(w, http.StatusBadRequest, "option is required")
		return
	}

	pollID := data.PoolID
	option := data.Option

	poll, err := h.store.GetPoll(pollID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "poll not found")
		return
	}
	if poll.Closed {
		h.respondWithError(w, http.StatusBadRequest, "poll is already closed")
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
		h.respondWithError(w, http.StatusBadRequest, "invalid option")
		return
	}
	// Проверяем, что голос принадлежит пользователю
	votedOption, exists := poll.UserVotes[userID]
	if !exists || votedOption != data.Option {
		h.respondWithError(w, http.StatusBadRequest, "vote for this user not found")
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
		h.respondWithError(w, http.StatusInternalServerError, "failed to update poll")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "vote removed"})
}
