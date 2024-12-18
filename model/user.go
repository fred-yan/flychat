package model

import (
	"errors"
	"flychat/platform"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User 表示用户模型
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"type:varchar(255);not null;unique" json:"username"`
	Email     string    `gorm:"type:varchar(255);not null;unique" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string    `gorm:"type:varchar(255)" json:"nickname"`
	Phone     string    `gorm:"type:varchar(128)" json:"phone"`
	Avatar    string    `json:"avatar"`
	Role      Role      `gorm:"type:varchar(64)" json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate 在创建用户之前进行预处理
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// 例如：对密码进行哈希处理
	u.Role = RoleUser
	return nil
}

func CreateUser(user *User) error {
	db := platform.DB
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func UserExists(username, email string) bool {
	var count int64
	db := platform.DB
	if err := db.Model(&User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error; err != nil {
		log.Printf("Failed to check user existence: %v", err)
		return false
	}
	return count > 0
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	db := platform.DB
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &user, nil
}
