package dto

type Poll struct {
	ID        uint64            `json:"id"`
	Question  string            `json:"question"`
	Options   []string          `json:"options"`
	Votes     map[string]uint64 `json:"votes"`                  // ключ — вариант, значение — кол-во голосов
	Closed    bool              `json:"closed" default:"false"` // Новый флаг
	UserVotes map[string]string `json:"user_votes"`             // Отображение: userID -> выбранный вариант
}

// CreatePollRequest Пример структуры запроса на создание опроса
type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type VoteRequest struct {
	PoolID uint64 `json:"poll_id"`
	Option string `json:"option"`
}
