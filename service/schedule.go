package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func SummaryHackerNewsTask(count int) (string, error) {
	logger.Infof("[%s] Start scheduled task SummaryHackerNewsTask", "scheduled task")
	startTime := time.Now()

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

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	logger.Infof("[%s] Finished scheduled task SummaryHackerNewsTask cost %v seconds", "scheduled task", duration)
	return "OK", nil
}

func contains(slice []int, id int) bool {
	for _, item := range slice {
		if item == id {
			return true
		}
	}
	return false
}

func SendMailTask() error {
	logger.Infof("[%s] Start scheduled task SendMailTask", "scheduled task")
	// 创建 POST 请求
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/v1/psummary", nil)
	if err != nil {
		logger.Warnf("[%s] create request error, %s", "scheduled task", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warnf("[%s] send request error, %s", "scheduled task", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("[%s] unexpected response, %s, %s", "scheduled task", resp.Body, err)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	logger.Infof("[%s] Finished scheduled task SendMailTask", "scheduled task")
	return nil
}
