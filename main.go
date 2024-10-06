package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/spf13/cobra"
)

func main() {

	var (
		db                      = InitDB()
		httpClient              = NewHttpClient()
		rootCmd                 = cobra.Command{}
		defaultDotFileDirectory = func() string {
			homeDir, err := os.UserConfigDir()
			if err != nil {
				Error("Unable to access home directory:", err.Error())
				os.Exit(1)
			}

			return path.Join(homeDir, "dotfiles")
		}()
	)

	port := rootCmd.Flags().StringP("port", "p", "3000", "HTTP port to run on")
	webhookUrl := rootCmd.Flags().StringP("webhook", "w", "https://smee.io/awFay3gs7LCGYe2", "git webhook url")
	dotFilePath := rootCmd.Flags().StringP("dotfile-path", "d", defaultDotFileDirectory, "path to dotfile directory")

	if err := rootCmd.Execute(); err != nil {
		Error(err.Error())
		os.Exit(1)
	}

	config := &Config{
		DotfilePath: *dotFilePath,
		WebHook:     *webhookUrl,
		Port:        *port,
	}

	syncer := &Syncer{config, db, httpClient}

	syncHandler := NewSyncHandler(syncer, db, httpClient)

	// first time sync
	err := syncer.Sync(*dotFilePath, "Automatic")
	if err != nil {
		Error("could not perform first start-up sync: ", err.Error())
	}

	go func() {

		cmd := exec.Command("smee", "--url", config.WebHook, "--path", "/webhook", "--port", config.Port)
		stdOutput, err := cmd.StdoutPipe()
		stdErr, err := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		if stdErr != nil {
			buf := bufio.NewReader(stdErr)
			line, _ := buf.ReadString('\n')
			Error(line)
			return
		}

		cmdOutput := io.MultiReader(stdOutput, stdErr)
		bufOutput := bufio.NewReader(cmdOutput)
		line, err := bufOutput.ReadString('\n')

		for err == nil {
			Info(line)
			line, err = bufOutput.ReadString('\n')
		}
	}()

	time.Sleep(3 * time.Second) // wait for smee server startup

	Info("webhook server forwarded successfully from", config.WebHook, "to port", config.Port)

	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("/sync", syncHandler.Sync)
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {

		var commit GitWebHookCommitResponse

		err := json.NewDecoder(request.Body).Decode(&commit)
		if err != nil {
			Info(err.Error())
		}

		event := request.Header.Get("x-github-event")
		if event == "push" {
			Info("changes detected...")

			err := syncer.Sync(*dotFilePath, "Automatic")
			if err != nil {
				Info("error syncing on path:", *dotFilePath, err.Error())
			} else {
				t := &Commit{
					Id:   commit.HeadCommit.Id,
					Time: "",
				}

				syncStash := &SyncStash{
					Commit: t,
					Type:   "Automatic",
					Time:   time.Now().UTC().Format(time.RFC3339),
				}

				_ = db.Create(syncStash)
			}
		}
	})

	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
