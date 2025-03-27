package storage

import (
	"PluginMattermost/internal/dto"
	"PluginMattermost/internal/logger"
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"time"
)

type TarantoolStorage struct {
	conn *tarantool.Connection
	log  *logger.Logger
}

func MustNewTarantoolStorage(host string, port int, log *logger.Logger) *TarantoolStorage {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address: fmt.Sprintf("%s:%d", host, port),
		// TODO(George): Добавить пользователя и пароль
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		panic(err)
	}

	return &TarantoolStorage{conn: conn, log: log}
}

// CreatePoll Создать новый опрос
func (t *TarantoolStorage) CreatePoll(poll *dto.Poll) error {
	_, err := t.conn.Do(
		tarantool.NewInsertRequest("polls").Tuple(
			[]interface{}{poll.ID, poll.Question, poll.Options, poll.Votes, poll.Closed, poll.UserVotes},
		),
	).Get()
	if err != nil {
		t.log.Error("Got an error:", err)
	}

	return err
}

// GetPoll Получить опрос по ID
func (t *TarantoolStorage) GetPoll(id uint64) (*dto.Poll, error) {
	data, err := t.conn.Do(
		tarantool.NewSelectRequest("polls").Iterator(tarantool.IterEq).Key([]interface{}{id}),
	).Get()

	if err != nil {
		t.log.Error("Got an error:", err)
		return nil, err
	}

	if len(data) == 0 {
		t.log.Info("poll not found")
		return nil, fmt.Errorf("poll not found")
	}
	tuple := data[0].([]interface{})
	poll := dto.Poll{
		ID:        tuple[0].(uint64),
		Question:  tuple[1].(string),
		Options:   toStringSlice(tuple[2]),
		Votes:     toMapStringUint64(tuple[3]),
		Closed:    tuple[4].(bool),
		UserVotes: toMapStringString(tuple[5]),
	}
	return &poll, nil
}

// ReplacePoll Обновить данные опроса
func (t *TarantoolStorage) ReplacePoll(poll *dto.Poll) error {
	_, err := t.conn.Do(
		tarantool.NewReplaceRequest("polls").Tuple(
			[]interface{}{poll.ID, poll.Question, poll.Options, poll.Votes, poll.Closed, poll.UserVotes},
		),
	).Get()
	if err != nil {
		t.log.Error("Got an error:", err)
	}
	return err
}

func (t *TarantoolStorage) UpdatePoll(pollId uint64) error {
	_, err := t.conn.Do(
		tarantool.NewUpdateRequest("polls").Key(tarantool.UintKey{I: uint(pollId)}).Operations(
			tarantool.NewOperations().Assign(4, true),
		),
	).Get()
	if err != nil {
		t.log.Error("Got an error:", err)
	}
	return err
}

// Вспомогательные функции для приведения типов
func toStringSlice(val interface{}) []string {
	arr, ok := val.([]interface{})
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		result = append(result, v.(string))
	}
	return result
}

func toMapStringUint64(val interface{}) map[string]uint64 {
	m, ok := val.(map[interface{}]interface{})
	if !ok {
		return map[string]uint64{}
	}
	result := make(map[string]uint64)
	for k, v := range m {
		keyStr := k.(string)
		valUint := v.(uint64)
		result[keyStr] = valUint
	}
	return result
}
func toMapStringString(val interface{}) map[string]string {
	m, ok := val.(map[interface{}]interface{})
	if !ok {
		return map[string]string{}
	}
	result := make(map[string]string)
	for k, v := range m {
		keyStr := k.(string)
		valUint := v.(string)
		result[keyStr] = valUint
	}
	return result
}
