package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	pagination "github.com/jeffmingup/gorm-paginator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strconv"
)

// User 用户
type User struct {
	ID       int
	UserName string `gorm:"not null;size:100;unique"`
}

func main() {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/blog_service?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("migrate fail:%v", err)
	}

	var count int64
	db.Model(User{}).Count(&count)
	if count == 0 {
		db.Create(User{ID: 1, UserName: "biezhi"})
		db.Create(User{ID: 2, UserName: "rose"})
		db.Create(User{ID: 3, UserName: "jack"})
		db.Create(User{ID: 4, UserName: "lili"})
		db.Create(User{ID: 5, UserName: "bob"})
		db.Create(User{ID: 6, UserName: "tom"})
		db.Create(User{ID: 7, UserName: "anny"})
		db.Create(User{ID: 8, UserName: "wat"})
		fmt.Println("Insert OK!")
	}

	var users []User

	pagination.Paging(&pagination.Param{
		DB:       db.Where("id > ?", 5),
		Page:     1,
		PageSize: 3,
		OrderBy:  []string{"id desc"},
		ShowSQL:  true,
	}, &users)

	fmt.Println("users:", users)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "3"))
		var users []User

		paginator, err := pagination.Paging(&pagination.Param{
			DB:       db,
			Page:     page,
			PageSize: limit,
			OrderBy:  []string{"id desc"},
			ShowSQL:  true,
		}, &users)
		if err != nil {
			c.JSON(500, err.Error())
		}
		c.JSON(200, paginator)
	})

	if err := r.Run(":8086"); err != nil {
		log.Printf("server start error:%v", err)
	}

}
