<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Logout } from "../../wailsjs/go/handlers/AuthHandler";

    export let currentUser: any;
    export let view: string;

    const dispatch = createEventDispatcher();

    async function handleLogout() {
        await Logout();
        dispatch("logout");
    }

    const navItems = [
        {
            label: "Data Entry",
            view: "data-entry",
            roles: ["technician", "supervisor", "admin"],
        },
        {
            label: "Chart Review",
            view: "chart-review",
            roles: ["reviewer", "supervisor", "admin"],
        },
        { label: "Admin", view: "admin", roles: ["admin"] },
    ];

    $: visibleItems = navItems.filter((item) =>
        item.roles.includes(currentUser?.role),
    );
</script>

<div class="shell">
    <nav class="navbar">
        <div class="navbar-left">
            <span class="app-name">Cromulent</span>
            <div class="nav-links">
                {#each visibleItems as item}
                    <button
                        class="nav-link"
                        class:active={view === item.view}
                        on:click={() => dispatch("navigate", item.view)}
                    >
                        {item.label}
                    </button>
                {/each}
            </div>
        </div>
        <div class="navbar-right">
            <span class="user-info">
                <span class="username">{currentUser?.username}</span>
                <span class="role">{currentUser?.role}</span>
            </span>
            <button class="logout-btn" on:click={handleLogout}>Sign out</button>
        </div>
    </nav>

    <main class="content">
        <slot />
    </main>
</div>

<style>
    .shell {
        height: 100%;
        display: flex;
        flex-direction: column;
    }

    .navbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 1.5rem;
        height: 56px;
        background: var(--colour-surface);
        border-bottom: 1px solid var(--colour-border);
        box-shadow: var(--shadow-sm);
        flex-shrink: 0;
    }

    .navbar-left {
        display: flex;
        align-items: center;
        gap: 2rem;
    }

    .navbar-right {
        display: flex;
        align-items: center;
        gap: 1rem;
    }

    .app-name {
        font-weight: 700;
        font-size: 1.1rem;
        color: var(--colour-primary);
        letter-spacing: -0.01em;
    }

    .nav-links {
        display: flex;
        gap: 0.25rem;
    }

    .nav-link {
        background: none;
        border: none;
        padding: 0.375rem 0.75rem;
        border-radius: var(--radius);
        font-size: 0.9rem;
        color: var(--colour-text-muted);
        transition:
            background 0.15s,
            color 0.15s;
    }

    .nav-link:hover {
        background: var(--colour-bg);
        color: var(--colour-text);
    }

    .nav-link.active {
        background: var(--colour-bg);
        color: var(--colour-primary);
        font-weight: 700;
    }

    .user-info {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        line-height: 1.2;
    }

    .username {
        font-size: 0.875rem;
        font-weight: 700;
    }

    .role {
        font-size: 0.75rem;
        color: var(--colour-text-muted);
        text-transform: capitalize;
    }

    .logout-btn {
        background: none;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.375rem 0.75rem;
        font-size: 0.875rem;
        color: var(--colour-text-muted);
        transition:
            border-color 0.15s,
            color 0.15s;
    }

    .logout-btn:hover {
        border-color: var(--colour-danger);
        color: var(--colour-danger);
    }

    .content {
        flex: 1;
        overflow-y: auto;
        padding: 2rem;
    }
</style>
