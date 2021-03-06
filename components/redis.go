// Redis related structures and methods
package components

import (
	"errors"
	"strconv"
	"time"

	"gopkg.in/redis.v3"

	"github.com/RangelReale/osin"
	"github.com/quorumsco/logs"
)

// Represent an authentication session in redis
type RedisStorage struct {
	clients   map[string]osin.Client
	authorize map[string]*osin.AuthorizeData
	access    map[string]*osin.AccessData
	refresh   map[string]string

	client *redis.Client
}

// Creates a new redis storeage session
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

// Returns a client from it's id
func (s *RedisStorage) GetClient(id string) (osin.Client, error) {
	logs.Debug("GetClientDatabase: %s", id)

	if c, ok := s.clients[id]; ok {
		s.client.HMSet(id,
			"secret", c.GetSecret(),
			"redirect_uri", c.GetRedirectUri(),
		)
		return c, nil
	}
	return nil, errors.New("Client not found")
}

// Calls GetClient and returns the result or return a new client
func (s *RedisStorage) getClient(id string) (osin.Client, error) {
	logs.Debug("GetClientCache: %s", id)

	c_map, err := s.client.HGetAllMap(id).Result()
	if len(c_map) == 0 || err != nil {
		return s.GetClient(id)
	}

	client := &osin.DefaultClient{
		Id:          id,
		Secret:      c_map["secret"],
		RedirectUri: c_map["redirect_uri"],
	}

	return client, nil
}

// Creates a new client
func (s *RedisStorage) SetClient(id string, client osin.Client) error {
	logs.Debug("SetClient: %s", id)

	s.clients[id] = client
	return nil
}

// Saves a new client
func (s *RedisStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	logs.Debug("SaveAuthorize: %s", data.Code)

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
	s.client.Expire(data.Client.GetId(), time.Duration(data.ExpiresIn)*time.Second).Result()
	return nil
}

// Return a client
func (s *RedisStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	logs.Debug("LoadAuthorize: %s", code)

	d_map, err := s.client.HGetAllMap(code).Result()
	if err != nil {
		return nil, errors.New("Authorize not found")
	}

	client, err := s.getClient(d_map["client"])
	if err != nil {
		return nil, err
	}
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

// Deletes a client
func (s *RedisStorage) RemoveAuthorize(code string) error {
	logs.Debug("RemoveAuthorize: %s", code)

	s.client.Del(code).Result()
	return nil
}

// Saves a new storage session
func (s *RedisStorage) SaveAccess(data *osin.AccessData) error {
	logs.Debug("SaveAccess: %s", data.AccessToken)

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
		"user_data", data.UserData.(string),
	).Result()
	s.client.Expire(data.AccessToken, time.Duration(data.ExpiresIn)*time.Second).Result()
	// s.client.Expire(data.AccessToken, 10*time.Second).Result()

	if data.RefreshToken != "" {
		s.client.Set(data.RefreshToken, data.AccessToken, time.Duration(data.ExpiresIn)*time.Second).Result()
	}

	return nil
}

// Returns a storage session
func (s *RedisStorage) LoadAccess(code string) (*osin.AccessData, error) {
	logs.Debug("LoadAccess: %s", code)

	d_map, err := s.client.HGetAllMap(code).Result()
	if len(d_map) == 0 || err != nil {
		return nil, errors.New("Authorize not found")
	}

	client, err := s.getClient(d_map["client"])
	if err != nil {
		return nil, err
	}
	expires_in, _ := strconv.Atoi(d_map["expires_in"])
	created_at := new(time.Time)
	created_at.UnmarshalBinary([]byte(d_map["created_at"]))

	d := &osin.AccessData{
		Client:      client,
		ExpiresIn:   int32(expires_in),
		Scope:       d_map["scope"],
		RedirectUri: d_map["redirect_uri"],
		CreatedAt:   *created_at,
		UserData:    d_map["user_data"],
	}
	return d, nil
}

// Delete a storage session
func (s *RedisStorage) RemoveAccess(code string) error {
	logs.Debug("RemoveAccess: %s", code)

	s.client.Del(code).Result()
	return nil
}

func (s *RedisStorage) LoadRefresh(code string) (*osin.AccessData, error) {
	logs.Debug("LoadRefresh: %s", code)

	d, err := s.client.Get(code).Result()
	if err != nil {
		return nil, errors.New("Authorize not found")
	}
	return s.LoadAccess(d)
}

func (s *RedisStorage) RemoveRefresh(code string) error {
	logs.Debug("RemoveRefresh: %s", code)

	s.client.Del(code).Result()
	return nil
}
