<script lang="ts">
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { SearchRecords } from '../../wailsjs/go/main/App.js';
  import { searchQuery, records, actions, hasFile, isLoading } from '../stores';
  import type { SearchOptions, SearchResult, JSONRecord } from '../types';

  const dispatch = createEventDispatcher<{
    searchResults: SearchResult;
    searchError: string;
    searchCleared: void;
  }>();

  // Component state
  let searchInput = '';
  let caseSensitive = false;
  let isSearching = false;
  let searchError: string | null = null;
  let searchResults: SearchResult | null = null;
  let debounceTimer: number | null = null;

  let useLuceneSyntax = true;

  // Constants
  const DEBOUNCE_DELAY = 1000; // 300ms debounce delay
  const SEARCH_LIMIT = 10000; // Increased search limit to show more results



  // Subscribe to search query store to sync with external changes
  const unsubscribeSearchQuery = searchQuery.subscribe((query) => {
    if (query !== searchInput) {
      searchInput = query;
    }
  });

  // Cleanup on component destroy
  onDestroy(() => {
    unsubscribeSearchQuery();
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }
  });

  // Debounced search function
  function debounceSearch() {
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    debounceTimer = window.setTimeout(() => {
      performSearch();
    }, DEBOUNCE_DELAY);
  }

  // Perform the actual search
  async function performSearch() {
    // In Lucene mode, preserve whitespace; in regular mode, trim
    const query = useLuceneSyntax ? searchInput : searchInput.trim();
    
    // Clear previous results and errors
    searchError = null;
    
    // If query is empty, clear search
    if (!query) {
      clearSearch();
      return;
    }

    try {
      isSearching = true;
      
      // Set search query in store
      actions.setSearchQuery(query);

      // Prepare search options for backend
      const searchOptions: SearchOptions = {
        query: query,
        caseSensitive: caseSensitive,
        useLucene: useLuceneSyntax,
        selectedField: 'all', // Search all fields
        offset: 0,
        limit: SEARCH_LIMIT
      };

      // Use backend search for both regular and Lucene search
      const result = await SearchRecords(searchOptions);
      
      searchResults = {
        records: result.records,
        totalMatches: result.totalMatches,
        hasMore: result.hasMore
      };
      
      // Update the records store with search results
      actions.setRecords(result.records);
      
      // Dispatch search results event
      dispatch('searchResults', searchResults);
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Search failed';
      searchError = errorMessage;
      actions.setError(errorMessage);
      dispatch('searchError', errorMessage);
    } finally {
      isSearching = false;
    }
  }

  // Clear search and restore all records
  async function clearSearch() {
    searchInput = '';
    searchError = null;
    searchResults = null;
    
    // Clear the search query in store
    actions.setSearchQuery('');
    
    // Clear debounce timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
      debounceTimer = null;
    }

    // Clear records from store - let RecordViewer handle loading fresh records
    actions.setRecords([]);

    dispatch('searchCleared');
  }

  // Handle input change with debouncing
  function handleInputChange() {
    debounceSearch();
  }

  // Handle case sensitivity toggle
  function handleCaseSensitiveChange() {
    const query = useLuceneSyntax ? searchInput : searchInput.trim();
    if (query) {
      debounceSearch();
    }
  }



  // Handle Lucene syntax toggle
  function handleLuceneSyntaxChange() {
    const query = useLuceneSyntax ? searchInput : searchInput.trim();
    if (query) {
      debounceSearch();
    }
  }

  // Handle Enter key press for immediate search
  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }
      performSearch();
    } else if (event.key === 'Escape') {
      event.preventDefault();
      clearSearch();
    }
  }

  // Handle search button click
  function handleSearchClick() {
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }
    performSearch();
  }
</script>

<div class="search-component">
  <div class="search-controls">
    <!-- Search Input -->
    <div class="search-input-group">
      <div class="search-input-wrapper">
        <input
          type="text"
          class="search-input"
          placeholder={useLuceneSyntax ? 'e.g. level:error OR message:"failed"' : 'Search all fields...'}
          bind:value={searchInput}
          on:input={handleInputChange}
          on:keydown={handleKeyPress}
          disabled={!$hasFile || $isLoading}
          aria-label="Search query"
          autocapitalize="none"
          autocorrect="off"
          spellcheck="false"
        />
        
        <!-- Search Icon -->
        <div class="search-icon" class:searching={isSearching}>
          {#if isSearching}
            <div class="search-spinner"></div>
          {:else}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/>
              <path d="m21 21-4.35-4.35"/>
            </svg>
          {/if}
        </div>

        <!-- Clear Button -->
        {#if searchInput}
          <button
            class="clear-button"
            on:click={clearSearch}
            title="Clear search"
            aria-label="Clear search"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        {/if}
      </div>

      <!-- Search Button -->
      <button
        class="search-button"
        on:click={handleSearchClick}
        disabled={!$hasFile || $isLoading || !(useLuceneSyntax ? searchInput : searchInput.trim())}
        title="Search records"
      >
        Search
      </button>
    </div>

    <!-- Search Options -->
    <div class="search-options">
      <label class="checkbox-label">
        <input
          type="checkbox"
          bind:checked={useLuceneSyntax}
          on:change={handleLuceneSyntaxChange}
          disabled={!$hasFile || $isLoading}
        />
        <span class="checkbox-text">Lucene syntax (default)</span>
      </label>

      <label class="checkbox-label">
        <input
          type="checkbox"
          bind:checked={caseSensitive}
          on:change={handleCaseSensitiveChange}
          disabled={!$hasFile || $isLoading}
        />
        <span class="checkbox-text">Case sensitive</span>
      </label>
    </div>

    <!-- Lucene Syntax Help -->
    {#if useLuceneSyntax}
      <div class="lucene-help">
        <details>
          <summary>Lucene Syntax Examples</summary>
          <div class="help-examples">
            <code>field:value</code> - Search specific field<br>
            <code>term1 AND term2</code> - Both terms must match<br>
            <code>term1 OR term2</code> - Either term matches<br>
            <code>NOT term</code> - Exclude term<br>
          </div>
        </details>
      </div>
    {/if}
  </div>

  <!-- Search Results Summary -->
  {#if searchResults && (useLuceneSyntax ? searchInput : searchInput.trim())}
    <div class="search-results-summary">
      {#if searchResults.totalMatches === 0}
        <div class="no-results">
          <svg class="no-results-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <path d="m21 21-4.35-4.35"/>
            <line x1="8" y1="11" x2="14" y2="11"/>
          </svg>
          <h4>No results found</h4>
          <p>No records match your search query "<strong>{searchInput}</strong>"</p>
          <button class="clear-search-button" on:click={clearSearch}>
            Clear search
          </button>
        </div>
      {:else}
        <div class="results-info">
          <span class="results-count">
            Found <strong>{searchResults.totalMatches}</strong> 
            {searchResults.totalMatches === 1 ? 'record' : 'records'}
            matching "<strong>{searchInput}</strong>"
          </span>
          
          {#if searchResults.hasMore}
            <span class="results-pagination">
              Showing first {searchResults.records.length} results
            </span>
          {/if}
          
          <button class="clear-search-link" on:click={clearSearch}>
            Clear search
          </button>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Search Error -->
  {#if searchError}
    <div class="search-error">
      <svg class="error-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/>
        <line x1="15" y1="9" x2="9" y2="15"/>
        <line x1="9" y1="9" x2="15" y2="15"/>
      </svg>
      <div class="error-content">
        <strong>Search Error</strong>
        <p>{searchError}</p>
      </div>
      <button class="error-dismiss" on:click={() => searchError = null}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/>
          <line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>
  {/if}
</div>

<style>
  .search-component {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .search-controls {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .search-input-group {
    display: flex;
    gap: 0.5rem;
    align-items: stretch;
  }

  .search-input-wrapper {
    flex: 1;
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-input {
    width: 100%;
    padding: 0.5rem 2.5rem 0.5rem 2.5rem;
    border: 2px solid #e9ecef;
    border-radius: 6px;
    font-size: 0.85rem;
    transition: all 0.2s ease;
    background: white;
  }

  .search-input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
  }

  .search-input:disabled {
    background: #f8f9fa;
    color: #6c757d;
    cursor: not-allowed;
  }

  .search-icon {
    position: absolute;
    left: 0.75rem;
    width: 18px;
    height: 18px;
    color: #6c757d;
    pointer-events: none;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-icon.searching {
    color: #007bff;
  }

  .search-icon svg {
    width: 100%;
    height: 100%;
  }

  .search-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #e9ecef;
    border-top: 2px solid #007bff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .clear-button {
    position: absolute;
    right: 0.75rem;
    width: 18px;
    height: 18px;
    border: none;
    background: none;
    color: #6c757d;
    cursor: pointer;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 3px;
    transition: all 0.2s ease;
  }

  .clear-button:hover {
    color: #dc3545;
    background: rgba(220, 53, 69, 0.1);
  }

  .clear-button svg {
    width: 100%;
    height: 100%;
  }

  .search-button {
    padding: 0.5rem 1rem;
    background: #007bff;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    white-space: nowrap;
  }

  .search-button:hover:not(:disabled) {
    background: #0056b3;
    transform: translateY(-1px);
  }

  .search-button:disabled {
    background: #6c757d;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .search-options {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }



  .lucene-help {
    margin-top: 0.5rem;
  }

  .lucene-help details {
    background: #f0f8ff;
    border: 1px solid #b3d9ff;
    border-radius: 4px;
    padding: 0;
  }

  .lucene-help summary {
    padding: 0.5rem;
    cursor: pointer;
    font-size: 0.8rem;
    font-weight: 500;
    color: #0066cc;
    user-select: none;
  }

  .lucene-help summary:hover {
    background: #e6f3ff;
  }

  .help-examples {
    padding: 0.5rem;
    border-top: 1px solid #b3d9ff;
    font-size: 0.75rem;
    line-height: 1.4;
    color: #495057;
  }

  .help-examples code {
    background: #f1f3f4;
    padding: 0.1rem 0.3rem;
    border-radius: 2px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    color: #0066cc;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.9rem;
    color: #495057;
  }

  .checkbox-label input[type="checkbox"] {
    margin: 0;
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"]:disabled {
    cursor: not-allowed;
  }

  .checkbox-text {
    user-select: none;
  }

  .search-results-summary {
    background: #f8f9fa;
    border: 1px solid #e9ecef;
    border-radius: 6px;
    padding: 0.5rem;
  }

  .no-results {
    text-align: center;
    color: #6c757d;
  }

  .no-results-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 1rem;
    opacity: 0.5;
  }

  .no-results h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1.1rem;
    color: #495057;
  }

  .no-results p {
    margin: 0 0 1rem 0;
    font-size: 0.9rem;
  }

  .clear-search-button {
    background: #6c757d;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .clear-search-button:hover {
    background: #545b62;
  }

  .results-info {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 1rem;
    font-size: 0.9rem;
  }

  .results-count {
    color: #495057;
  }

  .results-pagination {
    color: #6c757d;
    font-size: 0.85rem;
  }

  .clear-search-link {
    background: none;
    border: none;
    color: #007bff;
    cursor: pointer;
    font-size: 0.85rem;
    text-decoration: underline;
    padding: 0;
  }

  .clear-search-link:hover {
    color: #0056b3;
  }

  .search-error {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    background: #f8d7da;
    border: 1px solid #f5c6cb;
    border-radius: 6px;
    padding: 1rem;
    color: #721c24;
  }

  .error-icon {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    margin-top: 0.1rem;
  }

  .error-content {
    flex: 1;
  }

  .error-content strong {
    display: block;
    margin-bottom: 0.25rem;
    font-size: 0.9rem;
  }

  .error-content p {
    margin: 0;
    font-size: 0.85rem;
    opacity: 0.9;
  }

  .error-dismiss {
    background: none;
    border: none;
    color: #721c24;
    cursor: pointer;
    padding: 0;
    width: 18px;
    height: 18px;
    flex-shrink: 0;
    opacity: 0.7;
    transition: opacity 0.2s ease;
  }

  .error-dismiss:hover {
    opacity: 1;
  }

  .error-dismiss svg {
    width: 100%;
    height: 100%;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .search-input-group {
      flex-direction: column;
    }

    .search-button {
      align-self: stretch;
    }

    .search-options {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }



    .results-info {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }
  }
</style>