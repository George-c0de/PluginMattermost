package storage

import (
	"PluginMattermost/internal/dto"
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"log"
	"time"
)

type TarantoolStorage struct {
	conn *tarantool.Connection
}

func NewTarantoolStorage(host string, port int) (*TarantoolStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address: fmt.Sprintf("%s:%d", host, port),
		//User:     "sampleuser",
		//Password: "123456",
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		fmt.Println("Connection refused:", err)
		panic(err)
	}

	return &TarantoolStorage{conn: conn}, nil
}

// CreatePoll Создать новый опрос
func (t *TarantoolStorage) CreatePoll(poll *dto.Poll) error {
	var future *tarantool.Future
	request := tarantool.NewInsertRequest("polls").Tuple([]interface{}{poll.ID, poll.Question, poll.Options, poll.Votes})
	future = t.conn.Do(request)
	result, err := future.Get()
	if err != nil {
		log.Println("Got an error:", err)
	} else {
		log.Println(result)
	}
	return err
}

// GetPoll Получить опрос по ID
func (t *TarantoolStorage) GetPoll(id uint64) (*dto.Poll, error) {
	data, err := t.conn.Do(
		tarantool.NewSelectRequest("polls").Iterator(tarantool.IterEq).Key([]interface{}{id}),
	).Get()

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("poll not found")
	}
	tuple := data[0].([]interface{})
	poll := dto.Poll{
		ID:       tuple[0].(uint64),
		Question: tuple[1].(string),
		Options:  toStringSlice(tuple[2]),
		Votes:    toMapStringUint64(tuple[3]),
	}
	return &poll, nil
}

// UpdatePoll Обновить данные опроса
func (t *TarantoolStorage) UpdatePoll(poll *dto.Poll) error {
	data, err := t.conn.Do(tarantool.NewReplaceRequest("polls").Tuple([]interface{}{poll.ID, poll.Question, poll.Options, poll.Votes})).Get()
	if err != nil {
		log.Println("Got an error:", err)
	}
	log.Println("Replaced tuple:", data)
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
