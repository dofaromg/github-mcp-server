package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	ghErrors "github.com/github/github-mcp-server/pkg/errors"
	"github.com/github/github-mcp-server/pkg/utils"
	"github.com/google/go-github/v79/github"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// hasFilter checks if a query string contains a filter of the specified type.
// It matches filter at start of string, after whitespace, or after non-word characters like '('.
// This implementation uses string operations instead of regex for better performance.
func hasFilter(query, filterType string) bool {
	// Build the filter prefix we're looking for (e.g., "is:", "repo:")
	prefix := filterType + ":"

	// Check if the filter appears anywhere in the query
	idx := strings.Index(query, prefix)
	if idx == -1 {
		return false
	}

	// Keep checking all occurrences
	for idx != -1 {
		// Check if this is a valid filter position:
		// - At start of string (idx == 0)
		// - After whitespace or non-word character
		if idx == 0 {
			// At start - need to check if followed by non-whitespace
			if idx+len(prefix) < len(query) && !unicode.IsSpace(rune(query[idx+len(prefix)])) {
				return true
			}
		} else {
			prevChar := rune(query[idx-1])
			// Valid if preceded by whitespace or non-word character (like '(' or other punctuation)
			if unicode.IsSpace(prevChar) || (!unicode.IsLetter(prevChar) && !unicode.IsDigit(prevChar) && prevChar != '_') {
				// Also need to check if followed by non-whitespace (actual value)
				if idx+len(prefix) < len(query) && !unicode.IsSpace(rune(query[idx+len(prefix)])) {
					return true
				}
			}
		}

		// Look for next occurrence - search after the current prefix
		remaining := query[idx+len(prefix):]
		nextIdx := strings.Index(remaining, prefix)
		if nextIdx == -1 {
			break
		}
		idx = idx + len(prefix) + nextIdx
	}

	return false
}

// hasSpecificFilter checks if a query string contains a specific filter:value pair.
// It matches at start, after whitespace, or after non-word characters, and ensures
// the value ends with a word boundary, whitespace, or non-word character.
// This implementation uses string operations instead of regex for better performance.
func hasSpecificFilter(query, filterType, filterValue string) bool {
	// Build the exact filter:value we're looking for (e.g., "is:issue")
	target := filterType + ":" + filterValue

	// Check if the target appears anywhere in the query
	idx := strings.Index(query, target)
	if idx == -1 {
		return false
	}

	// Keep checking all occurrences
	for idx != -1 {
		// Check if this is a valid position:
		// - At start of string (idx == 0)
		// - After whitespace or non-word character
		validStart := false
		if idx == 0 {
			validStart = true
		} else {
			prevChar := rune(query[idx-1])
			validStart = unicode.IsSpace(prevChar) || (!unicode.IsLetter(prevChar) && !unicode.IsDigit(prevChar) && prevChar != '_')
		}

		if validStart {
			// Check if followed by end of string, whitespace, or non-word character
			endIdx := idx + len(target)
			if endIdx == len(query) {
				return true
			}
			nextChar := rune(query[endIdx])
			if unicode.IsSpace(nextChar) || (!unicode.IsLetter(nextChar) && !unicode.IsDigit(nextChar) && nextChar != '_') {
				return true
			}
		}

		// Look for next occurrence - search after the current target
		remaining := query[idx+len(target):]
		nextIdx := strings.Index(remaining, target)
		if nextIdx == -1 {
			break
		}
		idx = idx + len(target) + nextIdx
	}

	return false
}

func hasRepoFilter(query string) bool {
	return hasFilter(query, "repo")
}

func hasTypeFilter(query string) bool {
	return hasFilter(query, "type")
}

func searchHandler(
	ctx context.Context,
	getClient GetClientFn,
	args map[string]any,
	searchType string,
	errorPrefix string,
) (*mcp.CallToolResult, error) {
	query, err := RequiredParam[string](args, "query")
	if err != nil {
		return utils.NewToolResultError(err.Error()), nil
	}

	if !hasSpecificFilter(query, "is", searchType) {
		query = fmt.Sprintf("is:%s %s", searchType, query)
	}

	owner, err := OptionalParam[string](args, "owner")
	if err != nil {
		return utils.NewToolResultError(err.Error()), nil
	}

	repo, err := OptionalParam[string](args, "repo")
	if err != nil {
		return utils.NewToolResultError(err.Error()), nil
	}

	if owner != "" && repo != "" && !hasRepoFilter(query) {
		query = fmt.Sprintf("repo:%s/%s %s", owner, repo, query)
	}

	sort, err := OptionalParam[string](args, "sort")
	if err != nil {
		return utils.NewToolResultError(err.Error()), nil
	}
	order, err := OptionalParam[string](args, "order")
	if err != nil {
		return utils.NewToolResultError(err.Error()), nil
	}
	pagination, err := OptionalPaginationParams(args)
	if err != nil {
		return utils.NewToolResultError(err.Error()), nil
	}

	opts := &github.SearchOptions{
		// Default to "created" if no sort is provided, as it's a common use case.
		Sort:  sort,
		Order: order,
		ListOptions: github.ListOptions{
			Page:    pagination.Page,
			PerPage: pagination.PerPage,
		},
	}

	client, err := getClient(ctx)
	if err != nil {
		return utils.NewToolResultErrorFromErr(errorPrefix+": failed to get GitHub client", err), nil
	}
	result, resp, err := client.Search.Issues(ctx, query, opts)
	if err != nil {
		return utils.NewToolResultErrorFromErr(errorPrefix, err), nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewToolResultErrorFromErr(errorPrefix+": failed to read response body", err), nil
		}
		return ghErrors.NewGitHubAPIStatusErrorResponse(ctx, errorPrefix, resp, body), nil
	}

	r, err := json.Marshal(result)
	if err != nil {
		return utils.NewToolResultErrorFromErr(errorPrefix+": failed to marshal response", err), nil
	}

	return utils.NewToolResultText(string(r)), nil
}
