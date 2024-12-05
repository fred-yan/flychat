package controller

import (
	"context"
	"flychat/model"
	"flychat/platform"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"log"
	"net/http"
)

type ChatController struct{}

func (ch ChatController) Chat(c *gin.Context) {
	// 获取请求参数
	type Message struct {
		Role    openai.ChatCompletionMessageParamRole `json:"role"`
		Content string                                `json:"content"`
	}

	var reqData struct {
		ConversationId string    `json:"conversation_id"`
		Messages       []Message `json:"messages"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Panic("server not support") //浏览器不兼容
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	conversationId := reqData.ConversationId
	messages := reqData.Messages
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
			log.Println(err)
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
				return
			}
			flusher.Flush()
		}
		if content, ok := acc.JustFinishedContent(); ok {
			log.Println("finished content:", content)
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
			log.Println(err)
		}
	}()
	if err := stream.Err(); err != nil {
		log.Println(err)
	}

}

func (ch ChatController) Test(c *gin.Context) {
	logger.Infof("[%s] Handling test request", c.GetString("requestId"))
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warnf("[%s] Invalid input, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	logger.Infof("[%s] Test success for username: %s, email: %s", c.GetString("requestId"), input.Username, input.Email)
	c.JSON(http.StatusOK, gin.H{input.Username: input.Email})
}
