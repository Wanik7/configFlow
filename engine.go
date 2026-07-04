package main

import (
	"fmt"
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
		return fmt.Errorf("no such storage: %w", storageName)
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
