package llm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

func TestLMStudioProvider_Ping_OK(t *testing.T) {
	srv := newTestLMStudioServer([]string{"meta-llama-3.2-3b"})
	defer srv.Close()

	p := llm.NewLMStudioProvider(srv.URL, "meta-llama-3.2-3b", time.Second)
	if !p.Ping(context.Background()) {
		t.Error("Ping should return true for a running LM Studio server")
	}
}

func TestLMStudioProvider_Ping_Unreachable(t *testing.T) {
	p := llm.NewLMStudioProvider("http://localhost:1", "model", 100*time.Millisecond)
	if p.Ping(context.Background()) {
		t.Error("Ping should return false for an unreachable server")
	}
}

func TestLMStudioProvider_ListModels(t *testing.T) {
	srv := newTestLMStudioServer([]string{"model-a", "model-b", "model-c"})
	defer srv.Close()

	p := llm.NewLMStudioProvider(srv.URL, "model-a", time.Second)
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 3 {
		t.Errorf("ListModels: got %d, want 3", len(models))
	}
}

func TestLMStudioProvider_Chat_Streaming(t *testing.T) {
	// Serveur simulant le SSE OpenAI de LM Studio
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]string{{"id": "test-model"}},
			})
			return
		}
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "text/event-stream")
			// Format SSE OpenAI
			type delta struct{ Content string }
			type choice struct {
				Delta        delta   `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			}
			type chunk struct {
				Choices []choice `json:"choices"`
			}

			sendChunk := func(content string, done bool) {
				var fr *string
				if done {
					s := "stop"
					fr = &s
				}
				c := chunk{Choices: []choice{{
					Delta:        delta{Content: content},
					FinishReason: fr,
				}}}
				b, _ := json.Marshal(c)
				fmt.Fprintf(w, "data: %s\n\n", b)
				w.(http.Flusher).Flush()
			}

			sendChunk("Salut", false)
			sendChunk(" !", true)
			fmt.Fprint(w, "data: [DONE]\n\n")
			w.(http.Flusher).Flush()
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := llm.NewLMStudioProvider(srv.URL, "test-model", time.Second)
	ch, err := p.Chat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Dis salut"},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	var result string
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatalf("Chunk error: %v", chunk.Err)
		}
		result += chunk.Content
	}

	if result != "Salut !" {
		t.Errorf("Chat result: got %q, want %q", result, "Salut !")
	}
}

func TestLMStudioProvider_Metadata(t *testing.T) {
	p := llm.NewLMStudioProvider("http://localhost:1234", "my-model", time.Second)
	if p.Name() != "LM Studio" {
		t.Errorf("Name: got %q, want %q", p.Name(), "LM Studio")
	}
	if p.ModelName() != "my-model" {
		t.Errorf("ModelName: got %q, want %q", p.ModelName(), "my-model")
	}
	if p.BaseURL() != "http://localhost:1234" {
		t.Errorf("BaseURL: got %q, want %q", p.BaseURL(), "http://localhost:1234")
	}
}
