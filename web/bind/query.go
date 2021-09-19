package bind

type Query struct {
	Search string `json:"search" form:"search"`
}
