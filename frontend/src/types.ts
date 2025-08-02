// TypeScript interfaces for JSONL Viewer data models

export interface JSONLFile {
  name: string;
  path: string;
  size: number;
  records: number;
  loadedAt: string;
}

export interface JSONRecord {
  lineNumber: number;
  content: Record<string, any>;
  rawJSON: string;
}

export interface SearchResult {
  records: JSONRecord[];
  totalMatches: number;
  hasMore: boolean;
}

export interface AppState {
  currentFile: JSONLFile | null;
  records: JSONRecord[];
  searchQuery: string;
  currentPage: number;
  pageSize: number;
  isLoading: boolean;
  error: string | null;
}

export interface SearchOptions {
  query: string;
  caseSensitive: boolean;
  useLucene: boolean;
  selectedField: string;
  offset: number;
  limit: number;
}

export interface FileStats {
  totalLines: number;
  validRecords: number;
  invalidLines: number[];
  commonFields: string[];
  fileSize: number;
}