package lib

import (
	"math"

	"gorm.io/gorm"
)

// PaginationPage is a sinlge, paginated page
type PaginationPage struct {
	Count         int64
	Pages         int
	Records       interface{}
	Offset        int
	Range         int
	Limit         int
	Page          int
	PrevPage      int
	PrevPageRange []int
	NextPage      int
	NextPageRange []int
	Ordered       bool
}

// Filter describes a column filter
type Filter struct {
	Column string
	Value  string
}

// Pagination has options for a Page
type Pagination struct {
	DB       *gorm.DB
	CurrPage int
	Limit    int
	OrderBy  []string
	FilterBy []Filter
}

// Page pages a dataset
func (p *Pagination) Page(data interface{}) (*PaginationPage, error) {

	var pagination PaginationPage
	var count int64
	var offset int

	db := p.DB

	if p.CurrPage < 1 {
		p.CurrPage = 1
	}
	if p.Limit == 0 {
		p.Limit = 21
	}
	if len(p.OrderBy) > 0 {
		for _, order := range p.OrderBy {
			db = db.Order(order)
		}
		pagination.Ordered = true
	} else {
		pagination.Ordered = false
	}

	if len(p.FilterBy) > 0 {
		for _, filter := range p.FilterBy {
			db = db.Where(filter.Column+" LIKE ?", "%"+filter.Value+"%")
		}
	}

	db.Model(data).Count(&count)

	if p.CurrPage == 1 {
		offset = 0
	} else {
		offset = (p.CurrPage - 1) * p.Limit
	}

	if err := db.Limit(p.Limit).Offset(offset).Preload("Technologies").Find(data).Error; err != nil {
		return nil, err
	}

	pagination.Count = count
	pagination.Records = data
	pagination.Page = p.CurrPage

	pagination.Offset = offset
	pagination.Limit = p.Limit
	pagination.Pages = int(math.Ceil(float64(count) / float64(p.Limit)))
	pagination.Range = pagination.Offset + pagination.Limit

	if p.CurrPage > 1 {
		pagination.PrevPage = p.CurrPage - 1
	} else {
		pagination.PrevPage = p.CurrPage
	}

	if p.CurrPage >= pagination.Pages {
		pagination.NextPage = p.CurrPage
	} else {
		pagination.NextPage = p.CurrPage + 1
	}

	pagination.PrevPageRange = makeSizedRange(1, pagination.NextPage-2, 5)
	pagination.NextPageRange = makeSizedRange(pagination.NextPage, pagination.Pages, 5)

	return &pagination, nil
}

func makeSizedRange(min, max, l int) []int {
	if min > max {
		return []int{}
	}

	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}

	return a
}
