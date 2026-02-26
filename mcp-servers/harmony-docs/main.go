package main

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"harmony-docs-mcp-server",
		"v1.0.0",
		server.WithToolCapabilities(false),
	)

	// ── Tool 1: list_api_modules ──────────────────────────────────────────
	// Returns all available HarmonyOS API kits/modules discovered from the
	// top-level Readme-EN.md index.
	toolListModules := mcp.NewTool("list_api_modules",
		mcp.WithDescription(
			"List all available HarmonyOS API kits/modules. "+
				"Returns a JSON array with each module's name and directory. "+
				"Call this first to discover which modules are available before "+
				"querying individual APIs."),
	)
	mcpServer.AddTool(toolListModules, handleListAPIModules)

	// ── Tool 2: get_module_apis ───────────────────────────────────────────
	// Returns light-weight metadata for every API file inside one kit.
	toolGetModuleAPIs := mcp.NewTool("get_module_apis",
		mcp.WithDescription(
			"Get a structured list of all API files in a specific HarmonyOS kit/module. "+
				"Returns file names, titles, kit name, Since version, Library, and "+
				"SystemCapability for each file. Use list_api_modules first to obtain "+
				"valid module_dir values."),
		mcp.WithString("module_dir",
			mcp.Required(),
			mcp.Description(
				"The kit directory name, e.g. 'apis-ability-kit', 'apis-audio-kit'. "+
					"Use list_api_modules to get the full list of valid values.")),
	)
	mcpServer.AddTool(toolGetModuleAPIs, handleGetModuleAPIs)

	// ── Tool 3: get_api_detail ────────────────────────────────────────────
	// Returns fully parsed detail for one markdown file.
	toolGetAPIDetail := mcp.NewTool("get_api_detail",
		mcp.WithDescription(
			"Get detailed information for a specific HarmonyOS API file, including "+
				"overview, all functions with signatures and parameter descriptions, "+
				"structs, enums, Since version, Library, and SystemCapability. "+
				"Use get_module_apis to obtain valid file_name values."),
		mcp.WithString("module_dir",
			mcp.Required(),
			mcp.Description("The kit directory name, e.g. 'apis-ability-kit'.")),
		mcp.WithString("file_name",
			mcp.Required(),
			mcp.Description(
				"The markdown file name inside the module, e.g. "+
					"'capi-ability-access-control-h.md'.")),
		mcp.WithBoolean("include_raw",
			mcp.Description(
				"If true, includes the full raw markdown content in the response. "+
					"Defaults to false to keep the response concise.")),
	)
	mcpServer.AddTool(toolGetAPIDetail, handleGetAPIDetail)

	// ── Tool 4: search_api ────────────────────────────────────────────────
	// Full-text search across all (or one) module's markdown files.
	toolSearchAPI := mcp.NewTool("search_api",
		mcp.WithDescription(
			"Search for a keyword across all HarmonyOS API reference files. "+
				"Returns matching files with context snippets. "+
				"Optionally scope the search to a single module."),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description(
				"The keyword or API name to search for, e.g. 'OH_AT_CheckSelfPermission', "+
					"'bluetooth', 'camera'.")),
		mcp.WithString("module_dir",
			mcp.Description(
				"Optional: restrict search to a specific kit directory, e.g. "+
					"'apis-audio-kit'. Omit to search all modules.")),
		mcp.WithNumber("max_results",
			mcp.Description(
				"Maximum number of results to return. Defaults to 20.")),
	)
	mcpServer.AddTool(toolSearchAPI, handleSearchAPI)

	// Start stdio transport (required for VS Code MCP Agent integration)
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
