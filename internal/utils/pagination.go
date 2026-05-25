package utils

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

func NormalizePagination(page, limit int) (normPage, normLimit, offset int) {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit, (page - 1) * limit
}

func SortClause(input string, allowed map[string]string, fallback string) string {
	if input == "" {
		return fallback
	}
	if clause, ok := allowed[input]; ok {
		return clause
	}
	return fallback
}
