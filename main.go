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
)

func main() {

	go func() {

		cmd := exec.Command("smee", "--url", "https://smee.io/awFay3gs7LCGYe2", "--path", "/webhook", "--port", "3000")
		stdOutput, err := cmd.StdoutPipe()
		stdErr, err := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			panic(err)
		}

		if stdErr != nil {
			buf := bufio.NewReader(stdErr)
			line, _ := buf.ReadString('\n')
			log.Println("Failed to start smee.io server:", line)
			os.Exit(1)
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

	log.Println("smee.io Server started successfully")

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {

		event := request.Header.Get("x-github-event")
		if event == "push" {
			log.Println("push event detected")

			homeDir, err := os.UserHomeDir()
			dotFileDir := path.Join(homeDir, "dotfiles")

			err = os.Chdir(dotFileDir)
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
				fmt.Println("stow execution failed: ", err)
			} else {
				log.Println("stow execution succeeded")
			}
		}
	})

	log.Fatal(http.ListenAndServe(":3000", mux))

}
