package model

import (
	"gorm.io/gorm"
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
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate 在创建用户之前进行预处理
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// 例如：对密码进行哈希处理
	u.Role = RoleUser
	return nil
}
