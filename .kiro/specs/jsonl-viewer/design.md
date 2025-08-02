# Design Document

## Overview

The JSONL Viewer will be built as a Wails desktop application using Go for the backend and Svelte with TypeScript for the frontend. The application will provide a clean, efficient interface for loading, viewing, searching, and analyzing JSONL (JSON Lines) files. The design emphasizes performance for large files through virtual scrolling and on-demand loading.

## Architecture

The application follows the Wails architecture pattern with clear separation between the Go backend and Svelte frontend:

```
┌─────────────────────────────────────────┐
│              Frontend (Svelte)          │
│  ┌─────────────┐  ┌─────────────────────┐│
│  │ File Loader │  │   Record Viewer     ││
│  │ Component   │  │   Component         ││
│  └─────────────┘  └─────────────────────┘│
│  ┌─────────────┐  ┌─────────────────────┐│
│  │   Search    │  │   Statistics        ││
│  │ Component   │  │   Component         ││
│  └─────────────┘  └─────────────────────┘│
└─────────────────────────────────────────┘
                    │
                    │ Wails Bridge
                    │
┌─────────────────────────────────────────┐
│              Backend (Go)               │
│  ┌─────────────┐  ┌─────────────────────┐│
│  │ File Service│  │   JSONL Parser      ││
│  │             │  │   Service           ││
│  └─────────────┘  └─────────────────────┘│
│  ┌─────────────┐  ┌─────────────────────┐│
│  │ Clipboard   │  │   Search Service    ││
│  │ Service     │  │                     ││
│  └─────────────┘  └─────────────────────┘│
└─────────────────────────────────────────┘
```

## Components and Interfaces

### Backend Components (Go)

#### 1. App Struct Enhancement
```go
type App struct {
    ctx context.Context
    currentFile *JSONLFile
    records []JSONRecord
}

type JSONLFile struct {
    Name     string
    Path     string
    Size     int64
    Records  int
    LoadedAt time.Time
}

type JSONRecord struct {
    LineNumber int
    Content    map[string]interface{}
    RawJSON    string
}
```

#### 2. File Service Methods
- `OpenFile() (string, error)` - Opens native file dialog and returns selected file path
- `LoadJSONLFile(filePath string) (*JSONLFile, error)` - Loads and parses JSONL file
- `LoadJSONLFromClipboard() (*JSONLFile, error)` - Loads JSONL data from clipboard
- `GetRecords(offset, limit int) ([]JSONRecord, error)` - Returns paginated records
- `GetFileStats() (*JSONLFile, error)` - Returns current file statistics

#### 3. Search Service Methods
- `SearchRecords(query string, offset, limit int) ([]JSONRecord, error)` - Searches through records
- `GetCommonFields() ([]string, error)` - Analyzes and returns common field names

### Frontend Components (Svelte)

#### 1. Main App Component
- Orchestrates all child components
- Manages global application state
- Handles file loading coordination

#### 2. FileLoader Component
- File selection button with native dialog
- Clipboard paste button
- Drag and drop support for files
- Loading states and error handling

#### 3. RecordViewer Component
- Virtual scrolling for performance
- JSON syntax highlighting
- Line number display
- Navigation controls (previous/next, go to line)

#### 4. SearchComponent
- Search input with debounced queries
- Search result highlighting
- Filter controls and options

#### 5. StatisticsPanel Component
- File information display
- Record count and statistics
- Common fields analysis
- File modification detection

## Data Models

### Frontend TypeScript Interfaces

```typescript
interface JSONLFile {
  name: string;
  path: string;
  size: number;
  records: number;
  loadedAt: string;
}

interface JSONRecord {
  lineNumber: number;
  content: Record<string, any>;
  rawJSON: string;
}

interface SearchResult {
  records: JSONRecord[];
  totalMatches: number;
  hasMore: boolean;
}

interface AppState {
  currentFile: JSONLFile | null;
  records: JSONRecord[];
  searchQuery: string;
  currentPage: number;
  pageSize: number;
  isLoading: boolean;
  error: string | null;
}
```

### Backend Go Structs

```go
type JSONLParser struct {
    file *os.File
    scanner *bufio.Scanner
    lineCount int
}

type SearchOptions struct {
    Query      string
    CaseSensitive bool
    Offset     int
    Limit      int
}

type FileStats struct {
    TotalLines    int
    ValidRecords  int
    InvalidLines  []int
    CommonFields  []string
    FileSize      int64
}
```

## Error Handling

### Backend Error Types
- `ErrInvalidJSONL` - Invalid JSONL format with line number
- `ErrFileNotFound` - File doesn't exist or can't be accessed
- `ErrClipboardEmpty` - No data in clipboard
- `ErrParsingFailed` - JSON parsing failed with details

### Frontend Error Handling
- Global error state management
- User-friendly error messages
- Retry mechanisms for transient failures
- Validation feedback for user inputs

## Testing Strategy

### Backend Testing (Go)
- Unit tests for JSONL parsing logic
- File I/O operation tests with mock files
- Search functionality tests with various query types
- Clipboard integration tests
- Performance tests with large files

### Frontend Testing (Svelte)
- Component unit tests using Jest/Vitest
- Integration tests for component interactions
- Virtual scrolling performance tests
- Search functionality tests
- User interaction tests

### End-to-End Testing
- File loading workflows
- Search and navigation scenarios
- Error handling paths
- Performance with large datasets

## Performance Considerations

### Virtual Scrolling Implementation
- Render only visible records plus buffer
- Lazy loading of record content
- Efficient DOM manipulation
- Memory management for large files

### Backend Optimization
- Streaming file parsing to avoid loading entire file into memory
- Indexed search for better performance
- Caching of parsed records
- Background processing for large files

### Search Performance
- Debounced search queries (300ms delay)
- Incremental search results
- Search result caching
- Efficient string matching algorithms

## Security Considerations

- File path validation to prevent directory traversal
- JSON parsing limits to prevent DoS attacks
- Clipboard access permissions
- Memory usage limits for large files
- Input sanitization for search queries