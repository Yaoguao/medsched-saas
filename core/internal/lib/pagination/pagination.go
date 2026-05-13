package pagination

const (
	DefaultLimit = 50
	MaxLimit     = 100
)

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func Normalize(page Page) Page {
	if page.Limit <= 0 {
		page.Limit = DefaultLimit
	}
	if page.Limit > MaxLimit {
		page.Limit = MaxLimit
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return page
}
