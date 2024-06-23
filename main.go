package main

import (
	"bufio"
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
		rootCmd                 = cobra.Command{}
		defaultDotFileDirectory = func() string {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Println("Unable to access home directory:", err)
				os.Exit(1)
			}

			return path.Join(homeDir, "dotfiles")
		}()
	)

	port := rootCmd.Flags().StringP("port", "p", "3000", "HTTP port to run on")
	webhookUrl := rootCmd.Flags().StringP("webhook", "w", "https://smee.io/awFay3gs7LCGYe2", "git webhook url")
	dotFilePath := rootCmd.Flags().StringP("dotfile-path", "d", defaultDotFileDirectory, "path to dotfile directory")

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	go func() {
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
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {

		event := request.Header.Get("x-github-event")
		if event == "push" {
			log.Println("push event detected")

			err := os.Chdir(*dotFilePath)
			if err != nil {
				log.Println(err)
			}

			err = exec.Command("git", "pull", "origin", "main").Run()
			if err != nil {
				log.Printf("git repository failed to pull [%s]\n", err)
			} else {
				log.Println("git repository pulled successfully")
			}

			// run stow
			err = exec.Command("stow", ".").Run()
			if err != nil {
				log.Println("stow execution failed: ", err)
			} else {
				log.Println("stow execution succeeded")
			}
		}
	})

	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
