<script lang="ts">
    import { onMount } from "svelte";
    import {
        ListMethodsWithMaterials,
        GetAnalytesForCombo,
        SaveChart,
        GetChartResults,
    } from "../../wailsjs/go/handlers/DataEntryHandler";

    import type { models } from "../../wailsjs/go/models";

    import { AddComment } from "../../wailsjs/go/handlers/CommentsHandler";

    export let currentUser: any;

    type Method = models.MethodWithMaterials;

    interface ComboAnalyte {
        mma_id: number;
        name: string;
        unit: string;
        display_order: number;
    }

    interface MeasurementResult {
        mma_id: number;
        value: number;
        ucl: number | null;
        lcl: number | null;
        pass: boolean;
        no_limits: boolean;
    }

    let methods: Method[] = [];
    let selectedMethodMaterialID: number | null = null;
    let selectedMethodName = "";
    let selectedMaterialName = "";
    let analytes: ComboAnalyte[] = [];
    let values: Record<number, string> = {};
    let results: Record<string, MeasurementResult> = {};
    let error = "";
    let saving = false;
    let chartID: number | null = null;
    let commentText = "";
    let savingComment = false;

    onMount(async () => {
        try {
            methods = (await ListMethodsWithMaterials()) ?? [];
        } catch (e: any) {
            error = e.toString();
        }
    });

    async function selectCombo(
        methodMaterialID: number,
        methodName: string,
        materialName: string
    ) {
        if (selectedMethodMaterialID === methodMaterialID) return;
        selectedMethodMaterialID = methodMaterialID;
        values = {};
        results = {};
        try {
            analytes = (await GetAnalytesForCombo(methodMaterialID)) ?? [];
        } catch (e: any) {
            error = e.toString();
        }
        chartID = null;
        commentText = "";
        selectedMethodName = methodName;
        selectedMaterialName = materialName;
    }

    function handleCellPaste(e: ClipboardEvent, fromIndex: number) {
        e.preventDefault();

        let parts: string[] = [];

        const html = e.clipboardData?.getData("text/html") ?? "";
        if (html) {
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, "text/html");
            const cells = doc.querySelectorAll("td, th");
            if (cells.length > 0) {
                parts = Array.from(cells).map(
                    (td) => td.textContent?.trim() ?? ""
                );
            }
        }

        if (parts.length === 0) {
            const text = e.clipboardData?.getData("text/plain") ?? "";
            const firstRow = text.split(/\r?\n/)[0];
            parts = firstRow.trim().split("\t");
        }

        const next = { ...values };
        parts.forEach((part, offset) => {
            const analyte = analytes[fromIndex + offset];
            if (!analyte) return;
            const cleaned = part.trim();
            if (/^-?\d+(\.\d+)?$/.test(cleaned)) {
                next[analyte.mma_id] = cleaned;
            }
        });
        values = next;
    }

    async function save() {
        if (!selectedMethodMaterialID) return;
        const hasValues = Object.keys(values).length > 0;
        if (!hasValues) {
            error = "No values to save.";
            return;
        }
        saving = true;
        error = "";
        results = {};
        try {
            const payload: Record<string, number> = {};
            for (const [k, v] of Object.entries(values)) {
                if (v !== "") payload[k] = parseFloat(v);
            }
            chartID = await SaveChart(
                selectedMethodMaterialID,
                currentUser.id,
                payload
            );
            const raw = (await GetChartResults(chartID)) ?? [];
            results = Object.fromEntries(
                raw.map((r) => [r.mma_id, r])
            ) as Record<string, MeasurementResult>;
            values = {};
        } catch (e: any) {
            error = e.toString();
        } finally {
            saving = false;
        }
    }

    async function saveComment() {
        if (!chartID || !commentText.trim()) return;
        savingComment = true;
        try {
            await AddComment(chartID, null, commentText.trim(), currentUser.id);
            commentText = "";
        } catch (e: any) {
            error = e.toString();
        } finally {
            savingComment = false;
        }
    }

    $: hasValues = Object.values(values).some((v) => v !== "");
</script>

<div class="data-entry">
    <aside class="sidebar">
        {#each methods as method}
            <div class="method-group">
                <div class="method-label">{method.name}</div>
                {#each method.materials as material}
                    <button
                        class="material-btn"
                        class:active={selectedMethodMaterialID ===
                            material.method_material_id}
                        on:click={() =>
                            selectCombo(
                                material.method_material_id,
                                method.name,
                                material.name
                            )}
                    >
                        {material.name}
                    </button>
                {/each}
            </div>
        {:else}
            <p class="empty">
                No active combinations.<br />Set them up in Library.
            </p>
        {/each}
    </aside>

    <main class="main">
        {#if error}
            <div class="banner error">
                {error}
                <button on:click={() => (error = "")}>✕</button>
            </div>
        {/if}

        {#if !selectedMethodMaterialID}
            <div class="empty-state">
                <p>Select a method and material from the sidebar to begin.</p>
            </div>
        {:else if analytes.length === 0}
            <div class="empty-state">
                <p>No analytes configured for this combination.</p>
            </div>
        {:else}
            <div class="combo-header">
                <div>
                    <h2>{selectedMethodName}</h2>
                    <p class="subtitle">{selectedMaterialName}</p>
                </div>
                <button
                    class="save-btn"
                    disabled={!hasValues || saving}
                    on:click={save}
                >
                    {saving ? "Saving…" : "Save Chart"}
                </button>
            </div>

            <div class="paste-hint">
                Enter values manually, or paste tab-delimited instrument output
                into any cell to fill from that position forward. Non-numeric
                values will be left blank.
            </div>

            <div class="grid">
                {#each analytes as a, i}
                    {@const result = results[a.mma_id]}
                    <div
                        class="analyte-card"
                        class:pass={result && !result.no_limits && result.pass}
                        class:fail={result && !result.no_limits && !result.pass}
                    >
                        <div class="card-header">
                            <span class="analyte-name">{a.name}</span>
                            <span class="analyte-unit">{a.unit}</span>
                        </div>
                        {#if result}
                            <span class="result-value">{result.value}</span>
                            {#if result.no_limits}
                                <span class="result-badge no-limits"
                                    >No limits</span
                                >
                            {:else if result.pass}
                                <span class="result-badge pass-badge">Pass</span
                                >
                            {:else}
                                <span class="result-badge fail-badge">Fail</span
                                >
                            {/if}
                        {:else}
                            <input
                                type="text"
                                bind:value={values[a.mma_id]}
                                placeholder="—"
                                class="value-input"
                                class:filled={values[a.mma_id] !== undefined &&
                                    values[a.mma_id] !== ""}
                                on:paste={(e) => handleCellPaste(e, i)}
                            />
                        {/if}
                    </div>
                {/each}
            </div>
            {#if chartID}
                <div class="comment-section">
                    <textarea
                        class="comment-input"
                        bind:value={commentText}
                        placeholder="Add a run comment (optional)…"
                        rows="2"
                    />
                    <button
                        class="save-btn"
                        disabled={!commentText.trim() || savingComment}
                        on:click={saveComment}
                    >
                        {savingComment ? "Saving…" : "Add Comment"}
                    </button>
                </div>
            {/if}
        {/if}
    </main>
</div>

<style>
    .data-entry {
        display: flex;
        height: 100%;
        gap: 0;
        margin: -2rem;
    }
    .sidebar {
        width: 220px;
        flex-shrink: 0;
        border-right: 1px solid var(--colour-border);
        padding: 1.25rem 0.75rem;
        overflow-y: auto;
        background: var(--colour-surface);
    }
    .method-group {
        margin-bottom: 1.25rem;
    }
    .method-label {
        font-size: 0.75rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.06em;
        color: var(--colour-text-muted);
        padding: 0 0.5rem;
        margin-bottom: 0.25rem;
    }
    .material-btn {
        display: block;
        width: 100%;
        text-align: left;
        background: none;
        border: none;
        border-radius: var(--radius);
        padding: 0.4rem 0.5rem;
        font-size: 0.875rem;
        color: var(--colour-text);
        cursor: pointer;
        transition:
            background 0.15s,
            color 0.15s;
    }
    .material-btn:hover {
        background: var(--colour-bg);
    }
    .material-btn.active {
        background: color-mix(in srgb, var(--colour-primary) 12%, transparent);
        color: var(--colour-primary);
        font-weight: 600;
    }
    .main {
        flex: 1;
        padding: 2rem;
        overflow-y: auto;
        overflow-x: auto;
    }
    .combo-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 1.5rem;
    }
    .combo-header h2 {
        font-size: 1.25rem;
        font-weight: 700;
        margin-bottom: 0.125rem;
    }
    .subtitle {
        font-size: 0.875rem;
        color: var(--colour-text-muted);
    }
    .save-btn {
        background: var(--colour-primary);
        color: white;
        border: none;
        border-radius: var(--radius);
        padding: 0.5rem 1.25rem;
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
        transition: background 0.15s;
    }
    .save-btn:hover:not(:disabled) {
        background: var(--colour-primary-hover);
    }
    .save-btn:disabled {
        opacity: 0.4;
        cursor: default;
    }
    .paste-hint {
        font-size: 0.8rem;
        color: var(--colour-text-muted);
        margin-bottom: 0.75rem;
    }
    .grid {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }
    .analyte-card {
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.5rem 0.75rem;
        width: 140px;
        background: var(--colour-surface);
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        transition:
            border-color 0.15s,
            background 0.15s;
    }
    .analyte-card.pass {
        border-color: var(--colour-success);
        background: color-mix(in srgb, var(--colour-success) 6%, transparent);
    }
    .analyte-card.fail {
        border-color: var(--colour-danger);
        background: color-mix(in srgb, var(--colour-danger) 6%, transparent);
    }
    .card-header {
        display: flex;
        flex-direction: column;
        gap: 0.125rem;
    }
    .analyte-name {
        font-size: 0.8rem;
        font-weight: 600;
    }
    .analyte-unit {
        font-size: 0.7rem;
        color: var(--colour-text-muted);
    }
    .value-input {
        width: 100%;
        border: none;
        background: none;
        font-family: var(--font-mono);
        font-size: 0.9rem;
        color: var(--colour-text);
        outline: none;
        padding: 0.25rem 0;
    }
    .value-input.filled {
        color: var(--colour-primary);
    }
    .value-input::placeholder {
        color: var(--colour-border);
    }
    .result-value {
        font-family: var(--font-mono);
        font-size: 0.9rem;
        font-weight: 600;
        padding: 0.25rem 0;
    }
    .result-badge {
        font-size: 0.7rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        padding: 0.1rem 0;
    }
    .pass-badge {
        color: var(--colour-success);
    }
    .fail-badge {
        color: var(--colour-danger);
    }
    .no-limits {
        color: var(--colour-text-muted);
    }
    .empty-state {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--colour-text-muted);
        font-size: 0.9rem;
    }
    .banner {
        padding: 0.625rem 1rem;
        border-radius: var(--radius);
        margin-bottom: 1rem;
        font-size: 0.875rem;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .banner.error {
        background: color-mix(in srgb, var(--colour-danger) 12%, transparent);
        color: var(--colour-danger);
        border: 1px solid
            color-mix(in srgb, var(--colour-danger) 30%, transparent);
    }
    .empty {
        font-size: 0.8rem;
        color: var(--colour-text-muted);
        text-align: center;
        padding: 1rem 0.5rem;
        line-height: 1.6;
    }
    .comment-section {
        margin-top: 1.5rem;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        max-width: 480px;
    }
    .comment-input {
        width: 100%;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.5rem 0.75rem;
        font-family: var(--font-sans);
        font-size: 0.875rem;
        color: var(--colour-text);
        background: var(--colour-surface);
        resize: vertical;
    }
    .comment-input:focus {
        outline: none;
        border-color: var(--colour-primary);
    }
</style>
