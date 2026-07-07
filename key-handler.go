package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type KeyHandler struct{}

func NewKeyHandler() *KeyHandler {
	return &KeyHandler{}
}

func (kh *KeyHandler) GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (kh *KeyHandler) EncodeKey(key []byte) string {
	return hex.EncodeToString(key)
}

func (kh *KeyHandler) DecodeKey(key string) ([32]byte, error) {
	decodedKey, err := hex.DecodeString(key)
	if err != nil {
		return [32]byte{}, fmt.Errorf("invalid hex key provided: %w", err)
	}

	if len(decodedKey) != 32 {
		return [32]byte{}, fmt.Errorf("key must be exactly 32 bytes (64 hex characters)")
	}

	var result [32]byte
	copy(result[:], decodedKey)

	return result, nil
}
