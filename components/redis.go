package components

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/redis.v3"

	"github.com/RangelReale/osin"
)

type RedisStorage struct {
	clients   map[string]osin.Client
	authorize map[string]*osin.AuthorizeData
	access    map[string]*osin.AccessData
	refresh   map[string]string

	client *redis.Client
}

func NewRedisStorage(client *redis.Client) *RedisStorage {
	r := &RedisStorage{
		clients:   make(map[string]osin.Client),
		authorize: make(map[string]*osin.AuthorizeData),
		access:    make(map[string]*osin.AccessData),
		refresh:   make(map[string]string),
		client:    client,
	}

	r.clients["1234"] = &osin.DefaultClient{
		Id:          "1234",
		Secret:      "aabbccdd",
		RedirectUri: "http://localhost:14000/appauth",
	}

	return r
}

func (s *RedisStorage) Clone() osin.Storage {
	return s
}

func (s *RedisStorage) Close() {
}

func (s *RedisStorage) GetClient(id string) (osin.Client, error) {
	fmt.Printf("GetClient: %s\n", id)
	if c, ok := s.clients[id]; ok {
		return c, nil
	}
	return nil, errors.New("Client not found")
}

func (s *RedisStorage) SetClient(id string, client osin.Client) error {
	fmt.Printf("SetClient: %s\n", id)
	s.clients[id] = client
	return nil
}

func (s *RedisStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	fmt.Printf("SaveAuthorize: %s\n", data.Code)

	binary, _ := data.CreatedAt.MarshalBinary()
	s.client.HMSet(data.Code,
		"client", data.Client.GetId(),
		"expires_in", strconv.Itoa(int(data.ExpiresIn)),
		"scope", data.Scope,
		"redirect_uri", data.RedirectUri,
		"state", data.State,
		"created_at", string(binary),
	).Result()
	s.client.Expire(data.Code, time.Duration(data.ExpiresIn)*time.Second).Result()
	return nil
}

func (s *RedisStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	fmt.Printf("LoadAuthorize: %s\n", code)

	d_map, err := s.client.HGetAllMap(code).Result()
	if err != nil {
		return nil, errors.New("Authorize not found")
	}

	client, _ := s.GetClient(d_map["client"])
	expires_in, _ := strconv.Atoi(d_map["expires_in"])
	created_at := new(time.Time)
	created_at.UnmarshalBinary([]byte(d_map["created_at"]))

	d := &osin.AuthorizeData{
		Client:      client,
		ExpiresIn:   int32(expires_in),
		Scope:       d_map["scope"],
		RedirectUri: d_map["redirect_uri"],
		CreatedAt:   *created_at,
	}
	return d, nil
}

func (s *RedisStorage) RemoveAuthorize(code string) error {
	fmt.Printf("RemoveAuthorize: %s\n", code)
	s.client.Del(code).Result()
	return nil
}

func (s *RedisStorage) SaveAccess(data *osin.AccessData) error {
	fmt.Printf("SaveAccess: %s\n", data.AccessToken)
	s.access[data.AccessToken] = data
	if data.RefreshToken != "" {
		s.refresh[data.RefreshToken] = data.AccessToken
	}

	binary, _ := data.CreatedAt.MarshalBinary()
	s.client.HMSet(data.AccessToken,
		"client", data.Client.GetId(),
		"expires_in", strconv.Itoa(int(data.ExpiresIn)),
		"scope", data.Scope,
		"redirect_uri", data.RedirectUri,
		"created_at", string(binary),
	).Result()
	s.client.Expire(data.AccessToken, time.Duration(data.ExpiresIn)*time.Second).Result()

	if data.RefreshToken != "" {
		s.client.Set(data.RefreshToken, data.AccessToken, time.Duration(data.ExpiresIn)*time.Second).Result()
	}

	return nil
}

func (s *RedisStorage) LoadAccess(code string) (*osin.AccessData, error) {
	fmt.Printf("LoadAccess: %s\n", code)

	d_map, err := s.client.HGetAllMap(code).Result()
	if err != nil {
		return nil, errors.New("Authorize not found")
	}

	client, _ := s.GetClient(d_map["client"])
	expires_in, _ := strconv.Atoi(d_map["expires_in"])
	created_at := new(time.Time)
	created_at.UnmarshalBinary([]byte(d_map["created_at"]))

	d := &osin.AccessData{
		Client:      client,
		ExpiresIn:   int32(expires_in),
		Scope:       d_map["scope"],
		RedirectUri: d_map["redirect_uri"],
		CreatedAt:   *created_at,
	}
	return d, nil
}

func (s *RedisStorage) RemoveAccess(code string) error {
	fmt.Printf("RemoveAccess: %s\n", code)
	s.client.Del(code).Result()
	return nil
}

func (s *RedisStorage) LoadRefresh(code string) (*osin.AccessData, error) {
	fmt.Printf("LoadRefresh: %s\n", code)
	d, err := s.client.Get(code).Result()
	if err != nil {
		return nil, errors.New("Authorize not found")
	}
	return s.LoadAccess(d)
}

func (s *RedisStorage) RemoveRefresh(code string) error {
	fmt.Printf("RemoveRefresh: %s\n", code)
	s.client.Del(code).Result()
	return nil
}
