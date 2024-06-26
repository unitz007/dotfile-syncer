package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/haibeey/doclite"
	"github.com/spf13/cobra"
)

func main() {

	var (
		db = func() *doclite.Doclite {
			return doclite.Connect("dotfile-agent.doclite")
		}
		rootCmd                 = cobra.Command{}
		defaultDotFileDirectory = func() string {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Println("Unable to access home directory:", err)
				os.Exit(1)
			}

			return path.Join(homeDir, "dotfiles")
		}()

		syncHandler = SyncHandler{dotFilePath: defaultDotFileDirectory}
	)

	port := rootCmd.Flags().StringP("port", "p", "3000", "HTTP port to run on")
	webhookUrl := rootCmd.Flags().StringP("webhook", "w", "https://smee.io/awFay3gs7LCGYe2", "git webhook url")
	dotFilePath := rootCmd.Flags().StringP("dotfile-path", "d", defaultDotFileDirectory, "path to dotfile directory")

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	go func() {
		fmt.Println("khnbbinb")
		cmd := exec.Command("smee", "--url", *webhookUrl, "--path", "/webhook", "--port", *port)
		stdOutput, err := cmd.StdoutPipe()
		stdErr, err := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}

		if stdErr != nil {
			buf := bufio.NewReader(stdErr)
			line, _ := buf.ReadString('\n')
			log.Println("Error:", line)
			return
		}

		cmdOutput := io.MultiReader(stdOutput, stdErr)
		bufOutput := bufio.NewReader(cmdOutput)
		line, err := bufOutput.ReadString('\n')

		for err == nil {
			log.Print(line)
			line, err = bufOutput.ReadString('\n')
		}
	}()

	time.Sleep(3 * time.Second) // wait for smee server startup

	log.Println("webhook server forwarded successfully from", *webhookUrl, "to port", *port)

	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("/sync", syncHandler.Sync)
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {

		var commit GitPull

		err := json.NewDecoder(request.Body).Decode(&commit)
		if err != nil {
			fmt.Println(err)
		}

		err = db.Create(commit)
		db.data.Commit()
		fmt.Println(err)

		fmt.Println(db.data.GetCol().Name)

		event := request.Header.Get("x-github-event")
		if event == "push" {
			log.Println("push event detected...")

			err := SyncExec(*dotFilePath)
			if err != nil {
				log.Println("error syncing:", err)
				return
			}
		}
	})

	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
