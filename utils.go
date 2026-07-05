package main

import (
	"crypto/rand"
	"io"
	"log"
)

/*
CloseSafe is a custom function that uses the Closer interface and
performs simple error handling after the Close method
*/

func CloseSafe(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("could not close resource: %v", err)
	}
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
