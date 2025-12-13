package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/manisprasad/llm-serve/internal/types"
	"github.com/manisprasad/llm-serve/internal/utils/response"
)

func New(baseUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var chat types.Chat

		fmt.Println("Base URL is:", baseUrl)

		if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
			response.WriteJson(w, http.StatusBadRequest, nil, "invalid request body: "+err.Error())
			return
		}

		slog.Info("Incoming query", "message", chat.Message)

		// Validate required fields
		if err := validator.New().Struct(chat); err != nil {
			response.WriteJson(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// Build request for Ollama
		body := map[string]any{
			"model": chat.ModelName,
			"messages": []map[string]string{
				{"role": "user", "content": chat.Message},
			},
			"stream":      chat.Stream,
			"temperature": 0.7,            // optional: controls randomness
			"max_tokens":  1024,           // optional: max output length
			"top_p":       0.9,            // optional: nucleus sampling
			"stop":        []string{"\n"}, // optional: stop sequences
		}

		bts, err := json.Marshal(body)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, nil, "failed to marshal request body: "+err.Error())
			return
		}

		// Ensure the base URL ends with /api/chat properly
		endpoint := baseUrl
		if endpoint[len(endpoint)-1] != '/' {
			endpoint += "/"
		}
		endpoint += "chat"

		req, err := http.NewRequest("POST", endpoint, bytes.NewReader(bts))
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, nil, "failed to create request: "+err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, nil, "failed to call LLM: "+err.Error())
			return
		}
		defer resp.Body.Close()

		// Non-streaming: decode once and return
		if !chat.Stream {
			var out any
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				response.WriteJson(w, http.StatusInternalServerError, nil, "invalid LLM response: "+err.Error())
				return
			}
			response.WriteJson(w, resp.StatusCode, out, "")
			return
		}

		// Streaming: setup SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			response.WriteJson(w, http.StatusInternalServerError, nil, "streaming not supported")
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				trimmed := bytes.TrimSpace(line)
				if len(trimmed) > 0 {
					// Send each chunk as SSE
					fmt.Fprintf(w, "data: %s\n\n", trimmed)
					flusher.Flush()
				}
			}
			if err != nil {
				break
			}
		}
	}
}
