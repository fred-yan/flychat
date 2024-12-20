package service

import (
	"context"
	"errors"
	"flychat/model"
	"flychat/platform"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/openai/openai-go"
	"io"
	"net/http"
	"strings"
)

var logger = platform.Logger

type SummaryService struct {
}

type SummaryResult struct {
	Summary string `json:"summary"`
	Url     string `json:"url"`
}

func htmlToMd(data []byte) (string, error) {
	content, err := htmltomarkdown.ConvertString(string(data))
	if err != nil {
		logger.Warnf("transfer body error, %s", err)
		return "", err
	}
	return content, nil
}

func mdToHtml(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func (s *SummaryService) GetSummary(c *gin.Context, url string) (*SummaryResult, error) {
	type Message struct {
		Role    openai.ChatCompletionMessageParamRole `json:"role"`
		Content string                                `json:"content"`
	}

	promptMessage := Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	}
	messages := []Message{promptMessage}

	res, err := http.Get(url)
	if err != nil {
		logger.Warnf("[%s] request %s error, %s", url, c.GetString("requestId"), err)
		return nil, err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Warnf("[%s] read body error, %s", c.GetString("requestId"), err)
		return nil, err
	}

	contentType := http.DetectContentType(data)
	if !strings.Contains(contentType, "html") {
		logger.Warnf("[%s] request %s content type is %s not support", c.GetString("requestId"), url, contentType)
		summary := "\n文章链接格式暂不支持，抱歉无法给出文章总结！"
		return &SummaryResult{Summary: summary, Url: url}, nil
	}

	logger.Infof("[%s] request %s success, data lengeth is %d", c.GetString("requestId"), url, len(data))
	content, err := htmlToMd(data)
	if err != nil {
		logger.Warnf("[%s] get url data error, %s", c.GetString("requestId"), err)
		return nil, err
	}

	if content == "" {
		logger.Warnf("[%s] get url markdwon data is null", c.GetString("requestId"))
		summary := "\n无法获取文章链接内容，抱歉无法给出文章总结！"
		return &SummaryResult{Summary: summary, Url: url}, nil
	}

	promptContent := "请您反复阅读以下markdown语法的正文后，先给出不超过200字的英文总结。\n" +
		"然后请在英文总结后面再补充一段不超过200字的中文总结\n" +
		"输出格式如下：\n" +
		"- **English Summary**\n\n" +
		"{英文总结}\n\n" +
		"- **中文总结**\n\n" +
		"{中文总结}"
	userContent := promptContent + content
	userMessage := Message{
		Role:    "user",
		Content: userContent,
	}

	messages = append(messages, userMessage)

	conversationId := c.GetString("requestId")
	params := openai.ChatCompletionNewParams{
		Messages:    openai.F([]openai.ChatCompletionMessageParamUnion{}),
		Model:       openai.F("qwen-turbo"),
		Temperature: openai.F(1.3),
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
		}).Error; err != nil {
			logger.Warnf("[%s] create messge for db error, %s", c.GetString("requestId"), err)
		}
	}()

	// 发送非流式请求
	response, err := platform.LLMClient.Chat.Completions.New(context.Background(), params)
	if err != nil {
		logger.Warnf("[%s] chat completion error, %s", c.GetString("requestId"), err)
		return nil, errors.New("chat completion error")
	}

	// 处理响应
	if len(response.Choices) == 0 {
		logger.Warnf("[%s] no choices in response", c.GetString("requestId"))
		return nil, errors.New("no choices in response")
	}

	summary := response.Choices[0].Message.Content

	// 保存消息到数据库
	if err := platform.DB.Create(&model.Message{
		ConversationId: conversationId,
		Role:           string(openai.ChatCompletionAssistantMessageParamRoleAssistant),
		Content:        summary,
	}).Error; err != nil {
		logger.Warnf("[%s] create messge for db error, %s", c.GetString("requestId"), err)
	}

	return &SummaryResult{Summary: summary, Url: url}, nil
}
