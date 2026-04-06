<script lang="ts">
    import { onMount } from "svelte";
    import Login from "./lib/Login.svelte";
    import Setup from "./lib/Setup.svelte";
    import Shell from "./lib/Shell.svelte";
    import Admin from "./lib/Admin.svelte";
    import Library from "./lib/Library.svelte";
    import DBPicker from "./lib/DBPicker.svelte";
    import Settings from "./lib/Settings.svelte";
    import DataEntry from "./lib/DataEntry.svelte";
    import ChartReview from "./lib/ChartReview.svelte";

    import { NeedsSetup } from "../wailsjs/go/handlers/SetupHandler";
    import { GetDBPath, InitDB } from "../wailsjs/go/handlers/ConfigHandler";

    let view = "loading";
    let currentUser: any = null;

    onMount(async () => {
        try {
            const dbPath = await GetDBPath();
            if (!dbPath) {
                view = "db-pick";
                return;
            }
            await InitDB();
            const needsSetup = await NeedsSetup();
            view = needsSetup ? "setup" : "login";
        } catch (e) {
            view = "db-pick";
        }
    });

    async function handleDBReady() {
        try {
            const needsSetup = await NeedsSetup();
            view = needsSetup ? "setup" : "login";
        } catch {
            view = "login";
        }
    }

    function handleLogin(event: CustomEvent) {
        currentUser = event.detail;
        view =
            currentUser.role === "technician" ? "data-entry" : "chart-review";
    }

    function handleLogout() {
        currentUser = null;
        view = "login";
    }

    function handleNavigate(event: CustomEvent) {
        view = event.detail;
    }
</script>

{#if view === "loading"}
    <div class="loading">Loading…</div>
{:else if view === "db-pick"}
    <DBPicker on:ready={handleDBReady} />
{:else if view === "setup"}
    <Setup on:done={() => (view = "login")} />
{:else if view === "login"}
    <Login on:success={handleLogin} />
{:else}
    <Shell
        {currentUser}
        {view}
        on:logout={handleLogout}
        on:navigate={handleNavigate}
    >
        {#if view === "data-entry"}
            <DataEntry {currentUser} />
        {:else if view === "chart-review"}
            <ChartReview {currentUser} />
        {:else if view === "library"}
            <Library {currentUser} />
        {:else if view === "admin"}
            <Admin {currentUser} />
        {:else if view === "settings"}
            <Settings on:logout={handleLogout} />
        {/if}
    </Shell>
{/if}

<style>
    .loading {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--colour-text-muted);
    }
</style>
