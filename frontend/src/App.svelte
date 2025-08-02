<script lang="ts">
  import { onMount } from 'svelte';
  import { currentFile, isLoading, error, hasFile, actions } from './stores';
  import FileLoader from './components/FileLoader.svelte';
  import RecordViewer from './components/RecordViewer.svelte';
  import SearchComponent from './components/SearchComponent.svelte';
  import StatisticsPanel from './components/StatisticsPanel.svelte';
  import FieldVisibilityControls from './components/FieldVisibilityControls.svelte';
  import type { JSONLFile, JSONRecord, SearchResult } from './types';

  onMount(() => {
    // Initialize application
    actions.clearError();
  });

  // Clear error when user interacts
  function clearError() {
    actions.clearError();
  }

  // Handle file loaded event from FileLoader
  function handleFileLoaded(event: CustomEvent<JSONLFile>) {
    console.log('File loaded:', event.detail);
    // The FileLoader component already updates the stores, so we don't need to do anything here
  }

  // Handle error event from FileLoader
  function handleFileLoadError(event: CustomEvent<string>) {
    console.error('File load error:', event.detail);
    // The FileLoader component already updates the error store, so we don't need to do anything here
  }

  // Handle record selection from RecordViewer
  function handleRecordSelected(event: CustomEvent<JSONRecord>) {
    console.log('Record selected:', event.detail);
    // Future: Could show record details in a modal or side panel
  }

  // Handle navigation requests from RecordViewer
  function handleNavigationRequest(event: CustomEvent<{ type: string, lineNumber: number }>) {
    console.log('Navigation request:', event.detail);
    // Future: Could implement "go to line" functionality
  }

  // Handle search results from SearchComponent
  function handleSearchResults(event: CustomEvent<SearchResult>) {
    console.log('Search results:', event.detail);
    // The SearchComponent already updates the stores, so we don't need to do anything here
  }

  // Handle search errors from SearchComponent
  function handleSearchError(event: CustomEvent<string>) {
    console.error('Search error:', event.detail);
    // The SearchComponent already updates the error store, so we don't need to do anything here
  }

  // Handle search cleared from SearchComponent
  function handleSearchCleared() {
    console.log('Search cleared');
    // The SearchComponent already restores the records, so we don't need to do anything here
  }

  // Handle statistics panel events
  function handleStatisticsError(event: CustomEvent<string>) {
    console.error('Statistics error:', event.detail);
    actions.setError(event.detail);
  }

  function handleReloadRequested() {
    console.log('File reload requested');
    // Future: Could implement file reload functionality
    // For now, just show a message to the user
    actions.setError('File reload functionality will be implemented in a future update');
  }
</script>

<main class="app">
  {#if $error}
    <div class="error-banner" on:click={clearError} on:keydown={(e) => e.key === 'Enter' && clearError()} tabindex="0" role="button">
      <span class="error-text">{$error}</span>
      <button class="error-close" on:click={clearError}>×</button>
    </div>
  {/if}

  <div class="app-content">
    {#if $isLoading}
      <div class="loading-overlay">
        <div class="loading-spinner"></div>
        <p>Loading...</p>
      </div>
    {/if}

    <div class="main-layout">
      <!-- File Loader Section -->
      <section class="file-loader-section">
        <FileLoader 
          on:fileLoaded={handleFileLoaded}
          on:error={handleFileLoadError}
        />
      </section>

      {#if $hasFile}
        <!-- Search Section -->
        <section class="search-section">
          <SearchComponent 
            on:searchResults={handleSearchResults}
            on:searchError={handleSearchError}
            on:searchCleared={handleSearchCleared}
          />
        </section>

        <!-- Field Visibility Controls Section -->
        <section class="field-visibility-section">
          <FieldVisibilityControls />
        </section>

        <!-- Main Content Area -->
        <div class="content-area">
          <!-- Record Viewer Section -->
          <section class="record-viewer-section">
            <RecordViewer 
              on:recordSelected={handleRecordSelected}
              on:navigationRequest={handleNavigationRequest}
            />
          </section>

          <!-- Statistics Panel Section -->
          <aside class="statistics-section">
            <StatisticsPanel 
              on:error={handleStatisticsError}
              on:reloadRequested={handleReloadRequested}
            />
          </aside>
        </div>


      {:else}
        <!-- Welcome State -->
        <div class="welcome-state">
          <h2>Welcome to JSONL Viewer</h2>
          <p>Load a JSONL file or paste data from clipboard to get started.</p>
        </div>
      {/if}
    </div>
  </div>
</main>

<style>
  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }

  .error-banner {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    background: #dc3545;
    color: white;
    padding: 0.75rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    cursor: pointer;
    z-index: 1000;
  }

  .error-text {
    flex: 1;
  }

  .error-close {
    background: none;
    border: none;
    color: white;
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    margin-left: 1rem;
  }

  .app-content {
    flex: 1;
    position: relative;
    overflow: hidden;
  }

  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(255, 255, 255, 0.9);
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    z-index: 999;
  }

  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #f3f3f3;
    border-top: 4px solid #007bff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .main-layout {
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 0.5rem;
    gap: 0.5rem;
  }

  .file-loader-section {
    background: white;
    border: 1px solid #e9ecef;
    border-radius: 8px;
    padding: 0.5rem;
    flex-shrink: 0;
    max-height: 10vh;
    overflow: hidden;
  }

  .search-section {
    background: white;
    border: 1px solid #e9ecef;
    border-radius: 8px;
    padding: 0.25rem;
    flex-shrink: 0;
  }

  .field-visibility-section {
    flex-shrink: 0;
  }

  .content-area {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr 300px;
    grid-template-rows: 1fr;
    gap: 0.5rem;
    min-height: 0;
    max-height: 100%;
    overflow: hidden;
  }

  .record-viewer-section {
    background: white;
    border: 1px solid #e9ecef;
    border-radius: 8px;
    padding: 1rem;
    overflow: hidden;
    min-height: 0;
    max-height: 100%;
    display: flex;
    flex-direction: column;
  }

  .statistics-section {
    background: white;
    border: 1px solid #e9ecef;
    border-radius: 8px;
    padding: 1rem;
    min-height: 0;
    max-height: 100%;
    overflow-y: auto;
  }

  .welcome-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    color: #6c757d;
  }

  .welcome-state h2 {
    margin-bottom: 0.5rem;
    color: #495057;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .content-area {
      grid-template-columns: 1fr;
    }
    
    .main-layout {
      padding: 0.5rem;
    }
    

  }
</style>
