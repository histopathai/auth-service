package request

type ListUsersRequest struct {
	PaginationRequest
}

const DefaultUserSortBy = "created_at"

func (r *ListUsersRequest) ApplyDefaults() {
	if r.SortBy == nil {
		defaultSort := DefaultUserSortBy
		r.SortBy = &defaultSort
	}

	r.PaginationRequest.ApplyDefaults()
}

func (r *ListUsersRequest) GetAllowedSortFields() []string {
	return []string{"created_at", "updated_at", "email", "display_name"}
}
