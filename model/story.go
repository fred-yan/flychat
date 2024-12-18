package model

import (
	"errors"
	"flychat/platform"
	"fmt"
	"gorm.io/gorm"
)

type Story struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	StoryId     int    `gorm:"index" db:"storyId"`
	Title       string `gorm:"type:text" db:"title"`
	Summary     string `gorm:"type:text" db:"summary"`
	By          string `gorm:"type:varchar(255)" db:"by"`
	Url         string `gorm:"type:text" db:"url"`
	Score       int    `db:"score"`
	Time        int    `db:"time"`
	Type        string `gorm:"type:varchar(64)" db:"type"`
	Descendants int    `db:"descendants"`
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

func GetStoryList() ([]Story, error) {
	db := platform.DB
	var stories []Story
	return stories, db.Find(&stories).Error
}
