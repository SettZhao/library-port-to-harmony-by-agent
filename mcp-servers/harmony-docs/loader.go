package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// ─────────────────────────────────────────────
// Data models
// ─────────────────────────────────────────────

// ModuleInfo represents one HarmonyOS API kit / module.
type ModuleInfo struct {
	Name       string `json:"name"`
	Directory  string `json:"directory"`
	ReadmePath string `json:"readme_path"`
}

// APIFile represents a single markdown API-reference file inside a module.
type APIFile struct {
	FileName      string `json:"file_name"`
	Title         string `json:"title"`
	Module        string `json:"module"`
	Kit           string `json:"kit"`
	Overview      string `json:"overview"`
	Since         string `json:"since"`
	RelatedMod    string `json:"related_module,omitempty"`
	Library       string `json:"library,omitempty"`
	SysCapability string `json:"system_capability,omitempty"`
}

// APIFunction represents one function / struct / enum entry extracted from a file.
type APIFunction struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Signature   string     `json:"signature,omitempty"`
	Since       string     `json:"since,omitempty"`
	Parameters  []APIParam `json:"parameters,omitempty"`
	Returns     string     `json:"returns,omitempty"`
}

// APIParam represents a single parameter of a function.
type APIParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// APIDetail holds the full parsed detail of one markdown file.
type APIDetail struct {
	APIFile
	Functions  []APIFunction `json:"functions,omitempty"`
	Structs    []APIFunction `json:"structs,omitempty"`
	Enums      []APIFunction `json:"enums,omitempty"`
	RawContent string        `json:"raw_content"`
}

// SearchResult is a single hit returned by search_api.
type SearchResult struct {
	Module   string `json:"module"`
	FileName string `json:"file_name"`
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
}

// ─────────────────────────────────────────────
// Cache
// ─────────────────────────────────────────────

type apiCache struct {
	mu      sync.RWMutex
	modules []ModuleInfo          // populated once on first list_api_modules call
	files   map[string][]APIFile  // key = module directory name
	details map[string]*APIDetail // key = "module/filename"
}

var cache = &apiCache{
	files:   make(map[string][]APIFile),
	details: make(map[string]*APIDetail),
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// docsRoot returns the absolute path to the harmony-API-reference directory.
func docsRoot() string {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	candidate := filepath.Join(dir, "harmony-API-reference")
	if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
		return candidate
	}
	// fallback: relative to working directory (dev mode)
	wd, _ := os.Getwd()
	return filepath.Join(wd, "harmony-API-reference")
}

// ─────────────────────────────────────────────
// Module listing
// ─────────────────────────────────────────────

var modulesDirRe = regexp.MustCompile(`\[(.+?)\]\((apis-[^/]+)/Readme-EN\.md\)`)

// loadModules parses the top-level Readme-EN.md to get the module list.
func loadModules() ([]ModuleInfo, error) {
	cache.mu.RLock()
	if cache.modules != nil {
		defer cache.mu.RUnlock()
		return cache.modules, nil
	}
	cache.mu.RUnlock()

	root := docsRoot()
	readmePath := filepath.Join(root, "Readme-EN.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return nil, err
	}

	var modules []ModuleInfo
	for _, match := range modulesDirRe.FindAllStringSubmatch(string(data), -1) {
		modules = append(modules, ModuleInfo{
			Name:       match[1],
			Directory:  match[2],
			ReadmePath: filepath.Join(root, match[2], "Readme-EN.md"),
		})
	}

	cache.mu.Lock()
	cache.modules = modules
	cache.mu.Unlock()
	return modules, nil
}

// ─────────────────────────────────────────────
// Module API files listing
// ─────────────────────────────────────────────

// loadModuleFiles returns light-weight metadata for every .md file in a module.
func loadModuleFiles(moduleDirName string) ([]APIFile, error) {
	cache.mu.RLock()
	if files, ok := cache.files[moduleDirName]; ok {
		cache.mu.RUnlock()
		return files, nil
	}
	cache.mu.RUnlock()

	root := docsRoot()
	modDir := filepath.Join(root, moduleDirName)
	entries, err := os.ReadDir(modDir)
	if err != nil {
		return nil, err
	}

	var files []APIFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "Readme-EN.md" {
			continue
		}
		fPath := filepath.Join(modDir, e.Name())
		af, err := parseFileHeader(fPath, moduleDirName, e.Name())
		if err != nil {
			continue
		}
		files = append(files, af)
	}

	cache.mu.Lock()
	cache.files[moduleDirName] = files
	cache.mu.Unlock()
	return files, nil
}

// ─────────────────────────────────────────────
// Full API detail
// ─────────────────────────────────────────────

// loadAPIDetail returns the full parsed detail of a single markdown file.
func loadAPIDetail(moduleDirName, fileName string) (*APIDetail, error) {
	cacheKey := moduleDirName + "/" + fileName

	cache.mu.RLock()
	if d, ok := cache.details[cacheKey]; ok {
		cache.mu.RUnlock()
		return d, nil
	}
	cache.mu.RUnlock()

	root := docsRoot()
	fPath := filepath.Join(root, moduleDirName, fileName)
	detail, err := parseFullFile(fPath, moduleDirName, fileName)
	if err != nil {
		return nil, err
	}

	cache.mu.Lock()
	cache.details[cacheKey] = detail
	cache.mu.Unlock()
	return detail, nil
}

// ─────────────────────────────────────────────
// Markdown parsing helpers
// ─────────────────────────────────────────────

var (
	kitRe     = regexp.MustCompile(`<!--Kit: (.+?)-->`)
	sinceRe   = regexp.MustCompile(`\*\*Since\*\*:\s*(.+)`)
	relatedRe = regexp.MustCompile(`\*\*Related module\*\*:\s*\[(.+?)\]`)
	libraryRe = regexp.MustCompile(`\*\*Library\*\*:\s*(.+)`)
	syscapRe  = regexp.MustCompile(`\*\*System capability\*\*:\s*(.+)`)
	h2Re      = regexp.MustCompile(`^## (.+)`)
	h3Re      = regexp.MustCompile(`^### (.+)`)
	codeRe    = regexp.MustCompile("^```")
)

// parseFileHeader reads just enough from a markdown file to populate APIFile.
func parseFileHeader(fPath, moduleDirName, fileName string) (APIFile, error) {
	data, err := os.ReadFile(fPath)
	if err != nil {
		return APIFile{}, err
	}
	text := string(data)

	af := APIFile{
		FileName: fileName,
		Module:   moduleDirName,
	}

	lines := strings.Split(text, "\n")
	// First non-empty line starting with # is the title
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "# ") {
			af.Title = strings.TrimPrefix(l, "# ")
			break
		}
	}

	if m := kitRe.FindStringSubmatch(text); m != nil {
		af.Kit = strings.TrimSpace(m[1])
	}
	if m := sinceRe.FindStringSubmatch(text); m != nil {
		af.Since = strings.TrimSpace(m[1])
	}
	if m := relatedRe.FindStringSubmatch(text); m != nil {
		af.RelatedMod = strings.TrimSpace(m[1])
	}
	if m := libraryRe.FindStringSubmatch(text); m != nil {
		af.Library = strings.TrimSpace(m[1])
	}
	if m := syscapRe.FindStringSubmatch(text); m != nil {
		af.SysCapability = strings.TrimSpace(m[1])
	}

	// Overview: paragraph after ## Overview
	if idx := strings.Index(text, "## Overview"); idx >= 0 {
		rest := text[idx+len("## Overview"):]
		af.Overview = extractFirstParagraph(rest)
	}

	return af, nil
}

// extractFirstParagraph returns the first non-empty paragraph from text.
func extractFirstParagraph(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var lines []string
	inPara := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inPara {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "<!--") {
			if inPara {
				break
			}
			continue
		}
		inPara = true
		lines = append(lines, trimmed)
	}
	return strings.Join(lines, " ")
}

// parseFullFile returns full APIDetail including functions/structs/enums.
func parseFullFile(fPath, moduleDirName, fileName string) (*APIDetail, error) {
	data, err := os.ReadFile(fPath)
	if err != nil {
		return nil, err
	}
	text := string(data)

	header, err := parseFileHeader(fPath, moduleDirName, fileName)
	if err != nil {
		return nil, err
	}

	detail := &APIDetail{
		APIFile:    header,
		RawContent: text,
	}

	// Extract functions from "## Function Description" section
	detail.Functions = extractItems(text, "## Function Description", h3Re)
	// Extract structs from "## Type Description" section
	detail.Structs = extractItems(text, "## Type Description", h3Re)
	// Extract enums from "## Enum Description" section
	detail.Enums = extractItems(text, "## Enum Description", h3Re)

	return detail, nil
}

// extractItems finds all H3 items under a given H2 section and parses them.
func extractItems(text, sectionHeader string, itemRe *regexp.Regexp) []APIFunction {
	idx := strings.Index(text, sectionHeader)
	if idx < 0 {
		return nil
	}
	// Find next H2 section
	sectionText := text[idx+len(sectionHeader):]
	nextH2 := h2Re.FindStringIndex(sectionText)
	if nextH2 != nil {
		sectionText = sectionText[:nextH2[0]]
	}

	var items []APIFunction
	lines := strings.Split(sectionText, "\n")
	var current *APIFunction
	var descLines []string
	var inCode bool
	var sigLines []string

	flush := func() {
		if current != nil {
			current.Description = strings.TrimSpace(strings.Join(descLines, " "))
			if len(sigLines) > 0 {
				current.Signature = strings.TrimSpace(strings.Join(sigLines, "\n"))
			}
			items = append(items, *current)
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if codeRe.MatchString(trimmed) {
			inCode = !inCode
			if inCode {
				sigLines = nil
			}
			continue
		}
		if inCode {
			sigLines = append(sigLines, line)
			continue
		}

		if m := itemRe.FindStringSubmatch(trimmed); m != nil {
			flush()
			current = &APIFunction{Name: strings.TrimSuffix(strings.TrimSpace(m[1]), "()")}
			descLines = nil
			sigLines = nil
			continue
		}

		if current != nil {
			// Capture **Since**: inside item
			if ms := sinceRe.FindStringSubmatch(trimmed); ms != nil {
				current.Since = strings.TrimSpace(ms[1])
				continue
			}
			// Capture **Description** paragraph
			if strings.HasPrefix(trimmed, "**Description**") {
				continue
			}
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") &&
				!strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "**Parameters**") &&
				!strings.HasPrefix(trimmed, "**Returns**") {
				descLines = append(descLines, trimmed)
			}
		}
	}
	flush()
	return items
}

// ─────────────────────────────────────────────
// Search
// ─────────────────────────────────────────────

// searchAPI walks every .md file in every module looking for keyword matches.
func searchAPI(keyword, moduleFilter string) ([]SearchResult, error) {
	modules, err := loadModules()
	if err != nil {
		return nil, err
	}

	keyword = strings.ToLower(keyword)
	root := docsRoot()
	var results []SearchResult

	for _, mod := range modules {
		if moduleFilter != "" && !strings.EqualFold(mod.Directory, moduleFilter) {
			continue
		}
		modDir := filepath.Join(root, mod.Directory)
		entries, err := os.ReadDir(modDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "Readme-EN.md" {
				continue
			}
			fPath := filepath.Join(modDir, e.Name())
			data, err := os.ReadFile(fPath)
			if err != nil {
				continue
			}
			content := string(data)
			lowerContent := strings.ToLower(content)
			if !strings.Contains(lowerContent, keyword) {
				continue
			}

			snippet := extractSnippet(content, keyword, 200)
			title := extractTitle(content)
			results = append(results, SearchResult{
				Module:   mod.Directory,
				FileName: e.Name(),
				Title:    title,
				Snippet:  snippet,
			})
		}
	}
	return results, nil
}

// extractSnippet returns up to maxLen chars of context around the first match.
func extractSnippet(content, keyword string, maxLen int) string {
	lower := strings.ToLower(content)
	idx := strings.Index(lower, strings.ToLower(keyword))
	if idx < 0 {
		return ""
	}
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + len(keyword) + 120
	if end > len(content) {
		end = len(content)
	}
	snippet := strings.TrimSpace(content[start:end])
	// Collapse whitespace / markdown noise
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	if len(snippet) > maxLen {
		snippet = snippet[:maxLen] + "..."
	}
	return snippet
}

// extractTitle returns the first H1 title from markdown text.
func extractTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}
