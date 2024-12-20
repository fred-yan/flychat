package service

import (
	"encoding/json"
	"flychat/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jordan-wright/email"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"
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
			Time:        time.Unix(int64(story.Time), 0),
			Type:        story.Type,
			Descendants: story.Descendants,
		}
		storyExist, err := model.GetStory(story.Id)
		var storyExistSummaryResult SummaryResult
		if storyExist != nil {
			err = model.UpdateStoryScore(story.Id, story.Score)
			if err != nil {
				logger.Warnf("[%s] UpdateStoryScore error for storyId %d, %s", c.GetString("requestId"), story.Id, err)
			}

			logger.Infof("[%s] Story %d already exists in DB", c.GetString("requestId"), story.Id)
			storyExistSummaryResult.Url = storyExist.Url
			storyExistSummaryResult.Summary = storyExist.Summary
			hSummaryResult = append(hSummaryResult, storyExistSummaryResult)
			continue
		}

		summaryResult, err := summaryService.GetSummary(c, story.Url)
		if err != nil {
			logger.Warnf("[%s] Failed to get summary: %s", c.GetString("requestId"), err)
			errInfo := "\n文章解析出错，抱歉无法给出文章总结！"
			summaryResult = &SummaryResult{Summary: errInfo, Url: story.Url}
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

func produceMailText(c *gin.Context) (string, error) {
	mailText := ""
	var notPublishedStoryIds []int
	var hotStoryIds []int
	var newStoryIds []int

	storyNum, _ := strconv.Atoi(os.Getenv("HACKER_NEWS_HOT_STORIES_NUMBER"))
	days, _ := strconv.Atoi(os.Getenv("HACKER_NEWS_HOT_STORIES_DAYS"))
	hotStoryList, err := model.GetStoryListByScore(storyNum, days)
	if err != nil {
		logger.Warnf("[%s] get hot story list error, %s", c.GetString("requestId"), err)
		return "", err
	}

	mailText = mailText + "## Hacker News 热点\n\n"
	for _, hStory := range hotStoryList {
		hotStoryIds = append(hotStoryIds, hStory.StoryId)
		mailText = mailText + "[**" + hStory.Title + "**](" + hStory.Url + ") [**Hot:" +
			strconv.Itoa(hStory.Score) + "**]\n" +
			hStory.Summary + "\n\n---\n\n"

		if !hStory.IsPublished {
			notPublishedStoryIds = append(notPublishedStoryIds, hStory.StoryId)
		}
	}

	isPublished := false
	storyList, err := model.GetStoryList(isPublished)
	if err != nil {
		logger.Warnf("[%s] get hot story list error, %s", c.GetString("requestId"), err)
		return "", err
	}

	mailText = mailText + "## Hacker News 一天概览\n\n"
	for _, story := range storyList {
		if !contains(hotStoryIds, story.StoryId) {
			mailText = mailText + "[**" + story.Title + "**](" + story.Url + ") [**Hot:" +
				strconv.Itoa(story.Score) + "**]\n" +
				story.Summary + "\n\n---\n\n"
			notPublishedStoryIds = append(notPublishedStoryIds, story.StoryId)
			newStoryIds = append(newStoryIds, story.StoryId)
		}
	}

	logger.Infof("[%s] notPublishedStoryIds: %v, newStoryIds: %v", c.GetString("requestId"),
		notPublishedStoryIds, newStoryIds)
	err = model.UpdateStoriesPublishedStatus(notPublishedStoryIds, true)
	if err != nil {
		logger.Warnf("[%s] update story published status error, %s", c.GetString("requestId"), err)
		return "", err
	}

	logger.Infof("[%s] get hotStoryIds: %v , newStoryIds: %v", c.GetString("requestId"), hotStoryIds, newStoryIds)
	return mailText, nil
}

func (hn *HackerNewsService) MailSummary(c *gin.Context) error {
	logger.Infof("[%s] Start mail summary", c.GetString("requestId"))
	mailText, err := produceMailText(c)
	if err != nil {
		logger.Warnf("[%s] produceMailText error, %s", c.GetString("requestId"), err)
		return err
	}
	mailTextHtml := mdToHtml([]byte(mailText))

	var sendUsers []string
	userList, err := model.GetUserList()
	if err != nil {
		logger.Warnf("[%s] get user list error, %s", c.GetString("requestId"), err)
		return err
	}

	e := email.NewEmail()
	mailFrom := os.Getenv("MAIL_SEND_USER")
	mailSmtpAdd := os.Getenv("MAIL_SMTP_ADD")
	mailUsername := os.Getenv("MAIL_USERNAME")
	mailHost := os.Getenv("MAIL_HOST")
	mailSk := os.Getenv("MAIL_SK")

	for _, user := range userList {
		sendUsers = append(sendUsers, user.Username)
		e.From = mailFrom
		e.To = []string{user.Email}
		formattedDate := time.Now().Format("2006-01-02")
		e.Subject = formattedDate + " Hacker News 信息概览"
		e.HTML = mailTextHtml
		err := e.Send(mailSmtpAdd, smtp.PlainAuth("", mailUsername, mailSk, mailHost))
		if err != nil {
			logger.Warnf("[%s] send email error, %s", c.GetString("requestId"), err.Error())
		}
	}
	logger.Infof("[%s] mail have sent to userList: %v", c.GetString("requestId"), sendUsers)
	logger.Infof("[%s] Finished mail summary", c.GetString("requestId"))
	return nil
}
