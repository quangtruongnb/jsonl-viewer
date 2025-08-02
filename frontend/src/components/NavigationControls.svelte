<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    currentRecordIndex, 
    totalRecords, 
    canNavigatePrevious, 
    canNavigateNext,
    actions 
  } from '../stores';
  import { GetRecordByLineNumber } from '../../wailsjs/go/main/App.js';
  import type { JSONRecord } from '../types';

  const dispatch = createEventDispatcher<{
    recordSelected: JSONRecord;
    navigationChanged: { index: number };
  }>();

  let goToLineInput = '';
  let isGoingToLine = false;
  let goToLineError = '';

  // Handle previous record navigation
  async function handlePrevious() {
    if (!$canNavigatePrevious) return;
    
    actions.navigatePrevious();
    await loadCurrentRecord();
  }

  // Handle next record navigation
  async function handleNext() {
    if (!$canNavigateNext) return;
    
    actions.navigateNext();
    await loadCurrentRecord();
  }

  // Handle go to line functionality
  async function handleGoToLine() {
    const lineNumber = parseInt(goToLineInput.trim());
    
    if (isNaN(lineNumber) || lineNumber <= 0) {
      goToLineError = 'Please enter a valid line number';
      return;
    }

    try {
      isGoingToLine = true;
      goToLineError = '';
      
      // Try to get the record at the specified line number
      const record = await GetRecordByLineNumber(lineNumber);
      
      // Since JSONL files can have gaps in line numbers (due to empty lines or invalid JSON),
      // we need to find the actual index of this record in our records array.
      // For now, we'll use a simple approach where we assume the record index
      // corresponds to the order in which valid records appear.
      
      // The backend should return the record if it exists
      if (record) {
        // Find the record index by searching through records
        // This is a simplified approach - in practice, you might want to
        // maintain a mapping between line numbers and record indices
        const recordIndex = record.lineNumber - 1; // Simplified mapping
        
        if (recordIndex >= 0 && recordIndex < $totalRecords) {
          actions.setCurrentRecordIndex(recordIndex);
          dispatch('recordSelected', record);
          dispatch('navigationChanged', { index: recordIndex });
          goToLineInput = '';
        } else {
          goToLineError = `Line ${lineNumber} not found in current view`;
        }
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to go to line';
      if (errorMsg.includes('not found')) {
        goToLineError = `Line ${lineNumber} not found or contains invalid JSON`;
      } else {
        goToLineError = errorMsg;
      }
    } finally {
      isGoingToLine = false;
    }
  }

  // Load the current record and dispatch events
  async function loadCurrentRecord() {
    try {
      // For navigation, we'll dispatch the navigation change event
      // and let the parent component handle loading the actual record data
      dispatch('navigationChanged', { index: $currentRecordIndex });
    } catch (error) {
      console.error('Failed to load current record:', error);
    }
  }

  // Handle keyboard shortcuts
  function handleKeydown(event: KeyboardEvent) {
    if (event.target instanceof HTMLInputElement) {
      // Don't handle shortcuts when typing in input
      return;
    }

    switch (event.key) {
      case 'ArrowLeft':
      case 'ArrowUp':
        event.preventDefault();
        handlePrevious();
        break;
      case 'ArrowRight':
      case 'ArrowDown':
        event.preventDefault();
        handleNext();
        break;
      case 'g':
        if (event.ctrlKey || event.metaKey) {
          event.preventDefault();
          const goToInput = document.querySelector('.goto-input') as HTMLInputElement;
          goToInput?.focus();
        }
        break;
    }
  }

  // Handle Enter key in go-to-line input
  function handleGoToLineKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      handleGoToLine();
    } else if (event.key === 'Escape') {
      goToLineInput = '';
      goToLineError = '';
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if $totalRecords > 0}
  <div class="navigation-controls">
    <div class="nav-section">
      <div class="nav-buttons">
        <button
          class="nav-button"
          class:disabled={!$canNavigatePrevious}
          disabled={!$canNavigatePrevious}
          on:click={handlePrevious}
          title="Previous record (← or ↑)"
        >
          <span class="nav-icon">←</span>
          Previous
        </button>
        
        <div class="record-position">
          <span class="current-position">
            {$currentRecordIndex + 1}
          </span>
          <span class="position-separator">of</span>
          <span class="total-records">
            {$totalRecords.toLocaleString()}
          </span>
        </div>
        
        <button
          class="nav-button"
          class:disabled={!$canNavigateNext}
          disabled={!$canNavigateNext}
          on:click={handleNext}
          title="Next record (→ or ↓)"
        >
          Next
          <span class="nav-icon">→</span>
        </button>
      </div>
    </div>

    <div class="goto-section">
      <div class="goto-controls">
        <label for="goto-line" class="goto-label">Go to line:</label>
        <input
          id="goto-line"
          class="goto-input"
          type="number"
          min="1"
          bind:value={goToLineInput}
          on:keydown={handleGoToLineKeydown}
          placeholder="Line #"
          title="Enter line number and press Enter (Ctrl+G to focus)"
        />
        <button
          class="goto-button"
          class:loading={isGoingToLine}
          disabled={isGoingToLine || !goToLineInput.trim() || $totalRecords === 0}
          on:click={handleGoToLine}
          title="Jump to specified line"
        >
          {#if isGoingToLine}
            <span class="loading-spinner"></span>
          {:else}
            Go
          {/if}
        </button>
      </div>
      
      {#if goToLineError}
        <div class="goto-error">
          {goToLineError}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .navigation-controls {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: #f8f9fa;
    border-top: 1px solid #e9ecef;
    border-bottom: 1px solid #e9ecef;
    gap: 1rem;
  }

  .nav-section {
    flex: 1;
  }

  .nav-buttons {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .nav-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: white;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.15s ease;
    color: #495057;
  }

  .nav-button:hover:not(.disabled) {
    background: #007bff;
    color: white;
    border-color: #007bff;
  }

  .nav-button:active:not(.disabled) {
    transform: translateY(1px);
  }

  .nav-button.disabled {
    background: #f8f9fa;
    color: #6c757d;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .nav-icon {
    font-weight: bold;
    font-size: 1rem;
  }

  .record-position {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
    color: #495057;
    background: white;
    padding: 0.5rem 0.75rem;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    min-width: 120px;
    justify-content: center;
  }

  .current-position {
    font-weight: 600;
    color: #007bff;
  }

  .position-separator {
    color: #6c757d;
    font-size: 0.8rem;
  }

  .total-records {
    font-weight: 500;
  }

  .goto-section {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
  }

  .goto-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .goto-label {
    font-size: 0.9rem;
    color: #495057;
    font-weight: 500;
  }

  .goto-input {
    width: 80px;
    padding: 0.4rem 0.6rem;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    font-size: 0.9rem;
    text-align: center;
  }

  .goto-input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  .goto-button {
    padding: 0.4rem 0.8rem;
    background: #007bff;
    color: white;
    border: 1px solid #007bff;
    border-radius: 4px;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.15s ease;
    min-width: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .goto-button:hover:not(:disabled) {
    background: #0056b3;
    border-color: #0056b3;
  }

  .goto-button:disabled {
    background: #6c757d;
    border-color: #6c757d;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .goto-button.loading {
    cursor: wait;
  }

  .loading-spinner {
    width: 14px;
    height: 14px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .goto-error {
    font-size: 0.8rem;
    color: #dc3545;
    background: #f8d7da;
    border: 1px solid #f5c6cb;
    border-radius: 3px;
    padding: 0.25rem 0.5rem;
    max-width: 200px;
    text-align: center;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .navigation-controls {
      flex-direction: column;
      gap: 0.75rem;
      padding: 0.5rem;
    }

    .nav-buttons {
      justify-content: center;
    }

    .nav-button {
      padding: 0.4rem 0.8rem;
      font-size: 0.8rem;
    }

    .record-position {
      font-size: 0.8rem;
      padding: 0.4rem 0.6rem;
      min-width: 100px;
    }

    .goto-controls {
      gap: 0.4rem;
    }

    .goto-label {
      font-size: 0.8rem;
    }

    .goto-input {
      width: 70px;
      padding: 0.3rem 0.5rem;
      font-size: 0.8rem;
    }

    .goto-button {
      padding: 0.3rem 0.6rem;
      font-size: 0.8rem;
      min-width: 45px;
    }
  }

  /* High contrast mode support */
  @media (prefers-contrast: high) {
    .nav-button {
      border-width: 2px;
    }

    .nav-button:hover:not(.disabled) {
      border-width: 2px;
    }

    .goto-input:focus {
      border-width: 2px;
    }
  }

  /* Reduced motion support */
  @media (prefers-reduced-motion: reduce) {
    .nav-button,
    .goto-button,
    .loading-spinner {
      transition: none;
      animation: none;
    }
  }
</style>