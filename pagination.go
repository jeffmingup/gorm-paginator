package gorm_paginator

import (
	"context"
	"gorm.io/gorm"
	"math"
	"sync"
	"time"
)

// Param 分页参数
type Param struct {
	DB       *gorm.DB
	Page     int
	PageSize int
	OrderBy  []string
	ShowSQL  bool
}

// Paginator 分页返回
type Paginator struct {
	TotalRecord int64       `json:"total_record"`
	TotalPage   int         `json:"total_page"`
	Records     interface{} `json:"records"`
	Offset      int         `json:"offset"`
	PageSize    int         `json:"page_size"`
	Page        int         `json:"page"`
	PrevPage    int         `json:"prev_page"`
	NextPage    int         `json:"next_page"`
}

// Paging 分页
func Paging(p *Param, result interface{}) *Paginator {
	db := p.DB

	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
	if len(p.OrderBy) > 0 {
		for _, o := range p.OrderBy {
			db = db.Order(o)
		}
	}

	var paginator Paginator
	var count int64
	var offset int
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(db.Statement.Context, time.Second*6)
	ctxDB := db.WithContext(ctx)
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctxDB.Model(result).Count(&count)
		if count == 0 {
			cancel()
		}
	}()

	if p.Page == 1 {
		offset = 0
	} else {
		offset = (p.Page - 1) * p.PageSize
	}
	ctxDB2 := db.WithContext(ctx)
	go func() {
		defer wg.Done()
		ctxDB2.Limit(p.PageSize).Offset(offset).Find(result)
	}()

	wg.Wait()

	paginator.TotalRecord = count
	paginator.Records = result
	paginator.Page = p.Page

	paginator.Offset = offset
	paginator.PageSize = p.PageSize
	paginator.TotalPage = int(math.Ceil(float64(count) / float64(p.PageSize)))

	if p.Page > 1 {
		paginator.PrevPage = p.Page - 1
	} else {
		paginator.PrevPage = p.Page
	}

	if p.Page == paginator.TotalPage {
		paginator.NextPage = p.Page
	} else {
		paginator.NextPage = p.Page + 1
	}
	return &paginator
}
