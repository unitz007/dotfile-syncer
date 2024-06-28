package main

import (
	"bufio"
	"fmt"
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
				fmt.Println("Unable to access home directory:", err)
				os.Exit(1)
			}

			return path.Join(homeDir, "dotfiles")
		}()
	)

	port := rootCmd.Flags().StringP("port", "p", "3000", "HTTP port to run on")
	webhookUrl := rootCmd.Flags().StringP("webhook", "w", "https://smee.io/awFay3gs7LCGYe2", "git webhook url")
	dotFilePath := rootCmd.Flags().StringP("dotfile-path", "d", defaultDotFileDirectory, "path to dotfile directory")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		cmd := exec.Command("smee", "--url", *webhookUrl, "--path", "/webhook", "--port", *port)
		stdOutput, err := cmd.StdoutPipe()
		stdErr, err := cmd.StderrPipe()

		if err != nil {
			fmt.Println(err)
		}

		if err := cmd.Start(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if stdErr != nil {
			buf := bufio.NewReader(stdErr)
			line, _ := buf.ReadString('\n')
			fmt.Println("Error:", line)
			return
		}

		cmdOutput := io.MultiReader(stdOutput, stdErr)
		bufOutput := bufio.NewReader(cmdOutput)
		line, err := bufOutput.ReadString('\n')

		for err == nil {
			fmt.Print(line)
			line, err = bufOutput.ReadString('\n')
		}
	}()

	time.Sleep(2 * time.Second) // wait for smee server startup

	fmt.Println("webhook server forwarded successfully from", *webhookUrl, "to port", *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {

		event := request.Header.Get("x-github-event")
		if event == "push" {
			fmt.Println("push event detected")

			err := os.Chdir(*dotFilePath)
			if err != nil {
				fmt.Println(err)
			}

			err = exec.Command("git", "pull", "origin", "main").Run()
			if err != nil {
				fmt.Printf("git repository failed to pull [%s]\n", err)
			} else {
				fmt.Println("git repository pulled successfully")
			}

			// run stow
			err = exec.Command("stow", ".").Run()
			if err != nil {
				fmt.Println("stow execution failed: ", err)
			} else {
				fmt.Println("stow execution succeeded")
			}
		}
	})

	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
