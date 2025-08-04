package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// JSONLFile represents a loaded JSONL file with metadata
type JSONLFile struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	Records    int       `json:"records"`
	LoadedAt   time.Time `json:"loadedAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// JSONRecord represents a single JSON record from a JSONL file
type JSONRecord struct {
	LineNumber int                    `json:"lineNumber"`
	Content    map[string]interface{} `json:"content"`
	RawJSON    string                 `json:"rawJSON"`
}

// FileStats provides detailed statistics about a JSONL file
type FileStats struct {
	TotalLines   int      `json:"totalLines"`
	ValidRecords int      `json:"validRecords"`
	InvalidLines []int    `json:"invalidLines"`
	CommonFields []string `json:"commonFields"`
	FileSize     int64    `json:"fileSize"`
}

// SearchOptions defines parameters for searching through records
type SearchOptions struct {
	Query         string `json:"query"`
	CaseSensitive bool   `json:"caseSensitive"`
	UseLucene     bool   `json:"useLucene"`
	SelectedField string `json:"selectedField"`
	Offset        int    `json:"offset"`
	Limit         int    `json:"limit"`
}

// LuceneQuery represents a parsed Lucene query
type LuceneQuery struct {
	Type  string       `json:"type"` // 'term', 'field', 'and', 'or', 'not', 'wildcard', 'phrase'
	Field string       `json:"field,omitempty"`
	Value string       `json:"value,omitempty"`
	Left  *LuceneQuery `json:"left,omitempty"`
	Right *LuceneQuery `json:"right,omitempty"`
	Query *LuceneQuery `json:"query,omitempty"`
}

// SearchResult represents a search result with highlighting information
type SearchResult struct {
	Records      []JSONRecord `json:"records"`
	Offset       int          `json:"offset"`
	Limit        int          `json:"limit"`
	Total        int          `json:"total"`
	TotalMatches int          `json:"totalMatches"`
	HasMore      bool         `json:"hasMore"`
	Query        string       `json:"query"`
}

// ExportData represents the data structure for exporting search results
type ExportData struct {
	Metadata Metadata `json:"metadata"`
	Records  []Record `json:"records"`
}

type Record struct {
	LineNumber  int                    `json:"lineNumber"`
	Content     map[string]interface{} `json:"content"`
	RawJSON     string                 `json:"rawJSON"`
	DisplayJSON string                 `json:"displayJSON"`
}

type Metadata struct {
	Timestamp       string  `json:"timestamp"`
	FileName        string  `json:"fileName"`
	TotalRecords    int     `json:"totalRecords"`
	SearchQuery     *string `json:"searchQuery"`
	FieldVisibility struct {
		ShownFields  []string `json:"shownFields"`
		HiddenFields []string `json:"hiddenFields"`
	} `json:"fieldVisibility"`
	Pagination struct {
		CurrentPage    int `json:"currentPage"`
		RecordsPerPage int `json:"recordsPerPage"`
		TotalPages     int `json:"totalPages"`
	} `json:"pagination"`
}

// HighlightMatch represents a text match with highlighting information
type HighlightMatch struct {
	Text      string `json:"text"`
	StartPos  int    `json:"startPos"`
	EndPos    int    `json:"endPos"`
	FieldName string `json:"fieldName"`
}

// Custom error types for JSONL operations
var (
	ErrInvalidJSONL   = errors.New("invalid JSONL format")
	ErrFileNotFound   = errors.New("file not found or cannot be accessed")
	ErrClipboardEmpty = errors.New("no data in clipboard")
	ErrParsingFailed  = errors.New("JSON parsing failed")
	ErrInvalidLineNum = errors.New("invalid line number")
	ErrNoFileLoaded   = errors.New("no file currently loaded")
)

// JSONLError provides detailed error information with line numbers
type JSONLError struct {
	Message    string `json:"message"`
	LineNumber int    `json:"lineNumber"`
	Line       string `json:"line"`
	Err        error  `json:"-"`
}

func (e *JSONLError) Error() string {
	if e.LineNumber > 0 {
		return fmt.Sprintf("%s at line %d: %s", e.Message, e.LineNumber, e.Line)
	}
	return e.Message
}

// RecordCache provides efficient caching for record retrieval
type RecordCache struct {
	records    []JSONRecord
	pageSize   int
	totalCount int
}

// PaginatedRecords represents a paginated response of records
type PaginatedRecords struct {
	Records []JSONRecord `json:"records"`
	Offset  int          `json:"offset"`
	Limit   int          `json:"limit"`
	Total   int          `json:"total"`
	HasMore bool         `json:"hasMore"`
}

// App struct
type App struct {
	ctx         context.Context
	currentFile *JSONLFile
	records     []JSONRecord
	cache       *RecordCache
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// JSONLParser handles parsing of JSONL files
type JSONLParser struct {
	file      *os.File
	scanner   *bufio.Scanner
	lineCount int
}

// NewJSONLParser creates a new JSONL parser for the given file path
func NewJSONLParser(filePath string) (*JSONLParser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, &JSONLError{
			Message: "Failed to open file",
			Err:     ErrFileNotFound,
		}
	}

	scanner := bufio.NewScanner(file)
	return &JSONLParser{
		file:      file,
		scanner:   scanner,
		lineCount: 0,
	}, nil
}

// Close closes the file and cleans up resources
func (p *JSONLParser) Close() error {
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

// ParseJSONL parses the entire JSONL file and returns all records
func (p *JSONLParser) ParseJSONL() ([]JSONRecord, *FileStats, error) {
	var records []JSONRecord
	var invalidLines []int
	fieldCounts := make(map[string]int)
	totalRecords := 0

	// Get file info for size
	fileInfo, err := p.file.Stat()
	if err != nil {
		return nil, nil, &JSONLError{
			Message: "Failed to get file information",
			Err:     err,
		}
	}

	for p.scanner.Scan() {
		p.lineCount++
		line := strings.TrimSpace(p.scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Try to parse the JSON line
		var content map[string]interface{}
		if err := json.Unmarshal([]byte(line), &content); err != nil {
			invalidLines = append(invalidLines, p.lineCount)
			continue
		}

		// Count fields for common fields analysis
		for field := range content {
			fieldCounts[field]++
		}

		// Create record
		record := JSONRecord{
			LineNumber: p.lineCount,
			Content:    content,
			RawJSON:    line,
		}
		records = append(records, record)
		totalRecords++
	}

	// Check for scanner errors
	if err := p.scanner.Err(); err != nil {
		return nil, nil, &JSONLError{
			Message: "Error reading file",
			Err:     err,
		}
	}

	// Calculate common fields (fields that appear in at least 50% of records)
	var commonFields []string
	threshold := totalRecords / 2
	for field, count := range fieldCounts {
		if count >= threshold {
			commonFields = append(commonFields, field)
		}
	}

	stats := &FileStats{
		TotalLines:   p.lineCount,
		ValidRecords: totalRecords,
		InvalidLines: invalidLines,
		CommonFields: commonFields,
		FileSize:     fileInfo.Size(),
	}

	return records, stats, nil
}

// ValidateJSONLLine validates a single line of JSONL format
func ValidateJSONLLine(line string, lineNumber int) error {
	line = strings.TrimSpace(line)

	// Empty lines are allowed
	if line == "" {
		return nil
	}

	// Try to parse as JSON
	var content map[string]interface{}
	if err := json.Unmarshal([]byte(line), &content); err != nil {
		return &JSONLError{
			Message:    "Invalid JSON format",
			LineNumber: lineNumber,
			Line:       line,
			Err:        ErrParsingFailed,
		}
	}

	return nil
}

// ParseJSONLFromString parses JSONL content from a string (useful for clipboard)
func ParseJSONLFromString(content string) ([]JSONRecord, *FileStats, error) {
	var records []JSONRecord
	var invalidLines []int
	fieldCounts := make(map[string]int)
	totalRecords := 0

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNumber := i + 1
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Try to parse the JSON line
		var jsonContent map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonContent); err != nil {
			invalidLines = append(invalidLines, lineNumber)
			continue
		}

		// Count fields for common fields analysis
		for field := range jsonContent {
			fieldCounts[field]++
		}

		// Create record
		record := JSONRecord{
			LineNumber: lineNumber,
			Content:    jsonContent,
			RawJSON:    line,
		}
		records = append(records, record)
		totalRecords++
	}

	// Calculate common fields (fields that appear in at least 50% of records)
	var commonFields []string
	threshold := totalRecords / 2
	for field, count := range fieldCounts {
		if count >= threshold {
			commonFields = append(commonFields, field)
		}
	}

	stats := &FileStats{
		TotalLines:   len(lines),
		ValidRecords: totalRecords,
		InvalidLines: invalidLines,
		CommonFields: commonFields,
		FileSize:     int64(len(content)),
	}

	return records, stats, nil
}

// OpenFile opens a native file dialog and returns the selected file path
func (a *App) OpenFile() (string, error) {
	// Configure file dialog options
	dialogOptions := runtime.OpenDialogOptions{
		Title: "Select JSONL File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSONL Files (*.jsonl)",
				Pattern:     "*.jsonl",
			},
			{
				DisplayName: "JSON Lines Files (*.jsonlines)",
				Pattern:     "*.jsonlines",
			},
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
			{
				DisplayName: "All Files",
				Pattern:     "*",
			},
		},
	}

	// Open the file dialog
	filePath, err := runtime.OpenFileDialog(a.ctx, dialogOptions)
	if err != nil {
		return "", &JSONLError{
			Message: "Failed to open file dialog",
			Err:     err,
		}
	}

	// Check if user cancelled the dialog
	if filePath == "" {
		return "", &JSONLError{
			Message: "File selection cancelled",
			Err:     errors.New("user cancelled file selection"),
		}
	}

	return filePath, nil
}

// LoadJSONLFile loads and parses a JSONL file from the given file path
func (a *App) LoadJSONLFile(filePath string) (*JSONLFile, error) {
	// Validate file path
	if filePath == "" {
		return nil, &JSONLError{
			Message: "File path cannot be empty",
			Err:     ErrFileNotFound,
		}
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, &JSONLError{
			Message: "File not found or cannot be accessed",
			Err:     ErrFileNotFound,
		}
	}

	// Create parser
	parser, err := NewJSONLParser(filePath)
	if err != nil {
		return nil, err
	}
	defer parser.Close()

	// Parse the file
	records, stats, err := parser.ParseJSONL()
	if err != nil {
		return nil, err
	}

	// Create JSONLFile metadata
	fileName := filepath.Base(filePath)
	jsonlFile := &JSONLFile{
		Name:       fileName,
		Path:       filePath,
		Size:       fileInfo.Size(),
		Records:    stats.ValidRecords,
		LoadedAt:   time.Now(),
		ModifiedAt: fileInfo.ModTime(),
	}

	// Store in app state
	a.currentFile = jsonlFile
	a.records = records

	// Initialize cache for efficient pagination
	a.cache = &RecordCache{
		records:    records,
		pageSize:   50, // Default page size for virtual scrolling
		totalCount: len(records),
	}

	return jsonlFile, nil
}

// GetFileStats returns detailed statistics about the currently loaded file
func (a *App) GetFileStats() (*FileStats, error) {
	if a.currentFile == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	// Re-parse the file to get fresh statistics
	parser, err := NewJSONLParser(a.currentFile.Path)
	if err != nil {
		return nil, err
	}
	defer parser.Close()

	_, stats, err := parser.ParseJSONL()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// CheckFileModification checks if the currently loaded file has been modified since it was loaded
func (a *App) CheckFileModification() (bool, error) {
	if a.currentFile == nil {
		return false, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	// Skip modification check for clipboard content
	if a.currentFile.Path == "<clipboard>" {
		return false, nil
	}

	// Get current file info
	fileInfo, err := os.Stat(a.currentFile.Path)
	if err != nil {
		return false, &JSONLError{
			Message: "Failed to check file modification time",
			Err:     err,
		}
	}

	// Compare modification times
	currentModTime := fileInfo.ModTime()
	return currentModTime.After(a.currentFile.ModifiedAt), nil
}

// GetFileModificationInfo returns information about file modification status
func (a *App) GetFileModificationInfo() (map[string]interface{}, error) {
	if a.currentFile == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	result := map[string]interface{}{
		"filePath":        a.currentFile.Path,
		"loadedAt":        a.currentFile.LoadedAt,
		"originalModTime": a.currentFile.ModifiedAt,
		"isClipboard":     a.currentFile.Path == "<clipboard>",
	}

	// Skip modification check for clipboard content
	if a.currentFile.Path == "<clipboard>" {
		result["isModified"] = false
		result["currentModTime"] = nil
		return result, nil
	}

	// Get current file info
	fileInfo, err := os.Stat(a.currentFile.Path)
	if err != nil {
		result["isModified"] = false
		result["currentModTime"] = nil
		result["error"] = "File no longer exists or cannot be accessed"
		return result, nil
	}

	currentModTime := fileInfo.ModTime()
	result["currentModTime"] = currentModTime
	result["isModified"] = currentModTime.After(a.currentFile.ModifiedAt)

	return result, nil
}

// ReloadCurrentFile reloads the currently loaded file if it has been modified
func (a *App) ReloadCurrentFile() (*JSONLFile, error) {
	if a.currentFile == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	// Cannot reload clipboard content
	if a.currentFile.Path == "<clipboard>" {
		return nil, &JSONLError{
			Message: "Cannot reload clipboard content",
			Err:     errors.New("clipboard content cannot be reloaded"),
		}
	}

	// Check if file has been modified
	isModified, err := a.CheckFileModification()
	if err != nil {
		return nil, err
	}

	if !isModified {
		return a.currentFile, nil
	}

	// Reload the file
	return a.LoadJSONLFile(a.currentFile.Path)
}

// GetRecords returns a paginated subset of records with offset and limit parameters
func (a *App) GetRecords(offset, limit int) (*PaginatedRecords, error) {
	if a.currentFile == nil || a.cache == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	// Validate parameters
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = a.cache.pageSize // Use default page size
	}
	if limit > 1000 {
		limit = 1000 // Cap maximum limit for performance
	}

	totalRecords := a.cache.totalCount

	// Check if offset is beyond available records
	if offset >= totalRecords {
		return &PaginatedRecords{
			Records: []JSONRecord{},
			Offset:  offset,
			Limit:   limit,
			Total:   totalRecords,
			HasMore: false,
		}, nil
	}

	// Calculate end index
	endIndex := offset + limit
	if endIndex > totalRecords {
		endIndex = totalRecords
	}

	// Extract the requested slice of records
	records := a.cache.records[offset:endIndex]

	// Determine if there are more records available
	hasMore := endIndex < totalRecords

	return &PaginatedRecords{
		Records: records,
		Offset:  offset,
		Limit:   limit,
		Total:   totalRecords,
		HasMore: hasMore,
	}, nil
}

// GetRecordByLineNumber retrieves a specific record by its line number
func (a *App) GetRecordByLineNumber(lineNumber int) (*JSONRecord, error) {
	if a.currentFile == nil || a.cache == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	if lineNumber <= 0 {
		return nil, &JSONLError{
			Message:    "Line number must be greater than 0",
			LineNumber: lineNumber,
			Err:        ErrInvalidLineNum,
		}
	}

	// Search for the record with the specified line number
	for _, record := range a.cache.records {
		if record.LineNumber == lineNumber {
			return &record, nil
		}
	}

	return nil, &JSONLError{
		Message:    "Record not found at specified line number",
		LineNumber: lineNumber,
		Err:        ErrInvalidLineNum,
	}
}

// GetRecordRange returns records within a specific line number range
func (a *App) GetRecordRange(startLine, endLine int) ([]JSONRecord, error) {
	if a.currentFile == nil || a.cache == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	if startLine <= 0 || endLine <= 0 || startLine > endLine {
		return nil, &JSONLError{
			Message: "Invalid line number range",
			Err:     ErrInvalidLineNum,
		}
	}

	var result []JSONRecord
	for _, record := range a.cache.records {
		if record.LineNumber >= startLine && record.LineNumber <= endLine {
			result = append(result, record)
		}
	}

	return result, nil
}

// GetTotalRecordCount returns the total number of records in the current file
func (a *App) GetTotalRecordCount() (int, error) {
	if a.currentFile == nil || a.cache == nil {
		return 0, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	return a.cache.totalCount, nil
}

// SetPageSize updates the default page size for pagination
func (a *App) SetPageSize(pageSize int) error {
	if a.cache == nil {
		return &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}
	if pageSize > 1000 {
		pageSize = 1000 // Cap maximum page size
	}

	a.cache.pageSize = pageSize
	return nil
}

// GetPageSize returns the current page size setting
func (a *App) GetPageSize() (int, error) {
	if a.cache == nil {
		return 0, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	return a.cache.pageSize, nil
}

// TestFileLoading tests the file loading functionality with the sample file
func (a *App) TestFileLoading() (string, error) {
	// Test loading the sample file
	filePath := "test_sample.jsonl"

	// Load the file
	jsonlFile, err := a.LoadJSONLFile(filePath)
	if err != nil {
		return "", err
	}

	// Get file statistics
	stats, err := a.GetFileStats()
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("File loading test successful:\n")
	result += fmt.Sprintf("File Name: %s\n", jsonlFile.Name)
	result += fmt.Sprintf("File Path: %s\n", jsonlFile.Path)
	result += fmt.Sprintf("File Size: %d bytes\n", jsonlFile.Size)
	result += fmt.Sprintf("Record Count: %d\n", jsonlFile.Records)
	result += fmt.Sprintf("Loaded At: %s\n", jsonlFile.LoadedAt.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("\nDetailed Statistics:\n")
	result += fmt.Sprintf("Total Lines: %d\n", stats.TotalLines)
	result += fmt.Sprintf("Valid Records: %d\n", stats.ValidRecords)
	result += fmt.Sprintf("Invalid Lines: %v\n", stats.InvalidLines)
	result += fmt.Sprintf("Common Fields: %v\n", stats.CommonFields)
	result += fmt.Sprintf("File Size: %d bytes\n", stats.FileSize)

	result += fmt.Sprintf("\nLoaded Records:\n")
	for i, record := range a.records {
		result += fmt.Sprintf("Line %d: %s\n", record.LineNumber, record.RawJSON)
		if i >= 2 { // Show only first 3 records
			break
		}
	}

	return result, nil
}

// LoadJSONLFromClipboard loads JSONL data from the system clipboard
func (a *App) LoadJSONLFromClipboard() (*JSONLFile, error) {
	// Get clipboard content using Wails runtime
	clipboardContent, err := runtime.ClipboardGetText(a.ctx)
	if err != nil {
		return nil, &JSONLError{
			Message: "Failed to access clipboard",
			Err:     err,
		}
	}

	// Check if clipboard is empty
	if strings.TrimSpace(clipboardContent) == "" {
		return nil, &JSONLError{
			Message: "Clipboard is empty or contains no text",
			Err:     ErrClipboardEmpty,
		}
	}

	// Parse the clipboard content as JSONL
	records, stats, err := ParseJSONLFromString(clipboardContent)
	if err != nil {
		return nil, &JSONLError{
			Message: "Failed to parse clipboard content as JSONL",
			Err:     err,
		}
	}

	// Check if we found any valid records
	if stats.ValidRecords == 0 {
		return nil, &JSONLError{
			Message: "No valid JSONL records found in clipboard content",
			Err:     ErrInvalidJSONL,
		}
	}

	// Create JSONLFile metadata for clipboard content
	jsonlFile := &JSONLFile{
		Name:       "Clipboard Content",
		Path:       "<clipboard>",
		Size:       stats.FileSize,
		Records:    stats.ValidRecords,
		LoadedAt:   time.Now(),
		ModifiedAt: time.Now(), // For clipboard content, use current time
	}

	// Store in app state
	a.currentFile = jsonlFile
	a.records = records

	// Initialize cache for clipboard content
	a.cache = &RecordCache{
		records:    records,
		pageSize:   50, // Default page size for virtual scrolling
		totalCount: len(records),
	}

	return jsonlFile, nil
}

// TestJSONLParsing is a helper method to test JSONL parsing functionality
func (a *App) TestJSONLParsing(filePath string) (string, error) {
	parser, err := NewJSONLParser(filePath)
	if err != nil {
		return "", err
	}
	defer parser.Close()

	records, stats, err := parser.ParseJSONL()
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Parsed JSONL file successfully:\n")
	result += fmt.Sprintf("Total lines: %d\n", stats.TotalLines)
	result += fmt.Sprintf("Valid records: %d\n", stats.ValidRecords)
	result += fmt.Sprintf("Invalid lines: %v\n", stats.InvalidLines)
	result += fmt.Sprintf("Common fields: %v\n", stats.CommonFields)
	result += fmt.Sprintf("File size: %d bytes\n", stats.FileSize)
	result += fmt.Sprintf("\nFirst few records:\n")

	for i, record := range records {
		if i >= 3 { // Show only first 3 records
			break
		}
		result += fmt.Sprintf("Line %d: %s\n", record.LineNumber, record.RawJSON)
	}

	return result, nil
}

// TestPagination tests the record pagination functionality
func (a *App) TestPagination() (string, error) {
	// First load a test file
	filePath := "test_sample.jsonl"
	_, err := a.LoadJSONLFile(filePath)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Pagination test results:\n")

	// Test getting total count
	totalCount, err := a.GetTotalRecordCount()
	if err != nil {
		return "", err
	}
	result += fmt.Sprintf("Total records: %d\n", totalCount)

	// Test getting first page
	page1, err := a.GetRecords(0, 2)
	if err != nil {
		return "", err
	}
	result += fmt.Sprintf("\nFirst page (offset=0, limit=2):\n")
	result += fmt.Sprintf("  Records returned: %d\n", len(page1.Records))
	result += fmt.Sprintf("  Has more: %t\n", page1.HasMore)
	result += fmt.Sprintf("  Total: %d\n", page1.Total)
	for i, record := range page1.Records {
		result += fmt.Sprintf("  Record %d - Line %d: %s\n", i+1, record.LineNumber, record.RawJSON)
	}

	// Test getting second page
	page2, err := a.GetRecords(2, 2)
	if err != nil {
		return "", err
	}
	result += fmt.Sprintf("\nSecond page (offset=2, limit=2):\n")
	result += fmt.Sprintf("  Records returned: %d\n", len(page2.Records))
	result += fmt.Sprintf("  Has more: %t\n", page2.HasMore)
	for i, record := range page2.Records {
		result += fmt.Sprintf("  Record %d - Line %d: %s\n", i+1, record.LineNumber, record.RawJSON)
	}

	// Test getting record by line number
	if totalCount > 0 {
		firstRecord, err := a.GetRecordByLineNumber(1)
		if err != nil {
			result += fmt.Sprintf("\nError getting record by line number: %v\n", err)
		} else {
			result += fmt.Sprintf("\nRecord at line 1: %s\n", firstRecord.RawJSON)
		}
	}

	// Test page size functionality
	originalPageSize, _ := a.GetPageSize()
	result += fmt.Sprintf("\nOriginal page size: %d\n", originalPageSize)

	err = a.SetPageSize(25)
	if err != nil {
		result += fmt.Sprintf("Error setting page size: %v\n", err)
	} else {
		newPageSize, _ := a.GetPageSize()
		result += fmt.Sprintf("New page size: %d\n", newPageSize)
	}

	// Test edge cases
	result += fmt.Sprintf("\nEdge case tests:\n")

	// Test offset beyond records
	emptyPage, err := a.GetRecords(1000, 10)
	if err != nil {
		result += fmt.Sprintf("  Error with large offset: %v\n", err)
	} else {
		result += fmt.Sprintf("  Large offset result: %d records, HasMore: %t\n", len(emptyPage.Records), emptyPage.HasMore)
	}

	// Test negative offset (should be corrected to 0)
	correctedPage, err := a.GetRecords(-5, 1)
	if err != nil {
		result += fmt.Sprintf("  Error with negative offset: %v\n", err)
	} else {
		result += fmt.Sprintf("  Negative offset corrected: offset=%d, records=%d\n", correctedPage.Offset, len(correctedPage.Records))
	}

	return result, nil
}

// TestClipboardLoading tests the clipboard loading functionality
func (a *App) TestClipboardLoading() (string, error) {
	// Load JSONL data from clipboard
	jsonlFile, err := a.LoadJSONLFromClipboard()
	if err != nil {
		return "", err
	}

	// For clipboard content, we need to recalculate stats since there's no file
	clipboardContent := strings.Join(func() []string {
		var lines []string
		for _, record := range a.records {
			lines = append(lines, record.RawJSON)
		}
		return lines
	}(), "\n")

	_, stats, err := ParseJSONLFromString(clipboardContent)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Clipboard loading test successful:\n")
	result += fmt.Sprintf("Source: %s\n", jsonlFile.Name)
	result += fmt.Sprintf("Path: %s\n", jsonlFile.Path)
	result += fmt.Sprintf("Content Size: %d bytes\n", jsonlFile.Size)
	result += fmt.Sprintf("Record Count: %d\n", jsonlFile.Records)
	result += fmt.Sprintf("Loaded At: %s\n", jsonlFile.LoadedAt.Format("2006-01-02 15:04:05"))

	result += fmt.Sprintf("\nDetailed Statistics:\n")
	result += fmt.Sprintf("Total Lines: %d\n", stats.TotalLines)
	result += fmt.Sprintf("Valid Records: %d\n", stats.ValidRecords)
	result += fmt.Sprintf("Invalid Lines: %v\n", stats.InvalidLines)
	result += fmt.Sprintf("Common Fields: %v\n", stats.CommonFields)

	result += fmt.Sprintf("\nLoaded Records:\n")
	for i, record := range a.records {
		result += fmt.Sprintf("Line %d: %s\n", record.LineNumber, record.RawJSON)
		if i >= 2 { // Show only first 3 records
			break
		}
	}

	return result, nil
}

// SearchRecords searches through records with query filtering and returns paginated results
func (a *App) SearchRecords(options SearchOptions) (*SearchResult, error) {
	if a.currentFile == nil || a.cache == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	// Validate search options
	if strings.TrimSpace(options.Query) == "" {
		return &SearchResult{
			Records:      []JSONRecord{},
			Offset:       options.Offset,
			Limit:        options.Limit,
			Total:        0,
			TotalMatches: 0,
			HasMore:      false,
			Query:        options.Query,
		}, nil
	}

	// Normalize parameters
	if options.Offset < 0 {
		options.Offset = 0
	}
	if options.Limit <= 0 {
		options.Limit = 50 // Default limit
	}
	if options.Limit > 1000 {
		options.Limit = 1000 // Cap maximum limit
	}

	// Perform search
	var matchingRecords []JSONRecord

	if options.UseLucene {
		// Use Lucene syntax parsing
		luceneQuery := parseLuceneQuery(options.Query)

		if luceneQuery != nil {
			for _, record := range a.cache.records {
				if a.evaluateLuceneQuery(luceneQuery, record, options.CaseSensitive) {
					matchingRecords = append(matchingRecords, record)
				}
			}
		}
	} else {
		// Traditional search with optional field filtering
		query := options.Query
		if !options.CaseSensitive {
			query = strings.ToLower(query)
		}

		for _, record := range a.cache.records {
			var matches bool

			if options.SelectedField != "" && options.SelectedField != "all" {
				// Field-specific search
				if fieldValue, exists := record.Content[options.SelectedField]; exists {
					matches = a.matchFieldValue(fieldValue, options.Query, options.CaseSensitive)
				}
			} else {
				// Search all fields
				matches = a.recordMatches(record, query, options.CaseSensitive)
			}

			if matches {
				matchingRecords = append(matchingRecords, record)
			}
		}
	}

	totalMatches := len(matchingRecords)

	// Apply pagination to matching records
	startIndex := options.Offset
	if startIndex >= totalMatches {
		return &SearchResult{
			Records:      []JSONRecord{},
			Offset:       options.Offset,
			Limit:        options.Limit,
			Total:        a.cache.totalCount,
			TotalMatches: totalMatches,
			HasMore:      false,
			Query:        options.Query,
		}, nil
	}

	endIndex := startIndex + options.Limit
	if endIndex > totalMatches {
		endIndex = totalMatches
	}

	paginatedRecords := matchingRecords[startIndex:endIndex]
	hasMore := endIndex < totalMatches

	return &SearchResult{
		Records:      paginatedRecords,
		Offset:       options.Offset,
		Limit:        options.Limit,
		Total:        a.cache.totalCount,
		TotalMatches: totalMatches,
		HasMore:      hasMore,
		Query:        options.Query,
	}, nil
}

// recordMatches checks if a record matches the search query
func (a *App) recordMatches(record JSONRecord, query string, caseSensitive bool) bool {
	// Search in raw JSON string
	searchText := record.RawJSON
	if !caseSensitive {
		searchText = strings.ToLower(searchText)
	}

	if strings.Contains(searchText, query) {
		return true
	}

	// Also search in individual field values for more precise matching
	for _, value := range record.Content {
		valueStr := fmt.Sprintf("%v", value)
		if !caseSensitive {
			valueStr = strings.ToLower(valueStr)
		}
		if strings.Contains(valueStr, query) {
			return true
		}
	}

	return false
}

// parseLuceneQuery parses a Lucene query string into a structured query
func parseLuceneQuery(query string) *LuceneQuery {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	query = strings.TrimSpace(query)

	// Handle OR operator
	if strings.Contains(query, " OR ") {
		parts := strings.Split(query, " OR ")
		if len(parts) >= 2 {
			// For multiple OR conditions, create left-associative tree
			left := parseLuceneQuery(strings.TrimSpace(parts[0]))
			for i := 1; i < len(parts); i++ {
				right := parseLuceneQuery(strings.TrimSpace(parts[i]))
				left = &LuceneQuery{
					Type:  "or",
					Left:  left,
					Right: right,
				}
			}
			return left
		}
	}

	// Handle AND operator
	if strings.Contains(query, " AND ") {
		parts := strings.Split(query, " AND ")
		if len(parts) >= 2 {
			// For multiple AND conditions, create left-associative tree
			left := parseLuceneQuery(strings.TrimSpace(parts[0]))
			for i := 1; i < len(parts); i++ {
				right := parseLuceneQuery(strings.TrimSpace(parts[i]))
				left = &LuceneQuery{
					Type:  "and",
					Left:  left,
					Right: right,
				}
			}
			return left
		}
	}

	// Handle NOT operator
	if strings.HasPrefix(query, "NOT ") {
		return &LuceneQuery{
			Type:  "not",
			Query: parseLuceneQuery(strings.TrimSpace(query[4:])),
		}
	}

	// Handle field:value syntax
	if strings.Contains(query, ":") {
		parts := strings.SplitN(query, ":", 2)
		if len(parts) == 2 {
			field := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Handle quoted phrases
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) > 1 {
				return &LuceneQuery{
					Type:  "phrase",
					Field: field,
					Value: value[1 : len(value)-1],
				}
			}

			// Handle wildcards
			if strings.Contains(value, "*") || strings.Contains(value, "?") {
				return &LuceneQuery{
					Type:  "wildcard",
					Field: field,
					Value: value,
				}
			}

			return &LuceneQuery{
				Type:  "field",
				Field: field,
				Value: value,
			}
		}
	}

	// Handle quoted phrases
	if strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"") && len(query) > 1 {
		return &LuceneQuery{
			Type:  "phrase",
			Value: query[1 : len(query)-1],
		}
	}

	// Handle wildcards
	if strings.Contains(query, "*") || strings.Contains(query, "?") {
		return &LuceneQuery{
			Type:  "wildcard",
			Value: query,
		}
	}

	// Default term search
	return &LuceneQuery{
		Type:  "term",
		Value: query,
	}
}

// evaluateLuceneQuery evaluates a Lucene query against a record
func (a *App) evaluateLuceneQuery(query *LuceneQuery, record JSONRecord, caseSensitive bool) bool {
	if query == nil {
		return false
	}

	switch query.Type {
	case "and":
		return a.evaluateLuceneQuery(query.Left, record, caseSensitive) &&
			a.evaluateLuceneQuery(query.Right, record, caseSensitive)

	case "or":
		return a.evaluateLuceneQuery(query.Left, record, caseSensitive) ||
			a.evaluateLuceneQuery(query.Right, record, caseSensitive)

	case "not":
		return !a.evaluateLuceneQuery(query.Query, record, caseSensitive)

	case "field":
		if fieldValue, exists := record.Content[query.Field]; exists {
			return a.matchFieldValue(fieldValue, query.Value, caseSensitive)
		}
		return false

	case "phrase":
		if query.Field != "" {
			if fieldValue, exists := record.Content[query.Field]; exists {
				return a.matchPhrase(fmt.Sprintf("%v", fieldValue), query.Value, caseSensitive)
			}
			return false
		} else {
			return a.matchPhrase(record.RawJSON, query.Value, caseSensitive)
		}

	case "wildcard":
		if query.Field != "" {
			if fieldValue, exists := record.Content[query.Field]; exists {
				return a.matchWildcard(fmt.Sprintf("%v", fieldValue), query.Value, caseSensitive)
			}
			return false
		} else {
			return a.matchWildcard(record.RawJSON, query.Value, caseSensitive)
		}

	case "term":
		if query.Field != "" {
			if fieldValue, exists := record.Content[query.Field]; exists {
				return a.matchFieldValue(fieldValue, query.Value, caseSensitive)
			}
			return false
		} else {
			return a.matchTerm(record.RawJSON, query.Value, caseSensitive)
		}

	default:
		return false
	}
}

// matchFieldValue checks if a field value matches the search value
func (a *App) matchFieldValue(fieldValue interface{}, searchValue string, caseSensitive bool) bool {
	if fieldValue == nil {
		return false
	}

	fieldStr := fmt.Sprintf("%v", fieldValue)
	searchStr := searchValue
	targetStr := fieldStr

	if !caseSensitive {
		searchStr = strings.ToLower(searchStr)
		targetStr = strings.ToLower(targetStr)
	}

	return strings.Contains(targetStr, searchStr)
}

// matchPhrase checks if text contains the exact phrase
func (a *App) matchPhrase(text, phrase string, caseSensitive bool) bool {
	if text == "" {
		return false
	}

	searchStr := phrase
	targetStr := text

	if !caseSensitive {
		searchStr = strings.ToLower(searchStr)
		targetStr = strings.ToLower(targetStr)
	}

	return strings.Contains(targetStr, searchStr)
}

// matchWildcard checks if text matches a wildcard pattern
func (a *App) matchWildcard(text, pattern string, caseSensitive bool) bool {
	if text == "" {
		return false
	}

	searchStr := pattern
	targetStr := text

	if !caseSensitive {
		searchStr = strings.ToLower(searchStr)
		targetStr = strings.ToLower(targetStr)
	}

	// Convert wildcard pattern to regex
	regexPattern := strings.ReplaceAll(searchStr, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
	regexPattern = "^" + regexPattern + "$"

	// Use simple pattern matching instead of regex for better performance
	return a.simpleWildcardMatch(targetStr, searchStr)
}

// simpleWildcardMatch performs simple wildcard matching
func (a *App) simpleWildcardMatch(text, pattern string) bool {
	// Simple implementation for * and ? wildcards
	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		return strings.Contains(text, pattern)
	}

	// For complex patterns, use basic matching
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(text, middle)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(text, suffix)
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(text, prefix)
	}

	return strings.Contains(text, strings.ReplaceAll(pattern, "*", ""))
}

// matchTerm checks if text contains the search term
func (a *App) matchTerm(text, term string, caseSensitive bool) bool {
	if text == "" {
		return false
	}

	searchStr := term
	targetStr := text

	if !caseSensitive {
		searchStr = strings.ToLower(searchStr)
		targetStr = strings.ToLower(targetStr)
	}

	return strings.Contains(targetStr, searchStr)
}

// GetSearchHighlights returns highlighting information for search matches in a record
func (a *App) GetSearchHighlights(record JSONRecord, query string, caseSensitive bool) ([]HighlightMatch, error) {
	if strings.TrimSpace(query) == "" {
		return []HighlightMatch{}, nil
	}

	var highlights []HighlightMatch
	searchQuery := query
	if !caseSensitive {
		searchQuery = strings.ToLower(searchQuery)
	}

	// Find matches in raw JSON
	searchText := record.RawJSON
	if !caseSensitive {
		searchText = strings.ToLower(searchText)
	}

	// Find all occurrences of the query in the raw JSON
	startPos := 0
	for {
		index := strings.Index(searchText[startPos:], searchQuery)
		if index == -1 {
			break
		}

		actualPos := startPos + index
		highlights = append(highlights, HighlightMatch{
			Text:      record.RawJSON[actualPos : actualPos+len(query)],
			StartPos:  actualPos,
			EndPos:    actualPos + len(query),
			FieldName: "raw",
		})

		startPos = actualPos + len(searchQuery)
	}

	// Find matches in individual fields
	for fieldName, value := range record.Content {
		valueStr := fmt.Sprintf("%v", value)
		searchValueStr := valueStr
		if !caseSensitive {
			searchValueStr = strings.ToLower(searchValueStr)
		}

		if strings.Contains(searchValueStr, searchQuery) {
			// Find the position in the original raw JSON
			fieldStartPos := strings.Index(record.RawJSON, fmt.Sprintf("\"%s\"", fieldName))
			if fieldStartPos != -1 {
				highlights = append(highlights, HighlightMatch{
					Text:      valueStr,
					StartPos:  fieldStartPos,
					EndPos:    fieldStartPos + len(valueStr),
					FieldName: fieldName,
				})
			}
		}
	}

	return highlights, nil
}

// SearchRecordsWithHighlights searches records and includes highlighting information
func (a *App) SearchRecordsWithHighlights(options SearchOptions) (*SearchResult, [][]HighlightMatch, error) {
	searchResult, err := a.SearchRecords(options)
	if err != nil {
		return nil, nil, err
	}

	// Generate highlights for each record in the result
	var allHighlights [][]HighlightMatch
	for _, record := range searchResult.Records {
		highlights, err := a.GetSearchHighlights(record, options.Query, options.CaseSensitive)
		if err != nil {
			// If highlighting fails, continue with empty highlights for this record
			highlights = []HighlightMatch{}
		}
		allHighlights = append(allHighlights, highlights)
	}

	return searchResult, allHighlights, nil
}

// GetCommonFields analyzes and returns common field names across all records
func (a *App) GetCommonFields() ([]string, error) {
	if a.currentFile == nil || a.cache == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	fieldCounts := make(map[string]int)
	totalRecords := len(a.cache.records)

	// Count occurrences of each field
	for _, record := range a.cache.records {
		for fieldName := range record.Content {
			fieldCounts[fieldName]++
		}
	}

	// Calculate common fields (fields that appear in at least 50% of records)
	var commonFields []string
	threshold := totalRecords / 2
	for field, count := range fieldCounts {
		if count >= threshold {
			commonFields = append(commonFields, field)
		}
	}

	return commonFields, nil
}

// GetAllFields returns all unique field names found across all records
func (a *App) GetAllFields() ([]string, error) {
	if a.currentFile == nil || a.cache == nil {
		return nil, &JSONLError{
			Message: "No file currently loaded",
			Err:     ErrNoFileLoaded,
		}
	}

	// Collect all unique field names
	fieldSet := make(map[string]bool)

	for _, record := range a.cache.records {
		for field := range record.Content {
			fieldSet[field] = true
		}
	}

	// Convert to sorted slice
	var allFields []string
	for field := range fieldSet {
		allFields = append(allFields, field)
	}

	// Sort alphabetically for consistent ordering
	for i := 0; i < len(allFields)-1; i++ {
		for j := i + 1; j < len(allFields); j++ {
			if allFields[i] > allFields[j] {
				allFields[i], allFields[j] = allFields[j], allFields[i]
			}
		}
	}

	return allFields, nil
}

// TestSearchFunctionality tests the search functionality with various scenarios
func (a *App) TestSearchFunctionality() (string, error) {
	// First load a test file
	filePath := "test_sample.jsonl"
	_, err := a.LoadJSONLFile(filePath)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Search functionality test results:\n")

	// Test 1: Basic search
	searchOptions := SearchOptions{
		Query:         "test",
		CaseSensitive: false,
		Offset:        0,
		Limit:         10,
	}

	searchResult, err := a.SearchRecords(searchOptions)
	if err != nil {
		return "", err
	}

	result += fmt.Sprintf("\nTest 1 - Basic search for 'test' (case-insensitive):\n")
	result += fmt.Sprintf("  Total matches: %d\n", searchResult.TotalMatches)
	result += fmt.Sprintf("  Records returned: %d\n", len(searchResult.Records))
	result += fmt.Sprintf("  Has more: %t\n", searchResult.HasMore)

	for i, record := range searchResult.Records {
		result += fmt.Sprintf("  Match %d - Line %d: %s\n", i+1, record.LineNumber, record.RawJSON)
	}

	// Test 2: Case-sensitive search
	searchOptions.CaseSensitive = true
	searchOptions.Query = "Test"

	searchResult2, err := a.SearchRecords(searchOptions)
	if err != nil {
		return "", err
	}

	result += fmt.Sprintf("\nTest 2 - Case-sensitive search for 'Test':\n")
	result += fmt.Sprintf("  Total matches: %d\n", searchResult2.TotalMatches)
	result += fmt.Sprintf("  Records returned: %d\n", len(searchResult2.Records))

	// Test 3: Search with highlighting
	searchOptions.Query = "name"
	searchOptions.CaseSensitive = false

	searchResult3, highlights, err := a.SearchRecordsWithHighlights(searchOptions)
	if err != nil {
		return "", err
	}

	result += fmt.Sprintf("\nTest 3 - Search with highlighting for 'name':\n")
	result += fmt.Sprintf("  Total matches: %d\n", searchResult3.TotalMatches)
	result += fmt.Sprintf("  Records with highlights: %d\n", len(highlights))

	for i, record := range searchResult3.Records {
		result += fmt.Sprintf("  Record %d - Line %d: %s\n", i+1, record.LineNumber, record.RawJSON)
		if i < len(highlights) {
			result += fmt.Sprintf("    Highlights: %d matches\n", len(highlights[i]))
			for j, highlight := range highlights[i] {
				result += fmt.Sprintf("      %d. Field: %s, Text: '%s', Pos: %d-%d\n",
					j+1, highlight.FieldName, highlight.Text, highlight.StartPos, highlight.EndPos)
			}
		}
	}

	// Test 4: Empty search query
	searchOptions.Query = ""
	searchResult4, err := a.SearchRecords(searchOptions)
	if err != nil {
		return "", err
	}

	result += fmt.Sprintf("\nTest 4 - Empty search query:\n")
	result += fmt.Sprintf("  Total matches: %d\n", searchResult4.TotalMatches)
	result += fmt.Sprintf("  Records returned: %d\n", len(searchResult4.Records))

	// Test 5: Search with no matches
	searchOptions.Query = "nonexistentstring12345"
	searchResult5, err := a.SearchRecords(searchOptions)
	if err != nil {
		return "", err
	}

	result += fmt.Sprintf("\nTest 5 - Search with no matches:\n")
	result += fmt.Sprintf("  Total matches: %d\n", searchResult5.TotalMatches)
	result += fmt.Sprintf("  Records returned: %d\n", len(searchResult5.Records))

	// Test 6: Get common fields
	commonFields, err := a.GetCommonFields()
	if err != nil {
		return "", err
	}

	result += fmt.Sprintf("\nTest 6 - Common fields analysis:\n")
	result += fmt.Sprintf("  Common fields: %v\n", commonFields)

	return result, nil
}

// TestFileModificationDetection tests the file modification detection functionality
func (a *App) TestFileModificationDetection() (string, error) {
	// First load a test file
	filePath := "test_sample.jsonl"
	jsonlFile, err := a.LoadJSONLFile(filePath)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("File modification detection test results:\n")
	result += fmt.Sprintf("File loaded: %s\n", jsonlFile.Name)
	result += fmt.Sprintf("Loaded at: %s\n", jsonlFile.LoadedAt.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Original mod time: %s\n", jsonlFile.ModifiedAt.Format("2006-01-02 15:04:05"))

	// Test 1: Check modification status immediately after loading
	isModified, err := a.CheckFileModification()
	if err != nil {
		return "", err
	}
	result += fmt.Sprintf("\nTest 1 - Immediate modification check: %t\n", isModified)

	// Test 2: Get detailed modification info
	modInfo, err := a.GetFileModificationInfo()
	if err != nil {
		return "", err
	}
	result += fmt.Sprintf("\nTest 2 - Detailed modification info:\n")
	result += fmt.Sprintf("  File path: %s\n", modInfo["filePath"])
	result += fmt.Sprintf("  Is clipboard: %t\n", modInfo["isClipboard"])
	result += fmt.Sprintf("  Is modified: %t\n", modInfo["isModified"])
	if modInfo["currentModTime"] != nil {
		currentModTime := modInfo["currentModTime"].(time.Time)
		result += fmt.Sprintf("  Current mod time: %s\n", currentModTime.Format("2006-01-02 15:04:05"))
	}

	// Test 3: Test reload functionality
	reloadedFile, err := a.ReloadCurrentFile()
	if err != nil {
		return "", err
	}
	result += fmt.Sprintf("\nTest 3 - Reload file:\n")
	result += fmt.Sprintf("  File name: %s\n", reloadedFile.Name)
	result += fmt.Sprintf("  Records: %d\n", reloadedFile.Records)
	result += fmt.Sprintf("  Reloaded at: %s\n", reloadedFile.LoadedAt.Format("2006-01-02 15:04:05"))

	// Test 4: Test with simulated clipboard content (without actual clipboard access)
	result += fmt.Sprintf("\nTest 4 - Simulated clipboard test:\n")
	result += fmt.Sprintf("  Testing clipboard modification detection logic\n")

	// Simulate clipboard content by creating a temporary JSONLFile
	originalFile := a.currentFile
	a.currentFile = &JSONLFile{
		Name:       "Clipboard Content",
		Path:       "<clipboard>",
		Size:       100,
		Records:    3,
		LoadedAt:   time.Now(),
		ModifiedAt: time.Now(),
	}

	clipboardModified, err := a.CheckFileModification()
	if err != nil {
		result += fmt.Sprintf("  Error checking clipboard modification: %v\n", err)
	} else {
		result += fmt.Sprintf("  Clipboard is modified: %t (should be false)\n", clipboardModified)
	}

	clipboardModInfo, err := a.GetFileModificationInfo()
	if err != nil {
		result += fmt.Sprintf("  Error getting clipboard mod info: %v\n", err)
	} else {
		result += fmt.Sprintf("  Is clipboard content: %t\n", clipboardModInfo["isClipboard"])
	}

	// Restore original file
	a.currentFile = originalFile

	return result, nil
}

// ExportSearchResults exports all search results to a JSONL file
func (a *App) ExportSearchResults(searchQuery string, shownFields []string, hiddenFields []string) (string, error) {
	// Get user's downloads directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	downloadsDir := filepath.Join(homeDir, "Downloads")

	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	filename := fmt.Sprintf("jsonl-viewer-export-%s.jsonl", timestamp)
	filepath := filepath.Join(downloadsDir, filename)

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	// Debug logging
	fmt.Printf("Export: searchQuery='%s', shownFields=%v, hiddenFields=%v\n", searchQuery, shownFields, hiddenFields)

	// Get all records (not just current page)
	allRecords, err := a.GetAllRecords(searchQuery)
	if err != nil {
		return "", fmt.Errorf("failed to get all records: %w", err)
	}

	fmt.Printf("Export: Found %d records to export\n", len(allRecords))

	// Process each record and write to file
	exportedCount := 0
	for _, record := range allRecords {
		// Apply field visibility filtering
		displayJSON := a.getDisplayJSON(record, shownFields, hiddenFields)
		_, err := file.WriteString(displayJSON + "\n")
		if err != nil {
			return "", fmt.Errorf("failed to write to export file: %w", err)
		}
		exportedCount++
	}

	fmt.Printf("Export: Successfully exported %d records to %s\n", exportedCount, filepath)
	return filepath, nil
}

// GetAllRecords gets all records that match the search query
func (a *App) GetAllRecords(searchQuery string) ([]JSONRecord, error) {
	if a.currentFile == nil {
		return nil, fmt.Errorf("no file loaded")
	}

	fmt.Printf("GetAllRecords: Reading file %s with searchQuery='%s'\n", a.currentFile.Path, searchQuery)

	// Read all records from file
	var allRecords []JSONRecord
	lineNumber := 1
	totalLines := 0
	validLines := 0

	file, err := os.Open(a.currentFile.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		totalLines++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			lineNumber++
			continue
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
			lineNumber++
			continue
		}

		validLines++
		record := JSONRecord{
			LineNumber: lineNumber,
			Content:    jsonData,
			RawJSON:    line,
		}

		// If there's a search query, check if record matches using Lucene syntax
		if searchQuery != "" {
			// Parse Lucene query
			luceneQuery := parseLuceneQuery(searchQuery)
			if luceneQuery != nil {
				if !a.evaluateLuceneQuery(luceneQuery, record, false) {
					lineNumber++
					continue
				}
			} else {
				// Fallback to simple search if Lucene parsing fails
				if !a.recordMatches(record, searchQuery, false) {
					lineNumber++
					continue
				}
			}
		}

		allRecords = append(allRecords, record)
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	fmt.Printf("GetAllRecords: Total lines=%d, valid lines=%d, matched lines=%d\n", totalLines, validLines, len(allRecords))
	return allRecords, nil
}

// getDisplayJSON applies field visibility filtering to a record
func (a *App) getDisplayJSON(record JSONRecord, shownFields []string, hiddenFields []string) string {
	// If no field visibility is set, return the original JSON
	if len(shownFields) == 0 && len(hiddenFields) == 0 {
		return record.RawJSON
	}

	// Create a filtered copy of the content
	filteredContent := make(map[string]interface{})

	if len(shownFields) > 0 {
		// Show only specified fields
		for _, field := range shownFields {
			if value, exists := record.Content[field]; exists {
				filteredContent[field] = value
			}
		}
	} else {
		// Hide specified fields
		for field, value := range record.Content {
			shouldHide := false
			for _, hiddenField := range hiddenFields {
				if field == hiddenField {
					shouldHide = true
					break
				}
			}
			if !shouldHide {
				filteredContent[field] = value
			}
		}
	}

	// Convert back to JSON
	jsonBytes, err := json.Marshal(filteredContent)
	if err != nil {
		fmt.Printf("getDisplayJSON: Error marshaling filtered content: %v\n", err)
		return record.RawJSON // Fallback to original if marshaling fails
	}

	result := string(jsonBytes)

	return result
}
