package model

type User struct {
	Uname         string `json:"uname"`
	DocumentBlock string `json:"document_block" `
	Parent        uint64 `json:"parent"`
	UserId        uint64 `json:"user_id" gorm:"primaryKey"`
	AclRoleId     uint64 `json:"acl_role_id"`
	SendBlock     string `json:"send_block"`
}

func (u User) TableName() string  {
	return "user"
}
