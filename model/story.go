package model

import (
	"errors"
	"flychat/platform"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

type Story struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StoryId     int       `gorm:"index" db:"storyId"`
	Title       string    `gorm:"type:text" db:"title"`
	Summary     string    `gorm:"type:text" db:"summary"`
	By          string    `gorm:"type:varchar(255)" db:"by"`
	Url         string    `gorm:"type:text" db:"url"`
	Score       int       `db:"score"`
	Time        time.Time `db:"time"`
	Type        string    `gorm:"type:varchar(64)" db:"type"`
	Descendants int       `db:"descendants"`
	IsPublished bool      `db:"is_published" gorm:"default:false"`
}

func CreateStory(story *Story) error {
	db := platform.DB
	return db.Create(story).Error
}

func IsStoryExist(storyId int) bool {
	db := platform.DB
	var count int64
	db.Model(&Story{}).Where("story_id = ?", storyId).Count(&count)
	return count > 0
}

func GetStory(storyId int) (s *Story, err error) {
	var story Story
	db := platform.DB
	if err := db.Where("story_id = ?", storyId).First(&story).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("story not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &story, nil
}

func UpdateStorySummary(storyId int, summary string) error {
	db := platform.DB

	if err := db.Model(&Story{}).Where("story_id = ?", storyId).Update("summary", summary).Error; err != nil {
		return fmt.Errorf("failed to update story Summary: %w", err)
	}

	return nil
}

func UpdateStoryScore(storyId int, score int) error {
	db := platform.DB

	if err := db.Model(&Story{}).Where("story_id = ?", storyId).Update("score", score).Error; err != nil {
		return fmt.Errorf("failed to update story Score: %w", err)
	}

	return nil
}

func GetStoryList(isPublished bool) ([]Story, error) {
	db := platform.DB
	var stories []Story
	err := db.Where("is_published = ?", isPublished).Find(&stories).Error
	if err != nil {
		log.Printf("Failed to fetch stories: %v", err)
		return nil, err
	}

	return stories, nil
}

func UpdateStoriesPublishedStatus(storyIds []int, isPublished bool) error {
	db := platform.DB

	if len(storyIds) == 0 {
		return nil // 如果传入的切片为空，则无需执行任何操作
	}

	// 使用 IN 子句批量更新 is_published 字段为 true
	result := db.Model(&Story{}).
		Where("story_id IN ?", storyIds).
		Update("is_published", isPublished)

	if result.Error != nil {
		return fmt.Errorf("failed to publish stories: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no stories were updated")
	}

	return nil
}

func GetStoryListByScore(storyNum int, days int) ([]Story, error) {
	db := platform.DB
	var stories []Story

	now := time.Now()
	threeDaysAgo := now.Add(time.Duration(-days) * 24 * time.Hour)

	// 构建查询条件
	err := db.Where("time >= ? AND time <= ?", threeDaysAgo, now).
		Order("score DESC").
		Limit(storyNum).
		Find(&stories).Error

	if err != nil {
		log.Printf("Failed to fetch stories by score: %v", err)
		return nil, err
	}

	return stories, nil
}
