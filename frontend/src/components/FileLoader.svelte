<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { OpenFile, LoadJSONLFile, LoadJSONLFromClipboard } from '../../wailsjs/go/main/App.js';
  import { actions } from '../stores';
  import type { JSONLFile } from '../types';

  const dispatch = createEventDispatcher<{
    fileLoaded: JSONLFile;
    error: string;
  }>();

  let isLoadingFile = false;
  let isLoadingClipboard = false;

  // Handle file selection and loading
  async function handleFileSelect() {
    if (isLoadingFile) return;
    
    try {
      isLoadingFile = true;
      actions.clearError();
      actions.setLoading(true);

      // Open file dialog
      const filePath = await OpenFile();
      
      if (!filePath) {
        // User cancelled the dialog
        return;
      }

      // Load the selected file
      const jsonlFile = await LoadJSONLFile(filePath);
      
      // Update stores and dispatch event
      actions.setFile(jsonlFile);
      dispatch('fileLoaded', jsonlFile);
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load file';
      actions.setError(errorMessage);
      dispatch('error', errorMessage);
    } finally {
      isLoadingFile = false;
      actions.setLoading(false);
    }
  }

  // Handle clipboard paste
  async function handleClipboardPaste() {
    if (isLoadingClipboard) return;
    
    try {
      isLoadingClipboard = true;
      actions.clearError();
      actions.setLoading(true);

      // Load from clipboard
      const jsonlFile = await LoadJSONLFromClipboard();
      
      // Update stores and dispatch event
      actions.setFile(jsonlFile);
      dispatch('fileLoaded', jsonlFile);
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load from clipboard';
      actions.setError(errorMessage);
      dispatch('error', errorMessage);
    } finally {
      isLoadingClipboard = false;
      actions.setLoading(false);
    }
  }


</script>

<div class="file-loader">
  <div class="file-loader-actions">
    <!-- File Selection Button -->
    <button 
      class="file-button primary"
      on:click={handleFileSelect}
      disabled={isLoadingFile || isLoadingClipboard}
      title="Select a JSONL file from your computer"
    >
      {#if isLoadingFile}
        <div class="button-spinner"></div>
        <span>Opening...</span>
      {:else}
        <svg class="button-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14,2 14,8 20,8"/>
          <line x1="16" y1="13" x2="8" y2="13"/>
          <line x1="16" y1="17" x2="8" y2="17"/>
          <line x1="10" y1="9" x2="8" y2="9"/>
        </svg>
        <span>Open File</span>
      {/if}
    </button>

    <!-- Clipboard Paste Button -->
    <button 
      class="clipboard-button secondary"
      on:click={handleClipboardPaste}
      disabled={isLoadingFile || isLoadingClipboard}
      title="Load JSONL data from clipboard"
    >
      {#if isLoadingClipboard}
        <div class="button-spinner"></div>
        <span>Loading...</span>
      {:else}
        <svg class="button-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="8" y="2" width="8" height="4" rx="1" ry="1"/>
          <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2h2"/>
          <line x1="9" y1="12" x2="15" y2="12"/>
          <line x1="9" y1="16" x2="15" y2="16"/>
        </svg>
        <span>Paste from Clipboard</span>
      {/if}
    </button>
  </div>
</div>

<style>
  .file-loader {
    display: flex;
    justify-content: center;
  }

  .file-loader-actions {
    display: flex;
    gap: 0.75rem;
    align-items: center;
  }

  .file-button,
  .clipboard-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 140px;
    justify-content: center;
  }

  .file-button:disabled,
  .clipboard-button:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  .primary {
    background: #007bff;
    color: white;
  }

  .primary:hover:not(:disabled) {
    background: #0056b3;
    transform: translateY(-1px);
  }

  .primary:active:not(:disabled) {
    transform: translateY(0);
  }

  .secondary {
    background: #6c757d;
    color: white;
  }

  .secondary:hover:not(:disabled) {
    background: #545b62;
    transform: translateY(-1px);
  }

  .secondary:active:not(:disabled) {
    transform: translateY(0);
  }

  .button-icon {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
  }

  .button-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    flex-shrink: 0;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }





  /* Responsive design */
  @media (max-width: 768px) {
    .file-loader-actions {
      flex-direction: column;
      gap: 0.5rem;
    }

    .file-button,
    .clipboard-button {
      width: 100%;
      max-width: 200px;
    }
  }
</style>