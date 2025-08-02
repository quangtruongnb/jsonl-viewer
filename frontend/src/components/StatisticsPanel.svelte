<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { currentFile, hasFile } from '../stores';
  import { GetFileStats, GetCommonFields, GetFileModificationInfo } from '../../wailsjs/go/main/App.js';
  import type { JSONLFile, FileStats } from '../types';

  const dispatch = createEventDispatcher<{
    error: string;
    reloadRequested: void;
  }>();

  let fileStats: FileStats | null = null;
  let commonFields: string[] = [];
  let modificationInfo: any = null;
  let isLoading = false;
  let error: string | null = null;

  // Reactive statement to load statistics when file changes
  $: if ($hasFile && $currentFile) {
    loadStatistics();
  }

  // Load file statistics and common fields
  async function loadStatistics() {
    if (!$currentFile) return;

    try {
      isLoading = true;
      error = null;

      // Load file statistics and common fields in parallel
      const [statsResult, fieldsResult, modInfoResult] = await Promise.all([
        GetFileStats().catch(err => null),
        GetCommonFields().catch(err => []),
        GetFileModificationInfo().catch(err => null)
      ]);

      fileStats = statsResult;
      commonFields = fieldsResult;
      modificationInfo = modInfoResult;

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load statistics';
      error = errorMessage;
      dispatch('error', errorMessage);
    } finally {
      isLoading = false;
    }
  }

  // Format file size for display
  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  // Format date for display
  function formatDate(dateString: string): string {
    try {
      const date = new Date(dateString);
      return date.toLocaleString();
    } catch {
      return dateString;
    }
  }

  // Handle reload request
  function handleReloadRequest() {
    dispatch('reloadRequested');
  }

  // Refresh statistics manually
  function refreshStatistics() {
    loadStatistics();
  }

  onMount(() => {
    if ($hasFile && $currentFile) {
      loadStatistics();
    }
  });
</script>

{#if $hasFile && $currentFile}
  <div class="statistics-panel">
    <div class="panel-header">
      <h3>File Statistics</h3>
      <button 
        class="refresh-button"
        on:click={refreshStatistics}
        disabled={isLoading}
        title="Refresh statistics"
      >
        <svg class="refresh-icon" class:spinning={isLoading} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/>
          <polyline points="1 20 1 14 7 14"/>
          <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/>
        </svg>
      </button>
    </div>

    {#if isLoading}
      <div class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading statistics...</p>
      </div>
    {:else if error}
      <div class="error-state">
        <p class="error-message">{error}</p>
        <button class="retry-button" on:click={refreshStatistics}>
          Retry
        </button>
      </div>
    {:else}
      <!-- File Information Section -->
      <div class="stats-section">
        <h4>File Information</h4>
        <div class="stats-grid">
          <div class="stat-item">
            <span class="stat-label">Name:</span>
            <span class="stat-value" title={$currentFile.name}>{$currentFile.name}</span>
          </div>
          
          <div class="stat-item">
            <span class="stat-label">Size:</span>
            <span class="stat-value">{formatFileSize($currentFile.size)}</span>
          </div>
          
          <div class="stat-item">
            <span class="stat-label">Records:</span>
            <span class="stat-value">{$currentFile.records.toLocaleString()}</span>
          </div>
          
          <div class="stat-item">
            <span class="stat-label">Loaded:</span>
            <span class="stat-value">{formatDate($currentFile.loadedAt)}</span>
          </div>
        </div>
      </div>

      <!-- File Path Section -->
      <div class="stats-section">
        <h4>File Path</h4>
        <div class="path-display">
          <code class="file-path" title={$currentFile.path}>{$currentFile.path}</code>
          {#if $currentFile.path !== '<clipboard>'}
            <button 
              class="copy-path-button"
              on:click={() => navigator.clipboard?.writeText($currentFile.path)}
              title="Copy file path"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
              </svg>
            </button>
          {/if}
        </div>
      </div>

      <!-- Detailed Statistics Section -->
      {#if fileStats}
        <div class="stats-section">
          <h4>Detailed Statistics</h4>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-label">Total Lines:</span>
              <span class="stat-value">{fileStats.totalLines.toLocaleString()}</span>
            </div>
            
            <div class="stat-item">
              <span class="stat-label">Valid Records:</span>
              <span class="stat-value">{fileStats.validRecords.toLocaleString()}</span>
            </div>
            
            {#if fileStats.invalidLines && fileStats.invalidLines.length > 0}
              <div class="stat-item">
                <span class="stat-label">Invalid Lines:</span>
                <span class="stat-value error-text">{fileStats.invalidLines.length}</span>
              </div>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Common Fields Section -->
      {#if commonFields && commonFields.length > 0}
        <div class="stats-section">
          <h4>Common Fields</h4>
          <p class="section-description">Fields that appear in at least 50% of records:</p>
          <div class="fields-list">
            {#each commonFields as field}
              <span class="field-tag">{field}</span>
            {/each}
          </div>
        </div>
      {:else if commonFields}
        <div class="stats-section">
          <h4>Common Fields</h4>
          <p class="no-data">No common fields found (fields must appear in at least 50% of records)</p>
        </div>
      {/if}

      <!-- File Modification Status -->
      {#if modificationInfo && modificationInfo.filePath !== '<clipboard>'}
        <div class="stats-section">
          <h4>File Status</h4>
          {#if modificationInfo.isModified}
            <div class="modification-warning">
              <svg class="warning-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/>
                <line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
              <div class="warning-content">
                <p class="warning-text">File has been modified since loading</p>
                <button class="reload-button" on:click={handleReloadRequest}>
                  Reload File
                </button>
              </div>
            </div>
          {:else}
            <div class="status-ok">
              <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              <span>File is up to date</span>
            </div>
          {/if}
        </div>
      {/if}
    {/if}
  </div>
{:else}
  <div class="no-file-state">
    <svg class="no-file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
      <polyline points="14,2 14,8 20,8"/>
      <line x1="16" y1="13" x2="8" y2="13"/>
      <line x1="16" y1="17" x2="8" y2="17"/>
      <line x1="10" y1="9" x2="8" y2="9"/>
    </svg>
    <p>No file loaded</p>
    <small>Load a JSONL file to view statistics</small>
  </div>
{/if}

<style>
  .statistics-panel {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    height: 100%;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .panel-header h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #212529;
  }

  .refresh-button {
    background: none;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    padding: 0.25rem;
    cursor: pointer;
    color: #6c757d;
    transition: all 0.2s ease;
  }

  .refresh-button:hover:not(:disabled) {
    background: #f8f9fa;
    border-color: #adb5bd;
    color: #495057;
  }

  .refresh-button:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  .refresh-icon {
    width: 16px;
    height: 16px;
  }

  .refresh-icon.spinning {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    color: #6c757d;
  }

  .loading-spinner {
    width: 24px;
    height: 24px;
    border: 2px solid #f3f3f3;
    border-top: 2px solid #007bff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 0.5rem;
  }

  .error-state {
    padding: 1rem;
    background: #f8d7da;
    border: 1px solid #f5c6cb;
    border-radius: 4px;
    text-align: center;
  }

  .error-message {
    margin: 0 0 0.5rem 0;
    color: #721c24;
    font-size: 0.9rem;
  }

  .retry-button {
    background: #dc3545;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.25rem 0.75rem;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .retry-button:hover {
    background: #c82333;
  }

  .stats-section {
    background: #f8f9fa;
    border: 1px solid #e9ecef;
    border-radius: 6px;
    padding: 1rem;
  }

  .stats-section h4 {
    margin: 0 0 0.75rem 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: #495057;
  }

  .section-description {
    margin: 0 0 0.75rem 0;
    font-size: 0.8rem;
    color: #6c757d;
    font-style: italic;
  }

  .stats-grid {
    display: grid;
    gap: 0.5rem;
  }

  .stat-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.25rem 0;
    border-bottom: 1px solid #e9ecef;
  }

  .stat-item:last-child {
    border-bottom: none;
  }

  .stat-label {
    font-size: 0.85rem;
    color: #6c757d;
    font-weight: 500;
  }

  .stat-value {
    font-size: 0.85rem;
    color: #212529;
    font-weight: 500;
    text-align: right;
    word-break: break-all;
  }

  .error-text {
    color: #dc3545;
  }

  .path-display {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: white;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    padding: 0.5rem;
  }

  .file-path {
    flex: 1;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.8rem;
    color: #495057;
    background: none;
    border: none;
    word-break: break-all;
    line-height: 1.4;
  }

  .copy-path-button {
    background: none;
    border: none;
    color: #6c757d;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 3px;
    transition: all 0.2s ease;
    flex-shrink: 0;
  }

  .copy-path-button:hover {
    background: #e9ecef;
    color: #495057;
  }

  .copy-path-button svg {
    width: 14px;
    height: 14px;
  }

  .fields-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .field-tag {
    background: #007bff;
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .no-data {
    margin: 0;
    color: #6c757d;
    font-size: 0.85rem;
    font-style: italic;
  }

  .modification-warning {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    background: #fff3cd;
    border: 1px solid #ffeaa7;
    border-radius: 4px;
    padding: 0.75rem;
  }

  .warning-icon {
    width: 18px;
    height: 18px;
    color: #856404;
    flex-shrink: 0;
    margin-top: 0.1rem;
  }

  .warning-content {
    flex: 1;
  }

  .warning-text {
    margin: 0 0 0.5rem 0;
    color: #856404;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .reload-button {
    background: #ffc107;
    color: #212529;
    border: none;
    border-radius: 4px;
    padding: 0.25rem 0.75rem;
    font-size: 0.8rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .reload-button:hover {
    background: #e0a800;
  }

  .status-ok {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: #155724;
    font-size: 0.85rem;
  }

  .check-icon {
    width: 16px;
    height: 16px;
    color: #28a745;
  }

  .no-file-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    text-align: center;
    color: #6c757d;
    height: 100%;
  }

  .no-file-icon {
    width: 48px;
    height: 48px;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  .no-file-state p {
    margin: 0 0 0.25rem 0;
    font-weight: 500;
  }

  .no-file-state small {
    font-size: 0.8rem;
    opacity: 0.7;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .stats-section {
      padding: 0.75rem;
    }
    
    .stat-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.25rem;
    }
    
    .stat-value {
      text-align: left;
    }
    
    .path-display {
      flex-direction: column;
      align-items: stretch;
    }
    
    .copy-path-button {
      align-self: flex-end;
    }
  }
</style>