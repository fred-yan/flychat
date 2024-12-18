package model

import "time"

type Message struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId         string    `json:"user_id" gorm:"index:idx_user_id_conversation_id_created_at"`
	ConversationId string    `json:"conversation_id" gorm:"index:idx_user_id_conversation_id_created_at"`
	Role           string    `gorm:"type:varchar(64)" json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at" gorm:"index:idx_user_id_conversation_id_created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
