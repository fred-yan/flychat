package platform

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"os"
)

var (
	LLMClient *openai.Client
)

func InitLLMClient() {
	LLMClient = openai.NewClient(
		option.WithBaseURL(os.Getenv("LLM_BASE_URL")),
		option.WithAPIKey(os.Getenv("LLM_API_KEY")),
	)
}
