<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import {
        OpenDBFilePicker,
        SetDBPath,
        GetDBPath,
    } from "../../wailsjs/go/handlers/ConfigHandler";
    import { Logout } from "../../wailsjs/go/handlers/AuthHandler";

    const dispatch = createEventDispatcher();

    let currentPath = "";
    let pickedPath = "";
    let error = "";
    let busy = false;
    let showWarning = false;

    import { onMount } from "svelte";
    onMount(async () => {
        currentPath = await GetDBPath();
    });

    async function browse() {
        const path = await OpenDBFilePicker();
        if (path && path !== currentPath) {
            pickedPath = path;
            showWarning = true;
        }
    }

    async function confirmChange() {
        busy = true;
        error = "";
        try {
            await Logout();
            await SetDBPath(pickedPath);
            dispatch("logout"); // App.svelte will re-run the startup flow
        } catch (e: any) {
            error = e.toString();
        } finally {
            busy = false;
        }
    }

    function cancel() {
        pickedPath = "";
        showWarning = false;
    }
</script>

<div class="settings">
    <h2>Settings</h2>

    <section>
        <h3>Database</h3>
        <p class="label">Current database file</p>
        <p class="path">{currentPath || "—"}</p>

        <button on:click={browse} disabled={busy}>Change database…</button>

        {#if showWarning}
            <div class="warning">
                <p>
                    <strong>Changing the database will sign you out.</strong>
                    Different databases have separate user accounts.
                </p>
                <p class="new-path">{pickedPath}</p>
                <div class="warning-actions">
                    <button on:click={cancel} disabled={busy}>Cancel</button>
                    <button
                        class="danger"
                        on:click={confirmChange}
                        disabled={busy}
                    >
                        {busy ? "Switching…" : "Switch and sign out"}
                    </button>
                </div>
            </div>
        {/if}

        {#if error}
            <p class="error">{error}</p>
        {/if}
    </section>
</div>

<style>
    .settings {
        max-width: 600px;
    }

    h2 {
        margin-top: 0;
    }

    h3 {
        font-size: 1rem;
        margin-bottom: 0.5rem;
        border-bottom: 1px solid var(--colour-border);
        padding-bottom: 0.5rem;
    }

    .label {
        font-size: 0.8rem;
        color: var(--colour-text-muted);
        margin: 0 0 0.25rem;
    }

    .path {
        font-family: var(--font-mono);
        font-size: 0.85rem;
        margin: 0 0 1rem;
    }

    .warning {
        margin-top: 1rem;
        padding: 1rem;
        border: 1px solid var(--colour-danger);
        border-radius: var(--radius);
        font-size: 0.9rem;
    }

    .new-path {
        font-family: var(--font-mono);
        font-size: 0.8rem;
        color: var(--colour-text-muted);
        margin: 0.5rem 0;
    }

    .warning-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
        margin-top: 0.75rem;
    }

    .error {
        color: var(--colour-danger);
        font-size: 0.875rem;
        margin-top: 0.5rem;
    }

    .danger {
        background: var(--colour-danger);
        color: white;
        border-color: var(--colour-danger);
    }
</style>
