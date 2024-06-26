package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func SyncExec(dotFilePath string) error {

	log.Println("Sync starting...")

	// `cd ${dotFilePath}` command
	err := os.Chdir(dotFilePath)
	if err != nil {
		return err
	}

	// `git pull origin main` command
	err = exec.Command("git", "pull", "origin", "main").Run()
	if err != nil {
		return fmt.Errorf("git repository failed to pull [%s]\n", err)
	}

	// `stow .` command
	err = exec.Command("stow", ".").Run()
	if err != nil {
		return fmt.Errorf("stow execution failed: %v", err)
	}

	log.Println("Sync completed...")

	return nil
}
