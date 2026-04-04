<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import {
        OpenDBFilePicker,
        OpenDBFolderPicker,
        SetDBPath,
    } from "../../wailsjs/go/handlers/ConfigHandler";

    const dispatch = createEventDispatcher();

    let pickedPath = "";
    let error = "";
    let busy = false;

    async function openExisting() {
        const path = await OpenDBFilePicker();
        if (path) pickedPath = path;
    }

    async function createNew() {
        const path = await OpenDBFolderPicker();
        if (path) pickedPath = path;
    }

    async function confirm() {
        if (!pickedPath) return;
        busy = true;
        error = "";
        try {
            await SetDBPath(pickedPath);
            dispatch("ready");
        } catch (e: any) {
            error = e.toString();
        } finally {
            busy = false;
        }
    }
</script>

<div class="picker-wrap">
    <div class="picker-card">
        <h1>Cromulent</h1>
        <p class="subtitle">Open an existing database or create a new one.</p>

        <div class="actions">
            <button on:click={openExisting} disabled={busy} type="button"
                >Open existing…</button
            >
            <button on:click={createNew} disabled={busy} type="button"
                >Create new…</button
            >
        </div>

        {#if pickedPath}
            <p class="picked">{pickedPath}</p>
        {/if}

        {#if error}
            <p class="error">{error}</p>
        {/if}

        <button
            class="primary"
            on:click={confirm}
            disabled={!pickedPath || busy}
        >
            {busy ? "Opening…" : "Open"}
        </button>
    </div>
</div>

<style>
    .picker-wrap {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--colour-bg);
    }

    .picker-card {
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius-lg);
        box-shadow: var(--shadow-md);
        padding: 2.5rem;
        width: 480px;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    h1 {
        margin: 0;
        font-size: 1.5rem;
        color: var(--colour-primary);
    }

    .subtitle {
        margin: 0;
        color: var(--colour-text-muted);
        font-size: 0.9rem;
    }

    .actions {
        display: flex;
        gap: 0.5rem;
    }

    .picked {
        font-family: var(--font-mono);
        font-size: 0.8rem;
        color: var(--colour-text-muted);
        margin: 0;
        word-break: break-all;
    }

    .error {
        color: var(--colour-danger);
        font-size: 0.875rem;
        margin: 0;
    }

    .primary {
        align-self: flex-end;
    }

    button {
        padding: 0.25rem 0.5rem;
    }
</style>
