package service

import (
	"bytes"
	"encoding/json"
	"flychat/model"
	"fmt"
	"github.com/jordan-wright/email"
	"net/http"
	"net/smtp"
)

func StartHSummary(count int) (string, error) {
	logger.Infof("[%s] Start scheduled task StartHSummary", "scheduled task")
	requestBody := struct {
		Count int
	}{
		Count: count,
	}
	// 将请求体序列化为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建 POST 请求
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/v1/hsummary", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Warnf("[%s] create request error, %s", "scheduled task", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warnf("[%s] send request error, %s", "scheduled task", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("[%s] unexpected response, %s, %s", "scheduled task", resp.Body, err)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	logger.Infof("[%s] Finished scheduled task StartHSummary", "scheduled task")
	return "OK", nil
}

func SendEMail() {
	logger.Infof("[%s] Start scheduled task SendEMail", "scheduled task")
	storyList, err := model.GetStoryList()
	if err != nil {
		logger.Warnf("[%s] get story list error, %s", "scheduled task", err)
	}
	mailText := ""
	for _, story := range storyList {
		mailText = mailText + "###" + story.Title + "\n" + story.Url + "\n" + story.Summary + "\n\n"
	}

	userList, err := model.GetUserList()
	if err != nil {
		logger.Warnf("[%s] get user list error, %s", "scheduled task", err)
	}
	e := email.NewEmail()
	for _, user := range userList {
		e.From = "fredyan <262740590@qq.com>"
		e.To = []string{user.Email}
		e.Subject = "Hacker News Summary"
		e.Text = []byte(mailText)
		err := e.Send("smtp.qq.com:587", smtp.PlainAuth("", "262740590@qq.com", "", "smtp.qq.com"))
		if err != nil {
			logger.Warnf("[%s] send email error, %s", "scheduled task", err.Error())
		}
	}
	logger.Infof("[%s] Finished scheduled task SendEMail", "scheduled task")
}
