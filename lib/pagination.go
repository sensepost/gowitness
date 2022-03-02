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
	ShowHidden    bool
	OnlyShowNotes bool
	FiltTagMap    *map[interface{}]interface{}
	FiltRespCodes *map[string]bool
}

// Filter describes a column filter
type Filter struct {
	Column string
	Value  interface{}
	Oper   string
}

// Pagination has options for a Page
type Pagination struct {
	DB       *gorm.DB
	CurrPage int
	Limit    int
	OrderBy  []string
	FilterBy []Filter
	JoinsBy  []Filter
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

	// joins query
	if len(p.JoinsBy) > 0 {
		for _, filter := range p.JoinsBy{
			db = db.Joins("LEFT JOIN " + filter.Column + " on " + filter.Value.(string))
		}
	}

	// where query
	if len(p.FilterBy) > 0 {
		for _, filter := range p.FilterBy {
			if filter.Oper == "LIKE"{
				db = db.Where(filter.Column + " " + filter.Oper + " ?", "%" + filter.Value.(string) + "%")
			} else {
				db = db.Where(filter.Column + " " + filter.Oper + " ?", filter.Value)
			}
		}
	}
	// Need a specific select in order to prevent Model() from pulling the wrong IDs. This is probably due to an error
	// with the way the join is currently built. Maybe preload w/ an inline where clause could fix this
	db = db.Select("urls.id,urls.url, urls.final_url, urls.response_code, urls.response_reason, urls.proto, urls.content_length,urls.title, urls.filename, urls.perception_hash")
	
	db.Model(data).Count(&count)

	if p.CurrPage == 1 {
		offset = 0
	} else {
		offset = (p.CurrPage - 1) * p.Limit
	}

	if err := db.Limit(p.Limit).Offset(offset).Preload("Technologies").Preload("Filter").Preload("Filter.Tagmaps").Find(data).Error; err != nil {
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
		pagination.PrevPageRange = makeSizedRange(1, pagination.NextPage-1, 5)
		pagination.NextPageRange = []int{}
	} else {
		pagination.NextPage = p.CurrPage + 1
		pagination.PrevPageRange = makeSizedRange(1, pagination.NextPage-2, 5)
		pagination.NextPageRange = makeSizedRange(pagination.NextPage, pagination.Pages, 5)
	}
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