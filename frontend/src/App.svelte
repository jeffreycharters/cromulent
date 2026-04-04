<script lang="ts">
    import { onMount } from "svelte";
    import Login from "./lib/Login.svelte";
    import Setup from "./lib/Setup.svelte";
    import Shell from "./lib/Shell.svelte";
    import Admin from "./lib/Admin.svelte";

    import { NeedsSetup } from "../wailsjs/go/handlers/SetupHandler";

    let view = "loading";
    let currentUser: any = null;

    onMount(async () => {
        const needsSetup = await NeedsSetup();
        view = needsSetup ? "setup" : "login";
    });

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
            <p>Data entry — todo</p>
        {:else if view === "chart-review"}
            <p>Chart review — todo</p>
        {:else if view === "admin"}
            <Admin {currentUser} />
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
