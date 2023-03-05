package model

import "time"

type ApiKey struct {
	Id        uint64    `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key"`
	UserId    uint64    `json:"user_id" `
	Prefix    string    `json:"prefix"`
	Name      string    `json:"name"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
	User      User      `json:"user" gorm:"foreignKey:UserId;references:user_id"`
}
