package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─────────────────────────────────────────────
// list_api_modules handler
// ─────────────────────────────────────────────

func handleListAPIModules(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	modules, err := loadModules()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load modules: %v", err)), nil
	}

	type response struct {
		TotalModules int          `json:"total_modules"`
		Modules      []ModuleInfo `json:"modules"`
	}
	out := response{
		TotalModules: len(modules),
		Modules:      modules,
	}
	return jsonResult(out)
}

// ─────────────────────────────────────────────
// get_module_apis handler
// ─────────────────────────────────────────────

func handleGetModuleAPIs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleDirName, err := req.RequireString("module_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	files, err := loadModuleFiles(moduleDirName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load module %q: %v", moduleDirName, err)), nil
	}

	type response struct {
		Module    string    `json:"module"`
		TotalAPIs int       `json:"total_apis"`
		APIs      []APIFile `json:"apis"`
	}
	out := response{
		Module:    moduleDirName,
		TotalAPIs: len(files),
		APIs:      files,
	}
	return jsonResult(out)
}

// ─────────────────────────────────────────────
// get_api_detail handler
// ─────────────────────────────────────────────

func handleGetAPIDetail(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleDirName, err := req.RequireString("module_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fileName, err := req.RequireString("file_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Optional: whether to include raw markdown in the response.
	includeRaw := false
	if v, ok := req.GetArguments()["include_raw"]; ok {
		if b, ok := v.(bool); ok {
			includeRaw = b
		}
	}

	detail, err := loadAPIDetail(moduleDirName, fileName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load API detail for %q/%q: %v", moduleDirName, fileName, err)), nil
	}

	// Return a copy, optionally stripping raw content to keep response lean.
	result := *detail
	if !includeRaw {
		result.RawContent = ""
	}
	return jsonResult(result)
}

// ─────────────────────────────────────────────
// search_api handler
// ─────────────────────────────────────────────

func handleSearchAPI(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := req.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Optional module filter
	moduleFilter := ""
	if v, ok := req.GetArguments()["module_dir"]; ok {
		if s, ok := v.(string); ok {
			moduleFilter = s
		}
	}

	// Optional max results (default 20)
	maxResults := 20
	if v, ok := req.GetArguments()["max_results"]; ok {
		switch n := v.(type) {
		case float64:
			maxResults = int(n)
		case int:
			maxResults = n
		}
	}

	results, err := searchAPI(keyword, moduleFilter)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// Trim to max_results
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	type response struct {
		Keyword    string         `json:"keyword"`
		TotalFound int            `json:"total_found"`
		Results    []SearchResult `json:"results"`
	}
	out := response{
		Keyword:    keyword,
		TotalFound: len(results),
		Results:    results,
	}
	return jsonResult(out)
}

// ─────────────────────────────────────────────
// Shared helper
// ─────────────────────────────────────────────

// jsonResult serialises v to a pretty-printed JSON MCP tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
