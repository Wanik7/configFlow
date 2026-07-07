package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
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

type BackupJob struct {
	SourceData  string
	BackupDest  string
	StorageType string // archive or secure
}

func (se *SyncEngine) worker(id uint, jobs <-chan BackupJob, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		storage, ok := se.storages[job.StorageType]
		if !ok {
			results <- fmt.Sprintf("[Worker %d] No such storage: %s", id, job.StorageType)
			continue
		}

		targetFile, err := os.Open(job.SourceData)
		if err != nil {
			results <- fmt.Sprintf("[Worker %d] Could not open target file: %v", id, err)
			continue
		}

		err = storage.Store(job.BackupDest, targetFile)
		CloseSafe(targetFile)
		if err != nil {
			results <- fmt.Sprintf("[Worker %d] Could not backup target file: %v", id, err)
			continue
		} else {
			results <- fmt.Sprintf("[Worker %d] Successfully backed up %s to %s [%s]",
				id, job.SourceData, job.BackupDest, job.StorageType)
		}
	}
}

func (se *SyncEngine) ParallelBackup(jobList []BackupJob, maxWorkers int) {
	jobsChan := make(chan BackupJob, len(jobList))
	results := make(chan string, len(jobList))

	var wg sync.WaitGroup

	for id := 1; id <= maxWorkers; id++ {
		wg.Add(1)
		go se.worker(uint(id), jobsChan, results, &wg)
	}

	for _, job := range jobList {
		jobsChan <- job
	}

	close(jobsChan)

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Printf("--- Launch %d workers for %d jobs ---\n", maxWorkers, len(jobList))

	for res := range results {
		fmt.Println(res)
	}

	fmt.Println("\n--- Done ---")
}

func CloseSafe(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("could not close resource: %v", err)
	}
}
