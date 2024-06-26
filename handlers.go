package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type SyncHandler struct {
	dotFilePath string
}

func (s SyncHandler) Sync(writer http.ResponseWriter, request *http.Request) {

	writer.Header().Set("Content-Type", "application/json")

	switch request.Method {
	case http.MethodPost:
		err := SyncExec(s.dotFilePath)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			writeResponse(writer, err.Error())
		} else {
			writeResponse(writer, "Sync completed...")
		}
	case http.MethodGet:

	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func writeResponse(writer io.Writer, msg string) {
	body := make(map[string]string, 1)
	body["msg"] = msg
	_ = json.NewEncoder(writer).Encode(body)

}
