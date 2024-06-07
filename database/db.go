/*
 * @Author: flwfdd
 * @Date: 2024-06-06 16:42:56
 * @LastEditTime: 2024-06-06 21:59:43
 * @Description:
 * _(:з」∠)_
 */
package database

import (
	"BIT-Helper/util/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 枚举用户类型(需要与数据库中定义一致)
const (
	Identity_Normal = iota
	Identity_Admin
)

// 基本模型
type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"create_time"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"delete_time"`
}

// 用户
type User struct {
	Base
	Sid      string `gorm:"not null;uniqueIndex;size:42" json:"sid"`
	Nickname string `gorm:"not null;unique" json:"nickname"`
	Avatar   string `json:"avatar"`
	Intro    string `json:"intro"`
	Score    int    `json:"score"`
	Identity int    `json:"identity"`
}

// 图片
type Image struct {
	Base
	Mid  string `gorm:"not null;uniqueIndex;size:233" json:"mid"`
	Size uint   `gorm:"not null" json:"size"`
	Uid  uint   `gorm:"not null" json:"uid"`
}

// 商品
type Goods struct {
	Base
	Type  int    `gorm:"not null" json:"type"`
	Uid   uint   `gorm:"not null" json:"uid"`
	Title string `gorm:"not null" json:"title"`
	Intro string `gorm:"not null" json:"intro"`
	Num   int    `gorm:"not null" json:"num"`
	Price int    `gorm:"not null" json:"price"`
}

func Init() {
	dsn := config.Config.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db

	err = db.AutoMigrate(
		&User{}, &Image{}, &Goods{},
	)
	if err != nil {
		panic(err)
	}
}
