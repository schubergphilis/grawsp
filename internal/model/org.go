package model

type Org struct {
	Name        string            `json:"name"`
	Region      string            `json:"region"`
	StartUrl    string            `json:"start_url"`
	Roles       map[string]string `json:"roles"`
	DefaultRole string            `json:"default_role"`
}
