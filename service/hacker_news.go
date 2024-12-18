package service

import (
	"encoding/json"
	"flychat/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type HackerNewsService struct {
}

var summaryService = &SummaryService{}

func fetchTopStories(hackerNewsUrl string) ([]int, error) {
	resp, err := http.Get(hackerNewsUrl + "v0/topstories.json?print=pretty")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top stories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var topStories []int
	if err := json.Unmarshal(body, &topStories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return topStories, nil
}

type Story struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	By          string `json:"by"`
	Url         string `json:"url"`
	Score       int    `json:"score"`
	Time        int    `json:"time"`
	Type        string `json:"type"`
	Kids        []int  `json:"kids"`
	Descendants int    `json:"descendants"`
}

func getStoryDetails(c *gin.Context, hackerNewsUrl string, storyId int) (*Story, error) {
	logger.Infof("[%s] Fetching story details for ID: %d", c.GetString("requestId"), storyId)
	storyUrl := fmt.Sprintf("%s/v0/item/%d.json", hackerNewsUrl, storyId)
	resp, err := http.Get(storyUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch story details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var story Story
	if err := json.Unmarshal(body, &story); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &story, nil
}

func (hn *HackerNewsService) HackerNewsSummary(c *gin.Context, count int) ([]SummaryResult, error) {
	hackerNewsUrl := os.Getenv("HACKER_NEWS_URL")
	topStories, err := fetchTopStories(hackerNewsUrl)
	if err != nil {
		logger.Warnf("[%s] fetchTopStories error, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "request error: " + err.Error()})
		return nil, err
	}

	var hSummaryResult []SummaryResult

	for _, storyId := range topStories[:count] {
		story, err := getStoryDetails(c, hackerNewsUrl, storyId)
		if err != nil {
			logger.Warnf("[%s] getStoryDetails error for storyId %d, %s", c.GetString("requestId"), storyId, err)
			continue
		}

		storyDB := &model.Story{
			StoryId:     story.Id,
			Title:       story.Title,
			By:          story.By,
			Url:         story.Url,
			Score:       story.Score,
			Time:        story.Time,
			Type:        story.Type,
			Descendants: story.Descendants,
		}
		storyExist, err := model.GetStory(story.Id)
		var storyExistSummaryResult SummaryResult
		if storyExist != nil {
			logger.Infof("[%s] Story %d already exists in DB", c.GetString("requestId"), story.Id)
			storyExistSummaryResult.Url = storyExist.Url
			storyExistSummaryResult.Summary = storyExist.Summary
			hSummaryResult = append(hSummaryResult, storyExistSummaryResult)
			continue
		}

		summaryResult, err := summaryService.GetSummary(c, story.Url)
		if err != nil {
			logger.Warnf("[%s] Failed to get summary: %s", c.GetString("requestId"), err)
		}
		hSummaryResult = append(hSummaryResult, *summaryResult)

		storyDB.Summary = summaryResult.Summary
		if err := model.CreateStory(storyDB); err != nil {
			logger.Warnf("[%s] CreateStory error for storyId %d, %s", c.GetString("requestId"), storyId, err)
			continue
		}
	}
	return hSummaryResult, nil
}
