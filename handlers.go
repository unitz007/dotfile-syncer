package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type SyncHandler struct {
	syncer     *Syncer
	db         Database
	httpClient HttpClient
}

func NewSyncHandler(syncer *Syncer, db Database, httpClient HttpClient) *SyncHandler {
	return &SyncHandler{
		syncer,
		db,
		httpClient,
	}
}

func (s SyncHandler) Sync(writer http.ResponseWriter, request *http.Request) {

	config := s.syncer.config

	writer.Header().Set("Content-Type", "application/json")
	remoteCommit := func(c HttpClient) *Commit {
		remoteCommitResponse, err := s.httpClient.GetCommits()
		if err != nil {
			Error(err.Error())
			return nil
		}

		commit := remoteCommitResponse[0]

		return &Commit{
			Id: commit.Sha,
		}
	}

	switch request.Method {
	case http.MethodPost: // POST
		err := s.syncer.Sync(s.syncer.config.DotfilePath)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			Error(err.Error())
			writeResponse(writer, err.Error(), nil)
		} else {
			//remoteCommit := remoteCommit(s.httpClient)
			//err = s.db.Create(remoteCommit)
			//if err != nil {
			//	Error("Error: Could not persist remote commit to database:", err.Error())
			//}
			writeResponse(writer, "Sync completed...", nil)
		}
	case http.MethodGet: // GET

		defer func() {
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
					Error("Error:", line)
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

		}()

		localCommit, _ := s.db.Get(1)
		remoteCommit := remoteCommit(s.httpClient)

		response := InitGitTransform(localCommit, remoteCommit)

		writeResponse(writer, "sync details fetched successfully", response)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func writeResponse(writer io.Writer, msg string, payload any) {
	body := make(map[string]any, 2)
	body["msg"] = msg
	body["payload"] = payload
	_ = json.NewEncoder(writer).Encode(body)

}
