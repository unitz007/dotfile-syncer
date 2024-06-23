package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {

	var (
		webHookUrl  = useOrDefault()
		port        = "3000"
		dotFilePath = func() string {
			homeDir, _ := os.UserHomeDir()
			return path.Join(homeDir, "dotfiles")
		}()

		rootCmd = cobra.Command{
			Run: func(c *cobra.Command, args []string) {
				webHookUrl = useOrDefault(c.Flag("webhook"), "http://48.217.222.93:3000/ZLe6soOKhrpzUEA")
				port = useOrDefault(c.Flag("port"), "3000")
				dotFilePath = useOrDefault(c.Flag("dotfile-path"), 
			},
		}
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(webHookUrl)
	fmt.Println(port)
	fmt.Println(dotFilePath)

	go func() {
		cmd := exec.Command("smee", "--url", webHookUrl, "--path", "/webhook", "--port", port)
		stdOutput, err := cmd.StdoutPipe()
		stdErr, err := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if stdErr != nil {
			buf := bufio.NewReader(stdErr)
			line, _ := buf.ReadString('\n')
			fmt.Println("Failed to start smee.io server:", line)
			os.Exit(1)
		}

		cmdOutput := io.MultiReader(stdOutput, stdErr)
		bufOutput := bufio.NewReader(cmdOutput)
		line, err := bufOutput.ReadString('\n')

		for err == nil {
			fmt.Print(line)
			line, err = bufOutput.ReadString('\n')
		}
	}()

	time.Sleep(3 * time.Second) // wait for smee server startup

	fmt.Println("smee.io Server started successfully")

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {

		event := request.Header.Get("x-github-event")
		if event == "push" {
			fmt.Println("push event detected")

			err := os.Chdir(dotFilePath)
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
}

func useOrDefault(value *pflag.Flag, defaultValue func() string) string {
	if value == nil {
		_ = value.Value.Set(defaultValue())
	}

	return value.Value.String()
}
