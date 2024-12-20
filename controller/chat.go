package controller

import (
	"flychat/service"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type ChatController struct{}

var summaryService = service.SummaryService{}
var hackerNewsService = service.HackerNewsService{}

func (ch ChatController) Summary(c *gin.Context) {
	var reqData struct {
		Url string `json:"url"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warnf("[%s] Invalid input, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	summaryResult, err := summaryService.GetSummary(c, reqData.Url)
	if err != nil {
		logger.Warnf("[%s] Failed to get summary: %s", c.GetString("requestId"), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get summary"})
		return
	}
	c.JSON(http.StatusOK, summaryResult)
}

func (ch ChatController) SummaryPages(c *gin.Context) {
	logger.Infof("[%s] Handling SummaryPages request", c.GetString("requestId"))
	var input struct {
		Count int `json:"count"`
	}

	err := c.ShouldBindJSON(&input)
	if err != nil {
		logger.Warnf("[%s] Invalid input, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var hSummaryResult []service.SummaryResult
	hSummaryResult, err = hackerNewsService.HackerNewsSummary(c, input.Count)
	if err != nil {
		logger.Warnf("[%s] Failed to get HackerNews summary: %s", c.GetString("requestId"), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Result": hSummaryResult})
}

func (ch ChatController) PublishSummary(c *gin.Context) {
	err := hackerNewsService.MailSummary(c)
	if err != nil {
		logger.Warnf("[%s] Failed to publish summary: %s", c.GetString("requestId"), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish summary" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Status": "Success"})
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
