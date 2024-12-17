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

func (ch ChatController) Chat(c *gin.Context) {
	var reqData struct {
		Url string `json:"url"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warnf("[%s] Invalid input, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := summaryService.GetSummary(c, reqData.Url); err != nil {
		logger.Warnf("[%s] Failed to get summary: %s", c.GetString("requestId"), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get summary"})
		return
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
