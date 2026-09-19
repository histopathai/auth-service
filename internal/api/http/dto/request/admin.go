package request

type ListUsersRequest struct {
	PaginationRequest
}

// ApproveUserRequest carries the group a newly approved user joins. The body
// is optional: reactivating a suspended user needs none, they keep their role.
// The value is checked by the service, after legacy names are normalized, so a
// browser still running the old bundle (which sends "user") is not refused.
type ApproveUserRequest struct {
	Role string `json:"role" example:"pathologist" enums:"pathologist,datascientist"`
}

// SetUserRoleRequest moves an active user to another group.
type SetUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin pathologist datascientist" example:"datascientist"`
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
