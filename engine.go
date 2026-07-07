package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

type SyncEngine struct {
	storages map[string]Storage
}

func NewSyncEngine() *SyncEngine {
	return &SyncEngine{storages: make(map[string]Storage)}
}

func (se *SyncEngine) RegisterStorage(s Storage) {
	se.storages[s.GetName()] = s
}

func (se *SyncEngine) Backup(sourceFile, configName, storageName string) error {
	storage, ok := se.storages[storageName]
	if !ok {
		return fmt.Errorf("no such storage: %s", storageName)
	}

	targetFile, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("could not open target file: %w", err)
	}

	defer CloseSafe(targetFile)

	err = storage.Store(configName, targetFile)
	if err != nil {
		return fmt.Errorf("could not backup target file: %w", err)
	}

	return nil
}

func (se *SyncEngine) Restore(backupFile, destFile, storageName string) error {
	storage, ok := se.storages[storageName]
	if !ok {
		return fmt.Errorf("no such storage: %s", storageName)
	}

	outFile, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("could not create destination file: %w", err)
	}
	defer CloseSafe(outFile)

	if err := storage.Retrieve(backupFile, outFile); err != nil {
		return fmt.Errorf("could not restore backup: %w", err)
	}

	return nil
}

func CloseSafe(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("could not close resource: %v", err)
	}
}
