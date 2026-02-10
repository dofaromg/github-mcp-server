package github

import (
	"testing"
)

// Benchmark tests for hasFilter and hasSpecificFilter functions to demonstrate
// the performance improvement from using string operations instead of regex.

func BenchmarkHasFilter(b *testing.B) {
	testCases := []struct {
		name       string
		query      string
		filterType string
	}{
		{
			name:       "simple query",
			query:      "is:issue bug report",
			filterType: "is",
		},
		{
			name:       "complex query with OR",
			query:      "repo:github/github-mcp-server is:issue (label:critical OR label:urgent)",
			filterType: "is",
		},
		{
			name:       "long query",
			query:      "repo:github/github-mcp-server is:issue label:bug label:enhancement author:octocat created:>2024-01-01 updated:<2024-12-31",
			filterType: "repo",
		},
		{
			name:       "filter not present",
			query:      "some text with no filter present in this query at all",
			filterType: "is",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = hasFilter(tc.query, tc.filterType)
			}
		})
	}
}

func BenchmarkHasSpecificFilter(b *testing.B) {
	testCases := []struct {
		name        string
		query       string
		filterType  string
		filterValue string
	}{
		{
			name:        "simple query",
			query:       "is:issue bug report",
			filterType:  "is",
			filterValue: "issue",
		},
		{
			name:        "complex query with OR",
			query:       "repo:github/github-mcp-server is:issue (label:critical OR label:urgent)",
			filterType:  "is",
			filterValue: "issue",
		},
		{
			name:        "long query",
			query:       "repo:github/github-mcp-server is:issue label:bug label:enhancement author:octocat created:>2024-01-01 updated:<2024-12-31",
			filterType:  "is",
			filterValue: "issue",
		},
		{
			name:        "filter not present",
			query:       "some text with no specific filter value in this query",
			filterType:  "is",
			filterValue: "issue",
		},
		{
			name:        "multiple is: filters",
			query:       "is:issue is:open is:closed bug report",
			filterType:  "is",
			filterValue: "issue",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = hasSpecificFilter(tc.query, tc.filterType, tc.filterValue)
			}
		})
	}
}

func BenchmarkHasRepoFilter(b *testing.B) {
	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "with repo filter",
			query: "repo:github/github-mcp-server is:issue bug",
		},
		{
			name:  "without repo filter",
			query: "is:issue bug report critical",
		},
		{
			name:  "complex query",
			query: "repo:github/github-mcp-server is:issue (label:critical OR label:urgent) author:octocat",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = hasRepoFilter(tc.query)
			}
		})
	}
}

func BenchmarkHasTypeFilter(b *testing.B) {
	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "with type filter",
			query: "type:user location:seattle followers:>50",
		},
		{
			name:  "without type filter",
			query: "location:seattle followers:>50",
		},
		{
			name:  "complex query",
			query: "type:user (location:seattle OR location:california) followers:>100",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = hasTypeFilter(tc.query)
			}
		})
	}
}
