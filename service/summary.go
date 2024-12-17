package service

import (
	"context"
	"errors"
	"flychat/model"
	"flychat/platform"
	"fmt"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"io"
	"net/http"
)

var logger = platform.Logger

type summmaryService struct {
}

func getMKData(c *gin.Context, url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		logger.Warnf("[%s] request %s error, %s", url, c.GetString("requestId"), err)
		return "", err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Warnf("[%s] read body error, %s", c.GetString("requestId"), err)
		return "", err
	}

	content, err := htmltomarkdown.ConvertString(string(data))
	if err != nil {
		logger.Warnf("[%s] transfer body error, %s", c.GetString("requestId"), err)
		return "", err
	}
	return content, nil
}

func (s *summmaryService) GetSummary(c *gin.Context, url string) error {
	// 获取请求参数
	type Message struct {
		Role    openai.ChatCompletionMessageParamRole `json:"role"`
		Content string                                `json:"content"`
	}

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Warnf("[%s] get Writer flusher error", c.GetString("requestId")) //浏览器不兼容
		return errors.New("get Writer flusher error")
	}

	promptMessage := Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	}
	messages := []Message{promptMessage}

	content, err := getMKData(c, url)
	if err != nil {
		logger.Warnf("[%s] get url data error, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer body error: " + err.Error()})
		return err
	}

	promptContent := "请您反复阅读以下markdown语法的正文后，给出不超过100字的文章总结\n\n"
	userContent := promptContent + content
	userMessage := Message{
		Role:    "user",
		Content: userContent,
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messages = append(messages, userMessage)

	conversationId := c.GetString("requestId")
	params := openai.ChatCompletionNewParams{
		Messages:    openai.F([]openai.ChatCompletionMessageParamUnion{}),
		Model:       openai.F("qwen-turbo"),
		Temperature: openai.F(1.3),
		StreamOptions: openai.F(openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.F(true),
		}),
	}
	for _, message := range messages {
		var content any = message.Content
		params.Messages.Value = append(params.Messages.Value, openai.ChatCompletionMessageParam{
			Role:    openai.F(message.Role),
			Content: openai.F(content),
		})
	}

	go func() {
		userMessage := messages[len(messages)-1]
		if userMessage.Role != openai.ChatCompletionMessageParamRoleUser {
			return
		}
		if err := platform.DB.Create(&model.Message{
			ConversationId: conversationId,
			Role:           string(userMessage.Role),
			Content:        userMessage.Content,
		}); err != nil {
			logger.Warnf("[%s] create messge for db error, %s", c.GetString("requestId"), err)
		}
	}()

	stream := platform.LLMClient.Chat.Completions.NewStreaming(context.Background(), params)
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if _, err := fmt.Fprintf(w, content); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return err
			}
			flusher.Flush()
		}
		if content, ok := acc.JustFinishedContent(); ok {
			logger.Infof("[%s] finished content: %s", c.GetString("requestId"), content)
			break
		}
	}

	go func() {
		content := acc.Choices[0].Message.Content
		if err := platform.DB.Create(&model.Message{
			ConversationId: conversationId,
			Role:           string(openai.ChatCompletionAssistantMessageParamRoleAssistant),
			Content:        content,
		}).Error; err != nil {
			logger.Warnf("[%s] create messge content for db error, %s", c.GetString("requestId"), err)
		}
	}()
	if err := stream.Err(); err != nil {
		logger.Warnf("[%s] stream error, %s", c.GetString("requestId"), err)
		return err
	}
	return nil

}
