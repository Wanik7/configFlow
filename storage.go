package main

import (
	"bufio"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	Store(cfgName string, data io.Reader) error
	Retrieve(cfgName string, dest io.Writer) error
	GetName() string
}

type ArchiveStorage struct {
	backupDir string
}

func (as *ArchiveStorage) Store(cfgName string, data io.Reader) error {
	pathToBackup := filepath.Join(as.backupDir, cfgName+".gz")

	backUpFile, err := os.Create(pathToBackup)
	if err != nil {
		return fmt.Errorf("could not create backup file: %w", err)
	}
	defer backUpFile.Close()

	gzipWriter := gzip.NewWriter(backUpFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, data)
	if err != nil {
		return fmt.Errorf("could not copy data: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("could not close gzip writer: %w", err)
	}

	if err := backUpFile.Close(); err != nil {
		return fmt.Errorf("could not close backup file: %w", err)
	}

	return nil
}

func (as *ArchiveStorage) Retrieve(cfgName string, dest io.Writer) error {
	pathToBackup := filepath.Join(as.backupDir, cfgName)

	backupFile, err := os.Open(pathToBackup)
	if err != nil {
		return fmt.Errorf("could not open backup file: %w", err)
	}
	defer backupFile.Close()

	gzipReader, err := gzip.NewReader(backupFile)
	if err != nil {
		return fmt.Errorf("could not create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	if _, err := io.Copy(dest, gzipReader); err != nil {
		return fmt.Errorf("could not decompress data: %w", err)
	}

	return nil
}

func (as *ArchiveStorage) GetName() string {
	return "archive"
}

type SecureStorage struct {
	backupDir string
	secretKey [32]byte
}

func (ss *SecureStorage) Store(cfgName string, data io.Reader) error {
	pathToBackup := filepath.Join(ss.backupDir, cfgName+".enc")

	backUpFile, err := os.Create(pathToBackup)
	if err != nil {
		return fmt.Errorf("could not create backup file: %w", err)
	}
	defer backUpFile.Close()

	block, err := aes.NewCipher(ss.secretKey[:])
	if err != nil {
		return fmt.Errorf("could not create AES cipher: %w", err)
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("could not generate iv: %w", err)
	}

	_, err = backUpFile.Write(iv)
	if err != nil {
		return fmt.Errorf("unexpected error with crypto-key handle: %w", err)
	}

	bufferedWriter := bufio.NewWriter(backUpFile)
	stream := cipher.NewCTR(block, iv)
	writer := cipher.StreamWriter{S: stream, W: bufferedWriter}

	_, err = io.Copy(writer, data)
	if err != nil {
		return fmt.Errorf("could not handle crypto-stream: %w", err)
	}

	if err := bufferedWriter.Flush(); err != nil {
		return fmt.Errorf("could not flush buffered writer: %w", err)
	}

	if err := backUpFile.Close(); err != nil {
		return fmt.Errorf("could not close backup file: %w", err)
	}

	return nil
}

func (ss *SecureStorage) Retrieve(cfgName string, dest io.Writer) error {
	pathToBackup := filepath.Join(ss.backupDir, cfgName)

	backupFile, err := os.Open(pathToBackup)
	if err != nil {
		return fmt.Errorf("could not open backup file: %w", err)
	}
	defer backupFile.Close()

	block, err := aes.NewCipher(ss.secretKey[:])
	if err != nil {
		return fmt.Errorf("could not create AES cipher: %w", err)
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(backupFile, iv); err != nil {
		return fmt.Errorf("could not read iv: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	reader := cipher.StreamReader{S: stream, R: backupFile}

	if _, err := io.Copy(dest, reader); err != nil {
		return fmt.Errorf("could not decrypt data: %w", err)
	}

	return nil
}

func (ss *SecureStorage) GetName() string {
	return "secure"
}
