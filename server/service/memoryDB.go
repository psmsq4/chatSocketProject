package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisUser struct {
	UniqueID     uint64 `json:"unique_id"`
	SessionID    int32  `json:"server_id"`
	UserID       []byte `json:"user_id"`
	UserIDLength int8   `json:"user_id_length"`
	IsAuth       bool   `json:"is_auth"`
	ChatRoomID   int16  `json:"chatroom_id"`
}

type RedisConfig struct {
	Addr             string `json:"addr"`
	Password         string `json:"password"`
	DB               int    `json:"db"`
	SessionUniqueKey string `json:"session_unique_key"`
	UserPrefix       string `json:"user_prefix"`
}

var _redisClient *redis.Client
var _ctx context.Context
var _redisConfig RedisConfig

func InitRedis(redisConfig RedisConfig) error {
	_redisConfig = redisConfig
	_redisClient = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	_ctx = context.Background()

	_ = _redisClient
	return nil
}

/*
Redis에 저장된 세션 유니크값을 Lua Script로 가져옴
*/
func GetUniqueSessionId() uint64 {
	script := redis.NewScript(`
		local key = KEYS[1]
		local value = redis.call("EXISTS", key)
		if value == 0 then
			redis.call("SET", key, 1)
		else 
			redis.call("INCR", key)
		end

		local ret = redis.call("GET", key)
		return ret
	`)

	id, err := script.Run(_ctx, _redisClient, []string{_redisConfig.SessionUniqueKey}).Uint64()
	if err != nil {
		fmt.Println(err)
		return 0
	}

	return id
}

func UserPrefix() string {
	return _redisConfig.UserPrefix
}

func set(key string, value interface{}) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}

	redisCmd := _redisClient.Set(_ctx, key, p, 0)
	return redisCmd.Err()
}

func get(key string, dest interface{}) error {
	stringCmd := _redisClient.Get(_ctx, key)
	if stringCmd.Err() != nil {
		return stringCmd.Err()
	}

	p := []byte(stringCmd.Val())
	return json.Unmarshal(p, dest)
}

func del(key string) error {
	intCmd := _redisClient.Del(_ctx, key)
	return intCmd.Err()
}

/*
Redis에 유저정보를 저장한다.
*/
func StoreUserInfo(sessionUniqueID uint64, SessionId int32, userID []byte, IsAuth bool) error {
	redisUser := RedisUser{
		UniqueID:     sessionUniqueID,
		SessionID:    SessionId,
		UserID:       userID,
		UserIDLength: int8(len(userID)),
		IsAuth:       IsAuth,
	}

	userPrefix := UserPrefix()
	key := fmt.Sprintf("%s%d", userPrefix, sessionUniqueID)
	return set(key, redisUser)
}

/*
Redis로부터 유저정보를 불러온다.
*/
func LoadUserInfo(networkUniqueID uint64, serverSessionId int32) RedisUser {
	retrieved_UID_to_RedisUser := new(RedisUser)
	userPrefix := UserPrefix()
	key := fmt.Sprintf("%s%d", userPrefix, networkUniqueID)
	get(key, &retrieved_UID_to_RedisUser)

	return *retrieved_UID_to_RedisUser
}

/*
Redis에서 유저정보를 지운다.
*/
func RemoveUserInfo(sessionUniqueId uint64) error {
	userPrefix := UserPrefix()
	key := fmt.Sprintf("%s%d", userPrefix, sessionUniqueId)
	return del(key)
}

// func InsertCIDToUser(chatRoomID int16, sessionUniqueID uint64, sessionID int32) {
// 	script := redis.NewScript(`
// 		local value = redis.call("EXISTS", "user_valuelist")
// 		if value == 1 then
// 			for i = 1, redis.call("LLEN", "user_valuelist") do
// 				redis.call("LPOP", "user_valuelist")
// 			end
// 		end
// 		local user_keylist = redis.call("KEYS", 'user_*')
// 		for index, user in pairs(user_keylist) do
// 			local user_value = redis.call("GET", user)
// 			redis.call("LPUSH", "user_valuelist", user_value)
// 		end
// 		return redis.call("LRANGE", "user_valuelist", 0, -1)
// 	`) // key 목록 조회 -> key의 value 순차적으로 조회하여 반환 -> 메인프로그램에서 반환된 value값을 Unmarsal하여 chatRoomID값 비교

// 	result := script.Run(_ctx, _redisClient, []string{""})

// 	data := make(map[string]interface{})
// 	raw_data, _ := result.StringSlice()
// 	for _, test := range raw_data {
// 		json.Unmarshal([]byte(test), &data)
// 		// dec, _ := b64.RawStdEncoding.DecodeString(data["user_id"].(string))
// 		// fmt.Println(string(dec))
// 		fmt.Println("REDIS: ", uint64(data["unique_id"].(float64)))
// 		fmt.Println("SESSION: ", sessionUniqueID)
// 		if uint64(data["unique_id"].(float64)) == sessionUniqueID {
// 			userID, _ := base64.RawStdEncoding.DecodeString(data["user_id"].(string))
// 			sessionRevision := RedisUser{
// 				UniqueID:     sessionUniqueID,
// 				SessionID:    sessionID,
// 				UserID:       userID,
// 				UserIDLength: int8(len(userID)),
// 				IsAuth:       true,
// 				ChatRoomID:   chatRoomID,
// 			}
// 			key := fmt.Sprintf("%s%d", UserPrefix(), sessionUniqueID)
// 			err := set(key, sessionRevision)
// 			fmt.Println(err)
// 		}
// 	}
// }

func InsertCIDToUser(sessionUniqueID uint64, sessionID int32, chatRoomID int16) {
	revisedSession := new(RedisUser)
	key := fmt.Sprintf("%s%d", UserPrefix(), sessionUniqueID)
	get(key, &revisedSession)
	revisedSession.ChatRoomID = chatRoomID
	set(key, revisedSession)
}

func RetrieveUsersFromCID(chatRoomID int16) []int16 {
	script := redis.NewScript(`
		local value = redis.call("EXISTS", "user_valuelist")
		if value == 1 then
			for i = 1, redis.call("LLEN", "user_valuelist") do
				redis.call("LPOP", "user_valuelist")
			end
		end
		local user_keylist = redis.call("KEYS", 'user_*')
		for index, user in pairs(user_keylist) do
			local user_value = redis.call("GET", user)
			redis.call("LPUSH", "user_valuelist", user_value)
		end
		return redis.call("LRANGE", "user_valuelist", 0, -1)
	`) // key 목록 조회 -> key의 value 순차적으로 조회하여 반환 -> 메인프로그램에서 반환된 value값을 Unmarsal하여 chatRoomID값 비교

	result := script.Run(_ctx, _redisClient, []string{""})

	retrieve_users := make([]int16, 0, 10)

	data := make(map[string]interface{})
	raw_data, _ := result.StringSlice()
	for _, test := range raw_data {
		json.Unmarshal([]byte(test), &data)
		// dec, _ := b64.RawStdEncoding.DecodeString(data["user_id"].(string))
		// fmt.Println(string(dec))
		if int16(data["chatroom_id"].(float64)) == chatRoomID {
			retrieve_users = append(retrieve_users, int16(data["unique_id"].(float64)))
		}
	}
	fmt.Println(retrieve_users)

	return retrieve_users
}
