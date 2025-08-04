<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import {
    GetRecords,
    GetTotalRecordCount,
    ExportSearchResults,
  } from "../../wailsjs/go/main/App.js";
  import {
    currentFile,
    isLoading,
    error,
    actions,
    totalRecords,
    records,
    searchQuery,
    fieldsToShow,
    fieldsToHide,
  } from "../stores";
  import type { JSONRecord } from "../types";
  import { LogError, LogInfo } from "../../wailsjs/runtime";

  const dispatch = createEventDispatcher<{
    recordSelected: JSONRecord;
    navigationRequest: { type: "goto"; lineNumber: number };
  }>();

  // Pagination configuration
  let recordsPerPage = 50;
  const PAGE_SIZE_OPTIONS = [10, 25, 50, 100, 200, 500];

  // Component state
  let currentPage = 1;
  let totalRecordsCount = 0;
  let currentRecords: JSONRecord[] = [];
  let isLoadingRecords = false;
  
  // Modal state for beautified JSON view
  let showJsonModal = false;
  let selectedRecord: JSONRecord | null = null;
  let highlightedJson = '';
  
  // Export state
  let isExporting = false;

  // Pagination calculations
  $: totalPages = Math.ceil(totalRecordsCount / recordsPerPage);
  $: startRecord = (currentPage - 1) * recordsPerPage + 1;
  $: endRecord = Math.min(currentPage * recordsPerPage, totalRecordsCount);

  // Load records when file changes
  $: if ($currentFile) {
    initializeRecords();
  }

  // Handle search results from global store
  $: if ($searchQuery && $records.length > 0) {
    // Search results are available, use them for pagination
    totalRecordsCount = $records.length;
    currentPage = 1; // Reset to first page for search results
    paginateSearchResults();
  } else if (!$searchQuery && $currentFile) {
    currentPage = 1;
    loadCurrentPage();
  }

  // Create a reactive key that changes when field visibility changes
  $: fieldVisibilityKey = JSON.stringify($fieldsToShow) + JSON.stringify($fieldsToHide);

  // Create a reactive function that updates when field visibility changes
  $: getDisplayJSON = (record: JSONRecord) => {
    console.log('getDisplayJSON called with fieldVisibilityKey:', fieldVisibilityKey);
    return displayJSON(record);
  };

  // Handle keyboard events for modal
  $: if (showJsonModal) {
    // Modal is open, keyboard events will be handled by the handleKeydown function
  }

  // Update highlighted JSON when selected record changes
  $: if (selectedRecord) {
    const beautified = getBeautifiedJson(selectedRecord);
    console.log('Original JSON:', beautified);
    highlightedJson = highlightJson(beautified);
    console.log('Highlighted JSON:', highlightedJson);
  }

  // Add keyboard event listener
  onMount(() => {
    document.addEventListener('keydown', handleKeydown);
    return () => {
      document.removeEventListener('keydown', handleKeydown);
    };
  });

    // Export functionality
  async function exportSearchResults() {
    if (isExporting) return;
    
    try {
      isExporting = true;
      console.log('Exporting all search results...');
      
      // Use backend export function with search query and field visibility
      const filePath = await ExportSearchResults($searchQuery || '', $fieldsToShow, $fieldsToHide);
      console.log('Export completed, file saved to:', filePath);
      
      // Show success notification
      showNotification(`Export completed successfully! File saved to: ${filePath}`, 'success');
      
    } catch (error) {
      console.error('Export failed:', error);
      showNotification('Export failed: ' + error.message, 'error');
    } finally {
      isExporting = false;
    }
  }

  // Notification system
  let notificationMessage = '';
  let notificationType = 'info';
  let showNotificationFlag = false;

  function showNotification(message, type = 'info') {
    notificationMessage = message;
    notificationType = type;
    showNotificationFlag = true;
    
    // Auto-hide after 5 seconds
    setTimeout(() => {
      showNotificationFlag = false;
    }, 5000);
  }



  async function initializeRecords() {
    if (!$currentFile) {
      currentRecords = [];
      totalRecordsCount = 0;
      currentPage = 1;
      actions.setTotalRecords(0);
      return;
    }

    try {
      isLoadingRecords = true;
      actions.clearError();

      // Get total record count
      const totalCount = await GetTotalRecordCount();
      totalRecordsCount = totalCount;
      actions.setTotalRecords(totalCount);

      // Reset to first page
      currentPage = 1;

      // Load first page
      await loadCurrentPage();
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to initialize records";
      actions.setError(errorMessage);
      LogError(errorMessage)
    } finally {
      isLoadingRecords = false;
      LogInfo("Records initialized")
    }
  }

  async function loadCurrentPage() {
    if (!$currentFile || isLoadingRecords || $searchQuery) return;

    try {
      isLoadingRecords = true;
      const offset = (currentPage - 1) * recordsPerPage;
      const paginatedRecords = await GetRecords(offset, recordsPerPage);
      currentRecords = paginatedRecords.records;
    } catch (err) {
      console.error("Failed to load page:", err);
      actions.setError("Failed to load records");
      LogError("Failed to load records")
    } finally {
      isLoadingRecords = false;
      LogInfo("Loaded page")
    }
  }

  function paginateSearchResults() {
    if (!$records.length) {
      currentRecords = [];
      return;
    }

    const startIndex = (currentPage - 1) * recordsPerPage;
    const endIndex = startIndex + recordsPerPage;
    currentRecords = $records.slice(startIndex, endIndex);
  }

  function goToPage(page: number) {
    if (page < 1 || page > totalPages || page === currentPage) return;

    currentPage = page;

    if ($searchQuery) {
      paginateSearchResults();
    } else {
      loadCurrentPage();
    }
  }

  function goToFirstPage() {
    goToPage(1);
  }

  function goToLastPage() {
    goToPage(totalPages);
  }

  function goToPreviousPage() {
    goToPage(currentPage - 1);
  }

  function goToNextPage() {
    goToPage(currentPage + 1);
  }

  function handlePageSizeChange(newPageSize: number) {
    recordsPerPage = newPageSize;
    currentPage = 1; // Reset to first page when changing page size

    if ($searchQuery) {
      paginateSearchResults();
    } else {
      loadCurrentPage();
    }
  }

  function handleRecordClick(record: JSONRecord) {
    dispatch("recordSelected", record);
  }

  function handleGotoLine(lineNumber: number) {
    dispatch("navigationRequest", { type: "goto", lineNumber });
  }

  function handleViewJson(record: JSONRecord, event: Event) {
    event.stopPropagation(); // Prevent triggering the record click
    selectedRecord = record;
    showJsonModal = true;
  }

  function closeJsonModal() {
    showJsonModal = false;
    selectedRecord = null;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && showJsonModal) {
      closeJsonModal();
    }
  }

  function getBeautifiedJson(record: JSONRecord): string {
    try {
      const jsonObj = record.content || JSON.parse(record.rawJSON);
      return JSON.stringify(jsonObj, null, 2);
    } catch (error) {
      return record.rawJSON;
    }
  }

  function highlightJson(jsonString: string): string {
    console.log('Highlighting JSON:', jsonString);
    
    let highlighted = jsonString
      // First, highlight punctuation to avoid conflicts
      .replace(/([{}[\],])/g, '<span class="json-punctuation">$1</span>')
      // Then highlight keys (property names) - match quoted strings followed by colon
      .replace(/"([^"]+)":/g, '<span class="json-key">"$1"</span>:')
      // Then highlight string values - match colon followed by quoted strings
      .replace(/:\s*"([^"]*)"/g, ': <span class="json-string">"$1"</span>')
      // Then highlight numbers (including decimals and negative numbers)
      .replace(/:\s*(-?\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
      // Then highlight booleans
      .replace(/:\s*(true|false)/g, ': <span class="json-boolean">$1</span>')
      // Finally highlight null
      .replace(/:\s*(null)/g, ': <span class="json-null">$1</span>');
    
    console.log('Highlighted result:', highlighted);
    return highlighted;
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).then(() => {
      // Could add a toast notification here
      console.log('JSON copied to clipboard');
    }).catch(err => {
      console.error('Failed to copy to clipboard:', err);
    });
  }

  // Filter and display JSON based on field visibility settings
  function displayJSON(record: JSONRecord): string {
    if (!record.rawJSON) return "";
    
    console.log('displayJSON called for record:', record.lineNumber);
    
    try {
      // Parse the JSON content
      const jsonObj = record.content || JSON.parse(record.rawJSON);
      
      // Apply field visibility filtering
      const filteredObj = filterFields(jsonObj);
      
      // Return formatted JSON string
      const result = JSON.stringify(filteredObj);
      console.log('displayJSON result for record', record.lineNumber, ':', result);
      return result;
    } catch (error) {
      // If parsing fails, return the raw JSON
      return record.rawJSON;
    }
  }



  // Filter object fields based on visibility settings
  function filterFields(obj: Record<string, any>): Record<string, any> {
    if (!obj || typeof obj !== 'object') return obj;
    
    console.log('filterFields called with:', { fieldsToShow: $fieldsToShow, fieldsToHide: $fieldsToHide });
    
    const filtered: Record<string, any> = {};
    
    // Get all field names from the object
    const allFields = Object.keys(obj);
    
    // Determine which fields to include
    let fieldsToInclude: string[];
    
    if ($fieldsToShow.length > 0) {
      // If specific fields are selected to show, only show those
      fieldsToInclude = allFields.filter(field => $fieldsToShow.includes(field));
    } else {
      // If no specific fields to show, show all except hidden ones
      fieldsToInclude = allFields.filter(field => !$fieldsToHide.includes(field));
    }
    
    // Build the filtered object
    for (const field of fieldsToInclude) {
      filtered[field] = obj[field];
    }
    
    console.log('filterFields result:', filtered);
    return filtered;
  }
</script>

<div class="record-viewer">
  {#if !$currentFile}
    <div class="empty-state">
      <div class="empty-icon">📄</div>
      <h3>No File Loaded</h3>
      <p>Load a JSONL file to view records</p>
    </div>
  {:else if isLoadingRecords}
    <div class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading records...</p>
    </div>
  {:else if totalRecordsCount === 0}
    <div class="empty-state">
      <div class="empty-icon">📭</div>
      <h3>No Records Found</h3>
      <p>The loaded file contains no valid JSONL records</p>
    </div>
  {:else}
    <!-- Pagination Header -->
    <div class="pagination-header">
      <div class="pagination-info">
        <span
          >Showing {startRecord}-{endRecord} of {totalRecordsCount.toLocaleString()}
          records</span
        >
        {#if $searchQuery}
          <span class="search-indicator">• Search results</span>
        {/if}
      </div>

      <div class="page-size-selector">
        <label for="pageSize">Show:</label>
        <select
          id="pageSize"
          bind:value={recordsPerPage}
          on:change={() => handlePageSizeChange(recordsPerPage)}
        >
          {#each PAGE_SIZE_OPTIONS as size}
            <option value={size}>{size}</option>
          {/each}
        </select>
        <span>per page</span>
      </div>

      <div class="export-controls">
        <button
          class="export-btn"
          on:click={exportSearchResults}
          title="Export current results with visibility settings"
          disabled={currentRecords.length === 0 || isExporting}
          class:exporting={isExporting}
        >
          {#if isExporting}
            <svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" stroke-dasharray="31.416" stroke-dashoffset="31.416">
                <animate attributeName="stroke-dasharray" dur="2s" values="0 31.416;15.708 15.708;0 31.416" repeatCount="indefinite"/>
                <animate attributeName="stroke-dashoffset" dur="2s" values="0;-15.708;-31.416" repeatCount="indefinite"/>
              </circle>
            </svg>
            Exporting...
          {:else}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7,10 12,15 17,10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
            Export
          {/if}
        </button>
      </div>

      <div class="pagination-controls">
        <button
          class="page-btn"
          on:click={goToFirstPage}
          disabled={currentPage === 1}
          title="First page"
        >
          ⟪
        </button>
        <button
          class="page-btn"
          on:click={goToPreviousPage}
          disabled={currentPage === 1}
          title="Previous page"
        >
          ⟨
        </button>

        <span class="page-info">
          Page {currentPage} of {totalPages}
        </span>

        <button
          class="page-btn"
          on:click={goToNextPage}
          disabled={currentPage === totalPages}
          title="Next page"
        >
          ⟩
        </button>
        <button
          class="page-btn"
          on:click={goToLastPage}
          disabled={currentPage === totalPages}
          title="Last page"
        >
          ⟫
        </button>
      </div>
    </div>

    <!-- Records Container -->
    <div class="records-container">
      {#if isLoadingRecords}
        <div class="loading-overlay">
          <div class="loading-spinner"></div>
          <p>Loading page {currentPage}...</p>
        </div>
      {/if}

      {#each currentRecords as record}
        <div
          class="record-item"
          on:click={() => handleRecordClick(record)}
          on:keydown={(e) => e.key === "Enter" && handleRecordClick(record)}
          tabindex="0"
          role="button"
        >
          <div class="line-number-container">
            <span class="line-number">{record.lineNumber}</span>
          </div>
          <div class="action-container">
            <button
              class="view-json-btn"
              on:click={(e) => handleViewJson(record, e)}
              title="View beautified JSON"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                <circle cx="12" cy="12" r="3"></circle>
              </svg>
            </button>
          </div>
          <div class="record-content">
            <pre class="json-content">{getDisplayJSON(record)}</pre>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <!-- JSON Modal -->
  {#if showJsonModal && selectedRecord}
    <div class="modal-overlay" on:click={closeJsonModal} on:keydown={(e) => e.key === 'Escape' && closeJsonModal()}>
      <div class="modal-content" on:click|stopPropagation on:keydown|stopPropagation>
        <div class="modal-header">
          <h3>Beautified JSON - Line {selectedRecord.lineNumber}</h3>
          <div class="modal-actions">
            <button 
              class="modal-copy-btn" 
              on:click={() => copyToClipboard(getBeautifiedJson(selectedRecord))} 
              title="Copy JSON to clipboard"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
              </svg>
            </button>
            <button class="modal-close-btn" on:click={closeJsonModal} title="Close">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>
        </div>
        <div class="modal-body">
          <div class="json-container">
            <pre class="beautified-json">{@html highlightedJson}</pre>
            <!-- Dummy elements to ensure CSS classes are not removed during build -->
            <span class="json-key" style="display: none;"></span>
            <span class="json-string" style="display: none;"></span>
            <span class="json-number" style="display: none;"></span>
            <span class="json-boolean" style="display: none;"></span>
            <span class="json-null" style="display: none;"></span>
            <span class="json-punctuation" style="display: none;"></span>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Notification -->
  {#if showNotificationFlag}
    <div class="notification notification-{notificationType}">
      <span class="notification-message">{notificationMessage}</span>
      <button class="notification-close" on:click={() => showNotificationFlag = false}>×</button>
    </div>
  {/if}
</div>

<style>
  .record-viewer {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: white;
    min-height: 0;
  }

  .empty-state,
  .loading-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    color: #6c757d;
    text-align: center;
    padding: 2rem;
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  .empty-state h3 {
    margin: 0 0 0.5rem 0;
    color: #495057;
    font-weight: 600;
  }

  .empty-state p,
  .loading-state p {
    margin: 0;
    font-size: 0.9rem;
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

  .pagination-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: #f8f9fa;
    border-bottom: 1px solid #e9ecef;
    flex-shrink: 0;
    gap: 1rem;
  }

  .pagination-info {
    font-size: 0.85rem;
    color: #6c757d;
  }

  .search-indicator {
    color: #007bff;
    font-weight: 500;
  }

  .page-size-selector {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    color: #495057;
  }

  .page-size-selector label {
    font-weight: 500;
  }

  .page-size-selector select {
    padding: 0.25rem 0.5rem;
    border: 1px solid #ced4da;
    border-radius: 4px;
    background: white;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .page-size-selector select:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  .export-controls {
    display: flex;
    align-items: center;
  }

  .export-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: #28a745;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.5rem 0.75rem;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .export-btn:hover:not(:disabled) {
    background: #218838;
  }

  .export-btn:disabled {
    background: #6c757d;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .export-btn svg {
    width: 16px;
    height: 16px;
  }

  .export-btn.exporting {
    background: #6c757d;
    cursor: not-allowed;
  }

  .export-btn .spinner {
    animation: spin 1s linear infinite;
  }

  .pagination-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .page-btn {
    background: white;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    padding: 0.5rem 0.75rem;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .page-btn:hover:not(:disabled) {
    background: #e9ecef;
    border-color: #adb5bd;
  }

  .page-btn:disabled {
    background: #f8f9fa;
    color: #6c757d;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .page-info {
    font-size: 0.85rem;
    color: #495057;
    font-weight: 500;
    padding: 0 0.5rem;
  }

  .records-container {
    flex: 1;
    overflow: auto;
    position: relative;
    width: 100%;
    min-height: 0;
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
    z-index: 10;
  }

  .record-item {
    cursor: pointer;
    transition: background-color 0.15s ease;
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    border-bottom: 1px solid #f0f0f0;
    width: max-content;
    min-width: 100%;
  }

  .record-item:hover {
    background-color: #f8f9fa;
  }

  .record-item:focus {
    outline: 2px solid #007bff;
    outline-offset: -2px;
    background-color: #f0f8ff;
  }

  .line-number-container {
    flex-shrink: 0;
    width: 60px;
    text-align: right;
    padding-right: 0.5rem;
    border-right: 1px solid #e9ecef;
    padding-top: 0.25rem;
  }

  .line-number {
    font-size: 0.7rem;
    font-weight: 500;
    color: #6c757d;
    font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
  }

  .view-json-btn {
    background: none;
    border: none;
    padding: 0.5rem;
    cursor: pointer;
    color: #6c757d;
    border-radius: 4px;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .view-json-btn:hover {
    background: #e9ecef;
    color: #495057;
  }

  .view-json-btn svg {
    width: 16px;
    height: 16px;
  }

  .record-content {
    flex: 1;
    min-width: 0;
    overflow-x: visible;
    text-align: left;
  }

  .action-container {
    flex-shrink: 0;
    width: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 0.25rem;
    border-right: 1px solid #e9ecef;
  }

  .json-content {
    margin: 0;
    padding: 0.25rem 0;
    background: transparent;
    border: none;
    font-family: "Monaco", "Monaco", "Menlo", "Ubuntu Mono", monospace;
    font-size: 0.65rem;
    line-height: 1.2;
    overflow-x: visible;
    white-space: nowrap;
    word-wrap: normal;
    max-height: none;
    overflow-y: visible;
    color: #212529;
    display: inline-block;
    text-align: left;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .pagination-header {
      flex-direction: column;
      gap: 0.5rem;
      align-items: stretch;
    }

    .pagination-info {
      text-align: center;
    }

    .page-size-selector {
      justify-content: center;
    }

    .pagination-controls {
      justify-content: center;
    }

    .record-item {
      padding: 0.5rem;
    }

    .line-number-container {
      width: 50px;
    }

    .action-container {
      width: 20px;
    }

    .json-content {
      font-size: 0.6rem;
    }

    .line-number {
      font-size: 0.65rem;
    }

    .view-json-btn svg {
      width: 14px;
      height: 14px;
    }
  }

  /* Scrollbar styling */
  .records-container::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .records-container::-webkit-scrollbar-track {
    background: #f1f1f1;
  }

  .records-container::-webkit-scrollbar-thumb {
    background: #c1c1c1;
    border-radius: 4px;
  }

  .records-container::-webkit-scrollbar-thumb:hover {
    background: #a8a8a8;
  }

  .records-container::-webkit-scrollbar-corner {
    background: #f1f1f1;
  }

  /* Modal styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-content {
    background: white;
    border-radius: 8px;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    max-width: 90vw;
    max-height: 90vh;
    width: 800px;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid #e9ecef;
    background: #f8f9fa;
    border-radius: 8px 8px 0 0;
  }

  .modal-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #495057;
  }

  .modal-close-btn {
    background: none;
    border: none;
    padding: 0.5rem;
    cursor: pointer;
    color: #6c757d;
    border-radius: 4px;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-close-btn:hover {
    background: #e9ecef;
    color: #495057;
  }

  .modal-close-btn svg {
    width: 16px;
    height: 16px;
  }

  .modal-copy-btn {
    background: none;
    border: none;
    padding: 0.5rem;
    cursor: pointer;
    color: #6c757d;
    border-radius: 4px;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-copy-btn:hover {
    background: #e9ecef;
    color: #007bff;
  }

  .modal-copy-btn svg {
    width: 16px;
    height: 16px;
  }

  .modal-body {
    padding: 1.5rem;
    overflow: auto;
    flex: 1;
    max-height: 70vh;
  }

  .json-container {
    background: #1e1e1e;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid #333;
  }

  .beautified-json {
    margin: 0;
    padding: 1.5rem;
    background: #1e1e1e;
    font-family: "Monaco", "Menlo", "Ubuntu Mono", "Consolas", monospace;
    font-size: 0.9rem;
    line-height: 1.5;
    color: #d4d4d4;
    white-space: pre;
    overflow-x: auto;
    text-align: left;
    border: none;
    border-radius: 0;
  }

  /* JSON Syntax Highlighting - These classes are dynamically generated in JavaScript */
  /* DO NOT REMOVE - Used for JSON syntax highlighting in modal */
  .json-key {
    color: #4fc1ff !important;
    font-weight: 500;
  }

  .json-string {
    color: #ff8c42 !important;
  }

  .json-number {
    color: #98c379 !important;
  }

 .json-boolean {
    color: #569cd6 !important;
    font-weight: 500;
  }

  .json-null {
    color: #569cd6 !important;
    font-weight: 500;
  }

  .json-punctuation {
    color: #e6e6e6 !important;
  }

  /* Custom scrollbar for JSON container */
  .json-container::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .json-container::-webkit-scrollbar-track {
    background: #2d2d30;
  }

  .json-container::-webkit-scrollbar-thumb {
    background: #5a5a5a;
    border-radius: 4px;
  }

  .json-container::-webkit-scrollbar-thumb:hover {
    background: #7a7a7a;
  }

  .beautified-json::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .beautified-json::-webkit-scrollbar-track {
    background: #1e1e1e;
  }

  .beautified-json::-webkit-scrollbar-thumb {
    background: #5a5a5a;
    border-radius: 4px;
  }

  .beautified-json::-webkit-scrollbar-thumb:hover {
    background: #7a7a7a;
  }

  /* Responsive modal */
  @media (max-width: 768px) {
    .modal-content {
      width: 95vw;
      max-height: 95vh;
    }

    .modal-body {
      max-height: 60vh;
    }

    .beautified-json {
      font-size: 0.8rem;
      padding: 1rem;
    }
  }

  /* Notification styles */
  .notification {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 10000;
    padding: 1rem 1.5rem;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    display: flex;
    align-items: center;
    gap: 1rem;
    max-width: 400px;
    animation: slideIn 0.3s ease-out;
  }

  .notification-success {
    background: #d4edda;
    color: #155724;
    border: 1px solid #c3e6cb;
  }

  .notification-error {
    background: #f8d7da;
    color: #721c24;
    border: 1px solid #f5c6cb;
  }

  .notification-info {
    background: #d1ecf1;
    color: #0c5460;
    border: 1px solid #bee5eb;
  }

  .notification-message {
    flex: 1;
    font-size: 0.9rem;
    line-height: 1.4;
  }

  .notification-close {
    background: none;
    border: none;
    color: inherit;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.7;
    transition: opacity 0.2s;
  }

  .notification-close:hover {
    opacity: 1;
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  @media (max-width: 768px) {
    .notification {
      top: 10px;
      right: 10px;
      left: 10px;
      max-width: none;
    }
  }
</style>
