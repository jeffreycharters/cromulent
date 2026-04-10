<script lang="ts">
    import { onMount } from "svelte";
    import {
        ListUsers,
        DeactivateUser,
        ActivateUser,
        CreateUser,
    } from "../../wailsjs/go/handlers/AuthHandler";
    import {
        GetGlobalRuleSet,
        UpdateGlobalRuleSet,
    } from "../../wailsjs/go/handlers/SPCRuleSetHandler";

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

    // SPC global rule set
    let ruleSet: any = null;
    let ruleSetError = "";
    let ruleSetSuccess = "";
    let savingRules = false;

    // Editable copies
    let beyondLimitsEnabled = true;
    let warningLimitsEnabled = true;
    let warningConsecutiveCount = 3;
    let warningTriggerCount = 2;
    let trendEnabled = true;
    let trendConsecutiveCount = 6;
    let oneSideEnabled = true;
    let oneSideConsecutiveCount = 8;

    onMount(async () => {
        await Promise.all([fetchUsers(), fetchRuleSet()]);
    });

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

    async function fetchRuleSet() {
        try {
            ruleSet = await GetGlobalRuleSet();
            populateRuleForm(ruleSet);
        } catch (e: any) {
            ruleSetError = e.toString();
        }
    }

    function populateRuleForm(rs: any) {
        beyondLimitsEnabled = rs.beyondLimitsEnabled;
        warningLimitsEnabled = rs.warningLimitsEnabled;
        warningConsecutiveCount = rs.warningConsecutiveCount;
        warningTriggerCount = rs.warningTriggerCount;
        trendEnabled = rs.trendEnabled;
        trendConsecutiveCount = rs.trendConsecutiveCount;
        oneSideEnabled = rs.oneSideEnabled;
        oneSideConsecutiveCount = rs.oneSideConsecutiveCount;
    }

    async function saveRuleSet() {
        savingRules = true;
        ruleSetError = "";
        ruleSetSuccess = "";
        try {
            await UpdateGlobalRuleSet(
                beyondLimitsEnabled,
                warningLimitsEnabled,
                warningConsecutiveCount,
                warningTriggerCount,
                trendEnabled,
                trendConsecutiveCount,
                oneSideEnabled,
                oneSideConsecutiveCount,
                currentUser.id
            );
            ruleSet = await GetGlobalRuleSet();
            populateRuleForm(ruleSet);
            ruleSetSuccess = "Rules saved.";
            setTimeout(() => (ruleSetSuccess = ""), 3000);
        } catch (e: any) {
            ruleSetError = e.toString();
        } finally {
            savingRules = false;
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
    {#if error}
        <p class="error">{error}</p>
    {/if}

    <!-- User Management -->
    <details open class="accordion">
        <summary class="accordion-summary">User Management</summary>

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
                    <select
                        id="new-role"
                        bind:value={newRole}
                        disabled={creating}
                    >
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
    </details>

    <!-- SPC Rules -->
    <details open class="accordion">
        <summary class="accordion-summary">SPC Rules</summary>

        <section class="card">
            <p class="rules-note">
                Global defaults — apply to all combos unless overridden in the
                Library.
            </p>

            {#if ruleSetError}
                <p class="error">{ruleSetError}</p>
            {/if}
            {#if ruleSetSuccess}
                <p class="success">{ruleSetSuccess}</p>
            {/if}

            {#if ruleSet}
                <div class="rule-groups">
                    <!-- Control limits group -->
                    <div class="rule-group">
                        <h4 class="rule-group-label">Control limits</h4>

                        <div class="rule-row">
                            <label class="rule-toggle">
                                <input
                                    type="checkbox"
                                    bind:checked={beyondLimitsEnabled}
                                />
                                Beyond control limits (OOC)
                            </label>
                        </div>

                        <div
                            class="rule-row"
                            class:disabled={!warningLimitsEnabled}
                        >
                            <label class="rule-toggle">
                                <input
                                    type="checkbox"
                                    bind:checked={warningLimitsEnabled}
                                />
                                Warning limits
                            </label>
                            <span class="rule-desc">
                                <input
                                    type="number"
                                    class="count-input"
                                    bind:value={warningTriggerCount}
                                    min="1"
                                    max="10"
                                    disabled={!warningLimitsEnabled}
                                />
                                of
                                <input
                                    type="number"
                                    class="count-input"
                                    bind:value={warningConsecutiveCount}
                                    min="1"
                                    max="10"
                                    disabled={!warningLimitsEnabled}
                                /> points outside warning limits
                            </span>
                        </div>
                    </div>

                    <!-- Run rules group -->
                    <div class="rule-group">
                        <h4 class="rule-group-label">Run rules</h4>

                        <div class="rule-row" class:disabled={!trendEnabled}>
                            <label class="rule-toggle">
                                <input
                                    type="checkbox"
                                    bind:checked={trendEnabled}
                                />
                                Trend
                            </label>
                            <span class="rule-desc">
                                <input
                                    type="number"
                                    class="count-input"
                                    bind:value={trendConsecutiveCount}
                                    min="2"
                                    max="20"
                                    disabled={!trendEnabled}
                                /> consecutive points trending in one direction
                            </span>
                        </div>

                        <div class="rule-row" class:disabled={!oneSideEnabled}>
                            <label class="rule-toggle">
                                <input
                                    type="checkbox"
                                    bind:checked={oneSideEnabled}
                                />
                                One side of mean
                            </label>
                            <span class="rule-desc">
                                <input
                                    type="number"
                                    class="count-input"
                                    bind:value={oneSideConsecutiveCount}
                                    min="2"
                                    max="20"
                                    disabled={!oneSideEnabled}
                                /> consecutive points on one side of mean
                            </span>
                        </div>
                    </div>
                </div>

                <div class="rules-footer">
                    <button
                        class="save-btn"
                        on:click={saveRuleSet}
                        disabled={savingRules}
                    >
                        {savingRules ? "Saving…" : "Save rules"}
                    </button>
                </div>
            {:else if !ruleSetError}
                <p class="muted">Loading…</p>
            {/if}
        </section>
    </details>
</div>

<style>
    .admin {
        max-width: 800px;
    }

    /* Accordion */
    .accordion {
        margin-bottom: 1.5rem;
    }

    .accordion-summary {
        font-size: 1.25rem;
        font-weight: 700;
        padding: 0.5rem 0;
        margin-bottom: 1rem;
        cursor: pointer;
        list-style: none;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        user-select: none;
        color: var(--colour-text);
    }

    .accordion-summary::-webkit-details-marker {
        display: none;
    }

    .accordion-summary::before {
        content: "▸";
        font-size: 0.75rem;
        color: var(--colour-text-muted);
        transition: transform 0.15s;
        display: inline-block;
    }

    details[open] .accordion-summary::before {
        transform: rotate(90deg);
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

    .success {
        color: var(--colour-success);
        font-size: 0.875rem;
        margin-top: 0.75rem;
    }

    /* SPC Rules */
    .rules-note {
        font-size: 0.875rem;
        color: var(--colour-text-muted);
        margin-bottom: 1.5rem;
    }

    .rule-groups {
        display: flex;
        flex-direction: column;
        gap: 1.5rem;
    }

    .rule-group {
        display: flex;
        flex-direction: column;
        gap: 0.625rem;
    }

    .rule-group-label {
        font-size: 0.75rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.06em;
        color: var(--colour-text-muted);
        margin-bottom: 0.25rem;
    }

    .rule-row {
        display: flex;
        align-items: center;
        gap: 1rem;
        flex-wrap: wrap;
        padding: 0.5rem 0;
        border-bottom: 1px solid var(--colour-border);
        transition: opacity 0.15s;
    }

    .rule-row:last-child {
        border-bottom: none;
    }

    .rule-row.disabled {
        opacity: 0.45;
    }

    .rule-toggle {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
        min-width: 180px;
        /* override label default bold from create-form context */
        font-weight: 500;
    }

    .rule-toggle input[type="checkbox"] {
        width: 1rem;
        height: 1rem;
        border: 1px solid var(--colour-border);
        border-radius: 3px;
        accent-color: var(--colour-primary);
        cursor: pointer;
        /* reset height override from global input rule */
        height: unset;
        padding: unset;
    }

    .rule-desc {
        font-size: 0.875rem;
        color: var(--colour-text-muted);
        display: flex;
        align-items: center;
        gap: 0.375rem;
        flex-wrap: wrap;
    }

    .count-input {
        width: 3.5rem;
        height: unset;
        padding: 0.2rem 0.4rem;
        font-family: var(--font-mono);
        font-size: 0.875rem;
        text-align: center;
    }

    .rules-footer {
        margin-top: 1.5rem;
        display: flex;
        justify-content: flex-end;
    }

    .save-btn {
        background: var(--colour-primary);
        color: white;
        border: none;
        border-radius: var(--radius);
        padding: 0.4rem 1rem;
        font-size: 0.875rem;
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
</style>
