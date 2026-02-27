package dto

type ReqId struct {
	Id int `json:"id"`
}

type ReqIds struct {
	Ids []int `json:"ids"`
}

type BeginAndEnd struct {
	Begin int64 `json:"begin"`
	End   int64 `json:"end"`
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

type ReqPageSearch struct {
	Page
	ModelId int    `json:"model_id" form:"model_id"`
	Search  string `json:"search" form:"search"`
}
