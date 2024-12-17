package controller

import (
	"context"
	"flychat/model"
	"flychat/platform"
	"fmt"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"io"
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
		Url string `json:"url"`
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

	promptMessage := Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	}
	messages := []Message{promptMessage}

	res, err := http.Get(reqData.Url)
	if err != nil {
		logger.Warnf("[%s] request %s error, %s", reqData.Url, c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "request " + reqData.Url + " error: " + err.Error()})
		return
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Warnf("[%s] read body error, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "read request data error: " + err.Error()})
		return
	}

	content, err := htmltomarkdown.ConvertString(string(data))
	if err != nil {
		logger.Warnf("[%s] transfer body error, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer body error: " + err.Error()})
		return
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
	//messages := reqData.Messages
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
		Url      string `json:"url" binding:"url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warnf("[%s] Invalid input, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	res, err := http.Get(input.Url)
	if err != nil {
		logger.Warnf("[%s] request %s error, %s", input.Url, c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "request " + input.Url + " error: " + err.Error()})
		return
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Warnf("[%s] read body error, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "read request data error: " + err.Error()})
		return
	}

	content, err := htmltomarkdown.ConvertString(string(data))
	if err != nil {
		logger.Warnf("[%s] transfer body error, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer body error: " + err.Error()})
		return
	}

	logger.Infof("[%s] Test success for username: %s, email: %s", c.GetString("requestId"), input.Username, input.Email)
	c.JSON(http.StatusOK, gin.H{"content": content})
}
