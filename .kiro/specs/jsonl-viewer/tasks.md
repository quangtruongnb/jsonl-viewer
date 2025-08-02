# Implementation Plan

- [x] 1. Set up backend data structures and core interfaces
  - Create JSONLFile, JSONRecord, and related structs in app.go
  - Define error types for JSONL parsing and file operations
  - _Requirements: 1.4, 1.5_

- [x] 2. Implement JSONL file parsing functionality
  - Create JSONL parser that reads files line by line
  - Add validation for JSON format on each line
  - Implement error handling with specific line number reporting
  - _Requirements: 1.2, 1.4_

- [x] 3. Add file dialog and file loading backend methods
  - Implement OpenFile method using Wails runtime for native file dialog
  - Create LoadJSONLFile method to parse selected files
  - Add file statistics calculation (size, record count)
  - _Requirements: 1.1, 1.2, 1.5, 5.1, 5.3_

- [x] 4. Implement clipboard functionality for JSONL data
  - Add LoadJSONLFromClipboard method using Wails runtime clipboard access
  - Parse clipboard content as JSONL format
  - Handle clipboard-specific error cases
  - _Requirements: 1.3, 1.6_

- [x] 5. Create record pagination and retrieval system
  - Implement GetRecords method with offset and limit parameters
  - Add virtual scrolling support for large files
  - Create efficient record caching mechanism
  - _Requirements: 2.3, 4.1, 4.2_

- [x] 6. Build search functionality in backend
  - Create SearchRecords method with query filtering
  - Implement text highlighting for search matches
  - Add case-sensitive and case-insensitive search options
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 7. Add file statistics and analysis features
  - Implement GetFileStats method for file information
  - Create GetCommonFields method to analyze JSON structure
  - Add file modification detection capabilities
  - _Requirements: 5.1, 5.2, 5.4_

- [x] 8. Create main frontend application structure
  - Set up Svelte stores for global state management
  - Define TypeScript interfaces for data models
  - Create main App.svelte component structure
  - _Requirements: 2.1, 2.4_

- [x] 9. Build FileLoader component for file and clipboard input
  - Create file selection button with native dialog integration
  - Add clipboard paste button functionality
  - Implement loading states and error display
  - _Requirements: 1.1, 1.3, 1.4_

- [x] 10. Implement RecordViewer component with virtual scrolling
  - Create virtual scrolling container for performance
  - Add JSON syntax highlighting for record display
  - Implement line number display from original file
  - _Requirements: 2.1, 2.2, 2.4, 4.1, 4.2_

- [x] 11. Create SearchComponent for filtering records
  - Build search input with debounced queries
  - Implement search result highlighting
  - Add "No results found" state handling
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 12. Build StatisticsPanel for file information display
  - Create file information display (name, size, record count)
  - Add common fields analysis visualization
  - Implement file path display for reference
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 13. Add navigation controls and "Go to line" functionality
  - Implement previous/next record navigation
  - Create "Go to line" input and jump functionality
  - Add navigation button state management (disable at boundaries)
  - _Requirements: 4.3, 4.4_

- [ ] 14. Integrate all components and implement error handling
  - Connect all frontend components with backend methods
  - Implement global error state management
  - Add user-friendly error messages and retry mechanisms
  - _Requirements: 1.4, 3.3_

- [ ] 15. Add performance optimizations and final polish
  - Implement efficient DOM manipulation for virtual scrolling
  - Add search result caching for better performance
  - Optimize memory usage for large files
  - _Requirements: 4.1, 4.2_

- [x] 16. Implement field visibility controls
  - Remove the search field select box from SearchComponent
  - Create FieldVisibilityControls component with multi-select for showing fields
  - Add multi-select for hiding fields in the same component
  - Integrate field visibility state management in stores
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 17. Update RecordViewer to respect field visibility settings
  - Modify record display logic to filter fields based on visibility settings
  - Implement immediate updates when field visibility changes
  - Add field visibility reset when new files are loaded
  - _Requirements: 6.6, 6.7_
