package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Syncer struct {
	config     *Config
	db         Database
	httpClient HttpClient
}

func (s *Syncer) Sync(dotFilePath string, syncType string) error {

	Info("Sync starting...")

	// `cd ${dotFilePath}` command
	err := os.Chdir(dotFilePath)
	if err != nil {
		return err
	}

	// `git pull origin main` command
	err = exec.Command("git", "pull", "origin", "main").Run()
	if err != nil {
		return fmt.Errorf("sync failed [git pull command failed with: %s]", err)
	}

	// `stow .` command
	err = exec.Command("stow", ".").Run()
	if err != nil {
		return fmt.Errorf("sync failed [stow execution failed: %v]", err)
	}

	// update database
	// get remote commit
	remoteCommits, err := s.httpClient.GetCommits()
	if err != nil {
		return err
	}

	headCommit := remoteCommits[0]

	// update or create resource
	commit := &Commit{
		Id:   headCommit.Sha,
		Time: "",
	}

	syncStash := SyncStash{
		Commit: commit,
		Type:   syncType,
		Time:   time.Now().Format(time.RFC3339),
	}

	err = s.db.Create(&syncStash)
	if err != nil {
		return err
	}

	Info("Sync completed...")
	return nil
}
