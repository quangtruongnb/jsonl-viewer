// Svelte stores for global state management
import { writable, derived } from 'svelte/store';
import type { JSONLFile, JSONRecord, AppState } from './types';

// Main application state store
export const appState = writable<AppState>({
  currentFile: null,
  records: [],
  searchQuery: '',
  currentPage: 0,
  pageSize: 50,
  isLoading: false,
  error: null
});

// Individual stores for easier component access
export const currentFile = writable<JSONLFile | null>(null);
export const records = writable<JSONRecord[]>([]);
export const searchQuery = writable<string>('');
export const currentPage = writable<number>(0);
export const pageSize = writable<number>(50);
export const isLoading = writable<boolean>(false);
export const error = writable<string | null>(null);

// Navigation state
export const currentRecordIndex = writable<number>(0);
export const totalRecords = writable<number>(0);

// Field visibility state
export const availableFields = writable<string[]>([]);
export const fieldsToShow = writable<string[]>([]);
export const fieldsToHide = writable<string[]>([]);

// Derived stores for computed values
export const hasFile = derived(
  currentFile,
  ($currentFile) => $currentFile !== null
);

export const hasRecords = derived(
  records,
  ($records) => $records.length > 0
);

export const isSearching = derived(
  searchQuery,
  ($searchQuery) => $searchQuery.trim().length > 0
);

export const totalPages = derived(
  [records, pageSize],
  ([$records, $pageSize]) => Math.ceil($records.length / $pageSize)
);

// Navigation derived stores
export const canNavigatePrevious = derived(
  currentRecordIndex,
  ($currentRecordIndex) => $currentRecordIndex > 0
);

export const canNavigateNext = derived(
  [currentRecordIndex, totalRecords],
  ([$currentRecordIndex, $totalRecords]) => $currentRecordIndex < $totalRecords - 1
);

// Actions for updating state
export const actions = {
  setFile: (file: JSONLFile | null) => {
    currentFile.set(file);
    if (!file) {
      records.set([]);
      searchQuery.set('');
      currentPage.set(0);
      currentRecordIndex.set(0);
      totalRecords.set(0);
      error.set(null);
      // Reset field visibility when new file is loaded
      availableFields.set([]);
      fieldsToShow.set([]);
      fieldsToHide.set([]);
    }
  },
  
  setRecords: (newRecords: JSONRecord[]) => {
    records.set(newRecords);
  },
  
  setSearchQuery: (query: string) => {
    searchQuery.set(query);
    currentPage.set(0); // Reset to first page when searching
  },
  
  setCurrentPage: (page: number) => {
    currentPage.set(page);
  },
  
  setLoading: (loading: boolean) => {
    isLoading.set(loading);
  },
  
  setError: (errorMessage: string | null) => {
    error.set(errorMessage);
  },
  
  clearError: () => {
    error.set(null);
  },

  // Navigation actions
  setCurrentRecordIndex: (index: number) => {
    currentRecordIndex.set(index);
  },

  setTotalRecords: (total: number) => {
    totalRecords.set(total);
  },

  navigatePrevious: () => {
    currentRecordIndex.update(index => Math.max(0, index - 1));
  },

  navigateNext: () => {
    currentRecordIndex.update(index => {
      let total = 0;
      const unsubscribe = totalRecords.subscribe(t => total = t);
      unsubscribe();
      return Math.min(total - 1, index + 1);
    });
  },

  goToRecord: (index: number) => {
    let total = 0;
    const unsubscribe = totalRecords.subscribe(t => total = t);
    unsubscribe();
    const clampedIndex = Math.max(0, Math.min(total - 1, index));
    currentRecordIndex.set(clampedIndex);
  },

  // Field visibility actions
  setAvailableFields: (fields: string[]) => {
    availableFields.set(fields);
  },

  setFieldsToShow: (fields: string[]) => {
    console.log('setFieldsToShow called with:', fields);
    fieldsToShow.set(fields);
  },

  setFieldsToHide: (fields: string[]) => {
    console.log('setFieldsToHide called with:', fields);
    fieldsToHide.set(fields);
  },

  addFieldToShow: (field: string) => {
    console.log('addFieldToShow called with:', field);
    fieldsToShow.update(fields => {
      if (!fields.includes(field)) {
        const newFields = [...fields, field];
        console.log('fieldsToShow updated to:', newFields);
        return newFields;
      }
      return fields;
    });
  },

  removeFieldToShow: (field: string) => {
    console.log('removeFieldToShow called with:', field);
    fieldsToShow.update(fields => {
      const newFields = fields.filter(f => f !== field);
      console.log('fieldsToShow updated to:', newFields);
      return newFields;
    });
  },

  addFieldToHide: (field: string) => {
    console.log('addFieldToHide called with:', field);
    fieldsToHide.update(fields => {
      if (!fields.includes(field)) {
        const newFields = [...fields, field];
        console.log('fieldsToHide updated to:', newFields);
        return newFields;
      }
      return fields;
    });
  },

  removeFieldToHide: (field: string) => {
    console.log('removeFieldToHide called with:', field);
    fieldsToHide.update(fields => {
      const newFields = fields.filter(f => f !== field);
      console.log('fieldsToHide updated to:', newFields);
      return newFields;
    });
  }
};