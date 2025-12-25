package paginate

import (
	"context"
	"edgelink-api/internal/api/dto"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultPageNum  = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

func PaginateList[T any](ctx context.Context, db *gorm.DB, page dto.Page, fns ...func(*gorm.DB) *gorm.DB) (dto.Page, error) {
	if !page.UnlimitedPageSize {
		if page.PageNum < 1 {
			page.PageNum = defaultPageNum
		}
		if page.PageSize < 1 {
			page.PageSize = defaultPageSize
		} else if page.PageSize > maxPageSize {
			page.PageSize = maxPageSize
		}
	}

	var (
		result dto.Page
		list   []T
	)

	offset := (page.PageNum - 1) * page.PageSize

	query := db.WithContext(ctx).Model(new(T))
	if fns != nil {
		for _, fn := range fns {
			query = fn(query)
		}
	}

	// Count total
	if err := query.Count(&page.Total).Error; err != nil {
		return result, err
	}

	// order
	if page.Sort != "" {
		var orderBy string
		if strings.ToLower(page.Order) == "desc" {
			orderBy = fmt.Sprintf("%s desc", page.Sort)
		} else {
			orderBy = fmt.Sprintf("%s asc", page.Sort)
		}
		query = query.Order(orderBy)
	}

	// Query data
	if err := query.Limit(page.PageSize).Offset(offset).Find(&list).Error; err != nil {
		return result, err
	}

	page.Data = list
	return page, nil
}
