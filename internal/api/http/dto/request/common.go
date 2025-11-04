package request

type PaginationRequest struct {
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
	Offset    int    `form:"offset" binding:"omitempty,min=0" example:"0"`
	SortBy    string `form:"sort_by" binding:"omitempty" example:"created_at"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// Default values for pagination
const (
	DefaultLimit     = 20
	MaxLimit         = 100
	DefaultOffset    = 0
	DefaultSortOrder = "desc"
)

// ApplyDefaults sets default values for pagination
func (p *PaginationRequest) ApplyDefaults(defaultSortBy string) {
	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.Offset < 0 {
		p.Offset = DefaultOffset
	}
	if p.SortBy == "" {
		p.SortBy = defaultSortBy
	}
	if p.SortOrder == "" {
		p.SortOrder = DefaultSortOrder
	}
}
