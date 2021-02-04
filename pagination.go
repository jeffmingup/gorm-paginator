package pagination

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"strconv"
	"time"
)

// Param 分页参数
type Param struct {
	DB         *gorm.DB
	C          *gin.Context
	ExistModel bool
	OrderBy    []string
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
func Paging(p *Param, res interface{}) (*Paginator, error) {
	db := p.DB
	page, _ := strconv.Atoi(p.C.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(p.C.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if len(p.OrderBy) > 0 {
		for _, o := range p.OrderBy {
			db = db.Order(o)
		}
	}

	paginator := Paginator{
		PageSize: pageSize,
		Page:     1,
		PrevPage: 1,
		NextPage: 1,
		Records:  []interface{}{},
	}
	var count int64
	var offset int
	DBChannel := make(chan *gorm.DB, 1)
	ctx, cancel := context.WithTimeout(db.Statement.Context, time.Second*6)
	ctxDB := db.WithContext(ctx)
	go func() {
		var result *gorm.DB
		if p.ExistModel {
			result = ctxDB.Count(&count)
		} else {
			result = ctxDB.Model(res).Count(&count)
		}

		DBChannel <- result
		if count == 0 {
			cancel()
		}

	}()

	if page == 1 {
		offset = 0
	} else {
		offset = (page - 1) * pageSize
	}
	ctxDB2 := db.WithContext(ctx)
	go func() {
		result := ctxDB2.Limit(pageSize).Offset(offset).Find(res)
		DBChannel <- result
	}()

	for i := 0; i < 2; i++ {
		result := <-DBChannel
		if result.Error != nil {
			if result.Error == context.Canceled { //如果是count查询为零则手动关闭，直接返回空列表
				return &paginator, nil
			} else {
				cancel()
				return nil, result.Error
			}

		}
	}

	paginator.TotalRecord = count
	paginator.Records = res
	paginator.Page = page

	paginator.Offset = offset
	paginator.TotalPage = int(math.Ceil(float64(count) / float64(pageSize)))

	if page > 1 {
		paginator.PrevPage = page - 1
	}
	if page < paginator.TotalPage {
		paginator.NextPage = page + 1
	}
	return &paginator, nil
}
