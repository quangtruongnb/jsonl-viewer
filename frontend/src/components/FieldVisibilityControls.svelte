<script lang="ts">
    import { onMount } from "svelte";
    import { GetAllFields } from "../../wailsjs/go/main/App.js";
    import {
        availableFields,
        fieldsToShow,
        fieldsToHide,
        currentFile,
        actions,
    } from "../stores";

    // Component state
    let fieldsDropdownOpen = false;
    let fieldsSearch = "";

    // Reactive statements for filtered fields
    $: filteredAvailableFields = $availableFields.filter(
        (field) =>
            field.toLowerCase().includes(fieldsSearch.toLowerCase()) &&
            !$fieldsToShow.includes(field) &&
            !$fieldsToHide.includes(field),
    );

    // Extract available fields from backend when file changes
    $: if ($currentFile) {
        extractAvailableFieldsFromBackend();
    }

    async function extractAvailableFieldsFromBackend() {
        try {
            const fields = await GetAllFields();
            actions.setAvailableFields(fields);
        } catch (error) {
            console.error('Failed to get fields from backend:', error);
            actions.setAvailableFields([]);
        }
    }

    function toggleField(field: string) {
        if ($fieldsToShow.includes(field)) {
            // If field is in show list, remove it (move to available)
            actions.removeFieldToShow(field);
        } else if ($fieldsToHide.includes(field)) {
            // If field is in hide list, remove it (move to available)
            actions.removeFieldToHide(field);
        } else {
            // If field is not in either list, add it to hide list
            actions.addFieldToHide(field);
        }
    }

    function showAllFields() {
        actions.setFieldsToShow($availableFields);
        actions.setFieldsToHide([]);
    }

    function hideAllFields() {
        actions.setFieldsToShow([]);
        actions.setFieldsToHide($availableFields);
    }

    function resetFieldVisibility() {
        actions.setFieldsToShow([]);
        actions.setFieldsToHide([]);
    }

    // Close dropdowns when clicking outside
    function handleClickOutside(event: MouseEvent) {
        const target = event.target as Element;
        // Don't close if clicking on a checkbox or field item
        if (
            !target.closest(".field-dropdown") &&
            !target.closest(".field-item")
        ) {
            fieldsDropdownOpen = false;
        }
    }

    onMount(() => {
        document.addEventListener("click", handleClickOutside);
        return () => {
            document.removeEventListener("click", handleClickOutside);
        };
    });
</script>

<div class="field-visibility-controls">
    <div class="controls-header">
        <h3>Field Visibility</h3>
        <div class="header-actions">
            <button
                class="action-button show-all"
                on:click={showAllFields}
                title="Show all fields"
            >
                Show All
            </button>
            <button
                class="action-button hide-all"
                on:click={hideAllFields}
                title="Hide all fields"
            >
                Hide All
            </button>
            <button
                class="reset-button"
                on:click={resetFieldVisibility}
                title="Reset field visibility"
            >
                Reset
            </button>
        </div>
    </div>

    <div class="controls-row">
        <!-- Unified Field Control -->
        <div class="field-control">
            <label class="control-label" for="fields-dropdown">Fields:</label>
            <div class="field-dropdown">
                <button
                    id="fields-dropdown"
                    class="dropdown-toggle"
                    class:active={fieldsDropdownOpen}
                    on:click={() =>
                        (fieldsDropdownOpen = !fieldsDropdownOpen)}
                    disabled={$availableFields.length === 0}
                >
                    <span class="dropdown-text">
                        {#if $fieldsToShow.length === 0 && $fieldsToHide.length === 0}
                            All fields ({$availableFields.length})
                        {:else if $fieldsToShow.length > 0}
                            {$fieldsToShow.length} shown
                        {:else if $fieldsToHide.length > 0}
                            {$fieldsToHide.length} hidden
                        {/if}
                    </span>
                    <svg
                        class="dropdown-arrow"
                        class:rotated={fieldsDropdownOpen}
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <polyline points="6,9 12,15 18,9"></polyline>
                    </svg>
                </button>

                {#if fieldsDropdownOpen}
                    <div
                        class="dropdown-menu"
                        on:click|stopPropagation
                        on:keydown|stopPropagation
                    >
                        <div class="dropdown-header">
                            <input
                                type="text"
                                class="field-search"
                                placeholder="Search fields..."
                                bind:value={fieldsSearch}
                            />
                        </div>

                        <div class="field-list">
                            {#if $fieldsToShow.length > 0}
                                <div class="selected-fields">
                                    <div class="section-header">
                                        Shown Fields:
                                    </div>
                                    <div class="field-tags">
                                        {#each $fieldsToShow as field}
                                            <button
                                                class="field-tag selected"
                                                on:click|stopPropagation={() =>
                                                    toggleField(field)}
                                                title="Remove from shown fields"
                                            >
                                                <svg class="remove-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                    <line x1="18" y1="6" x2="6" y2="18"></line>
                                                    <line x1="6" y1="6" x2="18" y2="18"></line>
                                                </svg>
                                                <span class="field-name">{field}</span>
                                            </button>
                                        {/each}
                                    </div>
                                </div>
                            {/if}

                            {#if $fieldsToHide.length > 0}
                                <div class="hidden-fields">
                                    <div class="section-header">
                                        Hidden Fields:
                                    </div>
                                    <div class="field-tags">
                                        {#each $fieldsToHide as field}
                                            <button
                                                class="field-tag hidden"
                                                on:click|stopPropagation={() =>
                                                    toggleField(field)}
                                                title="Remove from hidden fields"
                                            >
                                                <svg class="remove-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                    <line x1="18" y1="6" x2="6" y2="18"></line>
                                                    <line x1="6" y1="6" x2="18" y2="18"></line>
                                                </svg>
                                                <span class="field-name">{field}</span>
                                            </button>
                                        {/each}
                                    </div>
                                </div>
                            {/if}

                            {#if filteredAvailableFields.length > 0}
                                <div class="available-fields">
                                    {#if $fieldsToShow.length > 0 || $fieldsToHide.length > 0}
                                        <div class="section-header">
                                            Available Fields:
                                        </div>
                                    {/if}
                                    <div class="field-tags">
                                        {#each filteredAvailableFields as field}
                                            <button
                                                class="field-tag available"
                                                on:click|stopPropagation={() =>
                                                    toggleField(field)}
                                                title="Add to hidden fields"
                                            >
                                                <span class="field-name">{field}</span>
                                            </button>
                                        {/each}
                                    </div>
                                </div>
                            {/if}

                            {#if filteredAvailableFields.length === 0 && $fieldsToShow.length === 0 && $fieldsToHide.length === 0}
                                <div class="no-fields">No fields available</div>
                            {/if}
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    </div>

    <!-- Field visibility summary -->
    {#if $fieldsToShow.length > 0 || $fieldsToHide.length > 0}
        <div class="visibility-summary">
            {#if $fieldsToShow.length > 0}
                <span class="summary-item show">
                    Showing {$fieldsToShow.length} of {$availableFields.length} fields
                </span>
            {:else if $fieldsToHide.length > 0}
                <span class="summary-item hide">
                    Hiding {$fieldsToHide.length} of {$availableFields.length} fields
                </span>
            {/if}
        </div>
    {/if}
</div>

<style>
    .field-visibility-controls {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        padding: 1rem;
        background: #f8f9fa;
        border: 1px solid #e9ecef;
        border-radius: 6px;
    }

    .controls-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .controls-header h3 {
        margin: 0;
        font-size: 1rem;
        font-weight: 600;
        color: #495057;
    }

    .header-actions {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }

    .action-button {
        background: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        padding: 0.25rem 0.5rem;
        font-size: 0.75rem;
        cursor: pointer;
        transition: background 0.2s ease;
    }

    .action-button:hover {
        background: #0056b3;
    }

    .action-button.show-all {
        background: #28a745;
    }

    .action-button.show-all:hover {
        background: #218838;
    }

    .action-button.hide-all {
        background: #dc3545;
    }

    .action-button.hide-all:hover {
        background: #c82333;
    }

    .reset-button {
        background: #6c757d;
        color: white;
        border: none;
        border-radius: 4px;
        padding: 0.25rem 0.5rem;
        font-size: 0.75rem;
        cursor: pointer;
        transition: background 0.2s ease;
    }

    .reset-button:hover {
        background: #545b62;
    }

    .controls-row {
        display: flex;
        gap: 1rem;
        align-items: flex-start;
    }

    .field-control {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex: 1;
    }

    .control-label {
        font-size: 0.85rem;
        font-weight: 500;
        color: #495057;
        white-space: nowrap;
        flex-shrink: 0;
    }

    .field-dropdown {
        position: relative;
        flex: 1;
        min-width: 0;
    }

    .dropdown-toggle {
        width: 100%;
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.5rem;
        background: white;
        border: 2px solid #e9ecef;
        border-radius: 4px;
        font-size: 0.85rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .dropdown-toggle:hover:not(:disabled) {
        border-color: #007bff;
    }

    .dropdown-toggle.active {
        border-color: #007bff;
        box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
    }

    .dropdown-toggle:disabled {
        background: #f8f9fa;
        color: #6c757d;
        cursor: not-allowed;
    }

    .dropdown-text {
        color: #495057;
    }

    .dropdown-arrow {
        width: 16px;
        height: 16px;
        color: #6c757d;
        transition: transform 0.2s ease;
    }

    .dropdown-arrow.rotated {
        transform: rotate(180deg);
    }

    .dropdown-menu {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        z-index: 1000;
        background: white;
        border: 1px solid #e9ecef;
        border-radius: 4px;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        max-height: 300px;
        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    .dropdown-header {
        padding: 0.5rem;
        border-bottom: 1px solid #e9ecef;
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }

    .field-search {
        flex: 1;
        padding: 0.25rem 0.5rem;
        border: 1px solid #e9ecef;
        border-radius: 3px;
        font-size: 0.8rem;
    }

    .field-search:focus {
        outline: none;
        border-color: #007bff;
    }



    .field-list {
        overflow-y: auto;
        max-height: 200px;
    }

    .section-header {
        padding: 0.5rem;
        font-size: 0.75rem;
        font-weight: 600;
        color: #6c757d;
        background: #f8f9fa;
        border-bottom: 1px solid #e9ecef;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }



    .field-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 0.25rem;
        padding: 0.5rem;
        border: 1px dashed #dee2e6;
        border-radius: 4px;
        background: white;
        min-height: 2rem;
    }

    .field-tag {
        display: inline-flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.25rem 0.5rem;
        background: #2563eb;
        color: #ffffff;
        border: 1px solid #1d4ed8;
        border-radius: 12px;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s ease;
        white-space: nowrap;
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .field-tag:hover {
        background: #1d4ed8;
        transform: translateY(-1px);
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
    }

    .field-tag.selected {
        background: #059669;
        border-color: #047857;
        color: #ffffff;
    }

    .field-tag.selected:hover {
        background: #047857;
    }

    .field-tag.hidden {
        background: #dc2626;
        border-color: #b91c1c;
        color: #ffffff;
    }

    .field-tag.hidden:hover {
        background: #b91c1c;
    }

    .field-tag.available {
        background: #374151;
        border-color: #1f2937;
        color: #ffffff;
    }

    .field-tag.available:hover {
        background: #1f2937;
    }

    .remove-icon {
        width: 12px;
        height: 12px;
        flex-shrink: 0;
        color: #ffffff;
    }

    .field-tag .field-name {
        font-size: 0.75rem;
        font-weight: 500;
        color: #ffffff;
    }



    .field-name {
        flex: 1;
        color: #495057;
        font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
        font-size: 0.8rem;
    }

    .no-fields {
        padding: 1rem;
        text-align: center;
        color: #6c757d;
        font-size: 0.85rem;
        font-style: italic;
    }

    .visibility-summary {
        padding: 0.5rem;
        background: white;
        border: 1px solid #e9ecef;
        border-radius: 4px;
        font-size: 0.8rem;
    }

    .summary-item.show {
        color: #28a745;
        font-weight: 500;
    }

    .summary-item.hide {
        color: #dc3545;
        font-weight: 500;
    }

    /* Responsive design */
    @media (max-width: 768px) {
        .controls-row {
            flex-direction: column;
            gap: 0.75rem;
        }

        .field-control {
            flex-direction: column;
            align-items: flex-start;
            gap: 0.25rem;
        }

        .field-dropdown {
            width: 100%;
        }

        .dropdown-menu {
            max-height: 250px;
        }

        .field-list {
            max-height: 150px;
        }
    }
</style>
