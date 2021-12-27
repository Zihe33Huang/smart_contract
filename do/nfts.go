package db

import "time"

type NFTS struct {
	ID              int       `gorm:"primary_key;AUTO_INCREMENT;not null;column:id" json:"id"`
	TokenId         string    `gorm:"not null;column:tokenid" json:"tokenid"`
	Variety         string    `gorm:"not null;column:variety" json:"variety"`
	Order           string    `gorm:"not null;column:order" json:"order"`
	Code            string    `gorm:"unique;not null;column:code" json:"code"`
	CID             string    `gorm:"type:varchar(255);not null;column:cid" json:"custodian"`
	Address         string    `gorm:"type:varchar(255);column:address" json:"address"`
	Status          int       `gorm:"not null;column:status;" json:"status"`
	Description     string    `gorm:"type:varchar(255);not null;column:description;" json:"decription"`
	TransactionHash string    `gorm:"type:varchar(255);column:transaction_hash;" json:"transactionHash"`
	UpdateTime      time.Time `gorm:"column:update_time"`
}

type User struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;not null;column:id" json:"id"`
	UserId     string    `gorm:"type:varchar(100);unique;not null;column:userid" json:"userid"`
	UserName   string    `gorm:"type:varchar(255);column:username;" json:"username"`
	Phone      string    `gorm:"type:varchar(100);unique;column:phone;index:user"  json:"phone"`
	Path       string    `gorm:"type:varchar(255);unique;column:path;" json:"path"`
	Password   string    `gorm:"type:varchar(255);column:password;" json:"password"`
	Role       int       `gorm:"column:role;default" json:"role"`
	Address    string    `gorm:"type:varchar(255);unique;column:address" json:"address"`
	UpdateTime time.Time `gorm:"column:update_time" `
}

type UserInfo struct {
	UserId  string `gorm:"type:varchar(100);unique;not null;column:userid" json:"userid"`
	Path    string `gorm:"type:varchar(255);unique;column:path;" json:"path"`
	Address string `gorm:"type:varchar(255);unique;column:address" json:"address"`
}
