package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/taskflow/taskflow/internal/utils"
)

func TestNormalizePagination(t *testing.T) {
	cases := []struct {
		name                                  string
		inPage, inLimit                       int
		wantPage, wantLimit, wantOffset       int
	}{
		{"defaults when zero", 0, 0, 1, 20, 0},
		{"clamps oversized limit", 1, 5000, 1, 100, 0},
		{"computes offset", 3, 25, 3, 25, 50},
		{"negative falls back", -5, -1, 1, 20, 0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, l, o := utils.NormalizePagination(tc.inPage, tc.inLimit)
			require.Equal(t, tc.wantPage, p)
			require.Equal(t, tc.wantLimit, l)
			require.Equal(t, tc.wantOffset, o)
		})
	}
}

func TestSortClause_Whitelist(t *testing.T) {
	allowed := map[string]string{
		"created_at":      "t.created_at ASC",
		"created_at_desc": "t.created_at DESC",
	}
	require.Equal(t, "t.created_at DESC", utils.SortClause("", allowed, "t.created_at DESC"))
	require.Equal(t, "t.created_at ASC", utils.SortClause("created_at", allowed, "fallback"))

	require.Equal(t, "fallback", utils.SortClause("1; DROP TABLE users;--", allowed, "fallback"))
}
