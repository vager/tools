package main

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

// TOKEN_RANDOM_LEN length
const tokenLen int = 24
const tokenExpireTime = 6 * 1000 * 1000 * 1000

var tokenLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type serverToken struct {
	key       string
	operation string
	now       time.Time
}

// FileServer struct
type FileServer struct {
	fStore *FileStore
	tokens map[string]serverToken
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewFileServer func
func NewFileServer(path string) *FileServer {
	if path == "" {
		return nil
	}
	fileStore := NewFileStore(path)

	return &FileServer{fStore: fileStore, tokens: map[string]serverToken{}}
}

// Token Generate Token
func (fServer *FileServer) Token(key string, op string) (string, error) {
	//
	tokenID := make([]byte, tokenLen)
	for i := 0; i < tokenLen; i++ {
		tokenID[i] = tokenLetters[rand.Intn(52)]
	}
	tokenStr := string(tokenID)
	log.Printf("Gen token: %s %s %s", key, op, tokenStr)
	if _, exists := fServer.tokens[tokenStr]; exists {
		log.Printf("token key[%s] already exist", tokenStr)
	}
	fServer.tokens[tokenStr] = serverToken{operation: op, now: time.Now()}
	return tokenStr, nil
}

// validate token
func (fServer *FileServer) validateToken(key, op string) (string, bool) {
	valid := false
	var msg string
	if token, exists := fServer.tokens[key]; exists == true {
		if time.Now().Sub(token.now) > tokenExpireTime {
			msg = "expired token"
		} else if token.operation != op {
			msg = "invalid operation"
		} else {
			msg = "valid"
			valid = true
		}
		delete(fServer.tokens, key)
	}

	return msg, valid
}

// Upload file
func (fServer *FileServer) Upload(token, key string, value []byte) (string, error) {
	if msg, valid := fServer.validateToken(token, "put"); valid == false {
		log.Println("invalid token: ", msg)
		return msg, errors.New("invalid token")
	}
	log.Printf("FileServer.Upload: %s %s", token, key)
	err := fServer.fStore.WriteDb(key, value)
	if err != nil {
		log.Println("WriteDb failed:", token, key, err)
		return "Write DB failed", err
	}
	return "Upload success", nil
}

// Download file
func (fServer *FileServer) Download(token, key string) ([]byte, error) {
	if msg, valid := fServer.validateToken(token, "get"); valid == false {
		log.Println("invalid token: ", msg)
		return []byte(msg), nil
	}

	log.Printf("FileServer.Download: %s %s", token, key)
	data, err := fServer.fStore.ReadDb(key)
	if err != nil {
		log.Println("ReadDb failed:", token, key, err)
		return []byte("Read DB error"), err
	}

	return data, nil
}
