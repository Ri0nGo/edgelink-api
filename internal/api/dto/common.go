package dto

type ReqId struct {
	Id int `json:"id"`
}

type Page struct {
	Total    int64  `json:"total" form:"total"`
	PageSize int    `json:"page_size" form:"page_size"`
	PageNum  int    `json:"page_num" form:"page_num"`
	Data     any    `json:"data" form:"data"`
	Order    string `json:"order" form:"order"`
	Sort     string `json:"sort" form:"sort"`

	// ext
	UnlimitedPageSize bool `json:"-" form:"-"` // 不限制分页
}
