<script lang="ts">
    import { onMount } from "svelte";
    import {
        ListUsers,
        DeactivateUser,
        ActivateUser,
        CreateUser,
    } from "../../wailsjs/go/handlers/AuthHandler";

    export let currentUser: any;

    let users: any[] = [];
    let error = "";
    let loading = true;

    // New user form
    let newUsername = "";
    let newPassword = "";
    let newRole = "technician";
    let createError = "";
    let creating = false;

    const roles = ["technician", "reviewer", "supervisor", "admin"];

    onMount(fetchUsers);

    async function fetchUsers() {
        loading = true;
        try {
            users = await ListUsers();
        } catch (e: any) {
            error = e;
        } finally {
            loading = false;
        }
    }

    async function toggleActive(user: any) {
        try {
            if (user.active) {
                await DeactivateUser(user.id);
            } else {
                await ActivateUser(user.id);
            }
            await fetchUsers();
        } catch (e: any) {
            error = e;
        }
    }

    async function handleCreate() {
        createError = "";
        if (!newUsername || !newPassword) {
            createError = "Username and password are required";
            return;
        }
        creating = true;
        try {
            await CreateUser(newUsername, newPassword, newRole);
            newUsername = "";
            newPassword = "";
            newRole = "technician";
            await fetchUsers();
        } catch (e: any) {
            createError = e;
        } finally {
            creating = false;
        }
    }
</script>

<div class="admin">
    <h2>User Management</h2>

    <section class="card">
        <h3>Create User</h3>
        <div class="create-form">
            <div class="field">
                <label for="new-username">Username</label>
                <input
                    id="new-username"
                    type="text"
                    bind:value={newUsername}
                    disabled={creating}
                    autocomplete="off"
                />
            </div>
            <div class="field">
                <label for="new-password">Password</label>
                <input
                    id="new-password"
                    type="password"
                    bind:value={newPassword}
                    disabled={creating}
                />
            </div>
            <div class="field">
                <label for="new-role">Role</label>
                <select id="new-role" bind:value={newRole} disabled={creating}>
                    {#each roles as role}
                        <option value={role}>{role}</option>
                    {/each}
                </select>
            </div>
            <button
                on:click={handleCreate}
                disabled={creating || !newUsername || !newPassword}
            >
                {creating ? "Creating…" : "Create user"}
            </button>
        </div>
        {#if createError}
            <p class="error">{createError}</p>
        {/if}
    </section>

    <section class="card">
        <h3>Users</h3>
        {#if loading}
            <p class="muted">Loading…</p>
        {:else if error}
            <p class="error">{error}</p>
        {:else}
            <table>
                <thead>
                    <tr>
                        <th>Username</th>
                        <th>Role</th>
                        <th>Created</th>
                        <th>Status</th>
                        <th></th>
                    </tr>
                </thead>
                <tbody>
                    {#each users as user}
                        <tr class:inactive={!user.active}>
                            <td>{user.username}</td>
                            <td class="role">{user.role}</td>
                            <td class="muted mono"
                                >{user.created_at.slice(0, 10)}</td
                            >
                            <td>
                                <span
                                    class="badge"
                                    class:badge-active={user.active}
                                    class:badge-inactive={!user.active}
                                >
                                    {user.active ? "Active" : "Inactive"}
                                </span>
                            </td>
                            <td>
                                {#if user.id !== currentUser.id}
                                    <button
                                        class="toggle-btn"
                                        class:danger={user.active}
                                        on:click={() => toggleActive(user)}
                                    >
                                        {user.active
                                            ? "Deactivate"
                                            : "Activate"}
                                    </button>
                                {:else}
                                    <span class="muted">—</span>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        {/if}
    </section>
</div>

<style>
    .admin {
        max-width: 800px;
    }

    h2 {
        font-size: 1.25rem;
        font-weight: 700;
        margin-bottom: 1.5rem;
    }

    h3 {
        font-size: 1rem;
        font-weight: 700;
        margin-bottom: 1rem;
    }

    .card {
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius-lg);
        padding: 1.5rem;
        margin-bottom: 1.5rem;
        box-shadow: var(--shadow-sm);
    }

    .create-form {
        display: flex;
        align-items: flex-end;
        gap: 1rem;
        flex-wrap: wrap;
    }

    .field {
        display: flex;
        flex-direction: column;
        gap: 0.375rem;
    }

    label {
        font-size: 0.875rem;
        font-weight: 700;
    }

    input,
    select {
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.5rem 0.75rem;
        font-size: 0.9rem;
        background: var(--colour-bg);
        color: var(--colour-text);
        transition: border-color 0.15s;
        height: 2.375rem;
    }

    input:focus,
    select:focus {
        outline: none;
        border-color: var(--colour-primary);
    }

    button {
        background: var(--colour-primary);
        color: white;
        border: none;
        border-radius: var(--radius);
        padding: 0.5rem 1rem;
        font-size: 0.9rem;
        font-weight: 700;
        transition: background 0.15s;
        white-space: nowrap;
    }

    button:hover:not(:disabled) {
        background: var(--colour-primary-hover);
    }

    button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.9rem;
    }

    th {
        text-align: left;
        font-size: 0.8rem;
        font-weight: 700;
        color: var(--colour-text-muted);
        padding: 0.5rem 0.75rem;
        border-bottom: 1px solid var(--colour-border);
    }

    td {
        padding: 0.625rem 0.75rem;
        border-bottom: 1px solid var(--colour-border);
    }

    tr:last-child td {
        border-bottom: none;
    }

    tr.inactive td {
        opacity: 0.5;
    }

    .role {
        text-transform: capitalize;
    }

    .mono {
        font-family: var(--font-mono);
    }

    .muted {
        color: var(--colour-text-muted);
        font-size: 0.875rem;
    }

    .badge {
        display: inline-block;
        padding: 0.2rem 0.5rem;
        border-radius: var(--radius);
        font-size: 0.75rem;
        font-weight: 700;
    }

    .badge-active {
        background: #dcfce7;
        color: var(--colour-success);
    }

    .badge-inactive {
        background: #fee2e2;
        color: var(--colour-danger);
    }

    .toggle-btn {
        background: none;
        border: 1px solid var(--colour-border);
        color: var(--colour-text-muted);
        font-weight: 400;
        padding: 0.25rem 0.625rem;
        font-size: 0.8rem;
    }

    .toggle-btn:hover:not(:disabled) {
        background: none;
        border-color: var(--colour-primary);
        color: var(--colour-primary);
    }

    .toggle-btn.danger:hover:not(:disabled) {
        border-color: var(--colour-danger);
        color: var(--colour-danger);
    }

    .error {
        color: var(--colour-danger);
        font-size: 0.875rem;
        margin-top: 0.75rem;
    }
</style>
