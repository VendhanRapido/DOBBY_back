package entitiesclient

type UserRoles struct {
	Roles []string `json:"roles"`
}

type GetUserResponse struct {
	Data UserRoles `json:"data"`
}
