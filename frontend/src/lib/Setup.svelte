<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { CreateAdminUser } from "../../wailsjs/go/handlers/SetupHandler";

    const dispatch = createEventDispatcher();

    let username = "";
    let password = "";
    let confirm = "";
    let error = "";
    let loading = false;

    async function handleSubmit() {
        error = "";
        if (password !== confirm) {
            error = "Passwords do not match";
            return;
        }
        if (password.length < 6) {
            error = "Password must be at least 6 characters";
            return;
        }
        loading = true;
        try {
            await CreateAdminUser(username, password);
            dispatch("done");
        } catch (e: any) {
            error = e;
        } finally {
            loading = false;
        }
    }
</script>

<div class="setup-wrap">
    <div class="setup-card">
        <h1>Welcome to Cromulent</h1>
        <p class="tagline">Let's create your admin account to get started.</p>

        <div class="fields">
            <div class="field">
                <label for="username">Username</label>
                <input
                    id="username"
                    type="text"
                    bind:value={username}
                    disabled={loading}
                    autocomplete="off"
                />
            </div>
            <div class="field">
                <label for="password">Password</label>
                <input
                    id="password"
                    type="password"
                    bind:value={password}
                    disabled={loading}
                />
            </div>
            <div class="field">
                <label for="confirm">Confirm password</label>
                <input
                    id="confirm"
                    type="password"
                    bind:value={confirm}
                    disabled={loading}
                    on:keydown={(e) => e.key === "Enter" && handleSubmit()}
                />
            </div>
            {#if error}
                <p class="error">{error}</p>
            {/if}
            <button
                on:click={handleSubmit}
                disabled={loading || !username || !password}
            >
                {loading ? "Creating account…" : "Create admin account"}
            </button>
        </div>
    </div>
</div>

<style>
    .setup-wrap {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--colour-bg);
    }

    .setup-card {
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius-lg);
        box-shadow: var(--shadow-md);
        padding: 2.5rem;
        width: 100%;
        max-width: 360px;
    }

    h1 {
        font-size: 1.5rem;
        font-weight: 700;
        margin-bottom: 0.25rem;
    }

    .tagline {
        color: var(--colour-text-muted);
        font-size: 0.875rem;
        margin-bottom: 2rem;
    }

    .fields {
        display: flex;
        flex-direction: column;
        gap: 1rem;
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

    input {
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.5rem 0.75rem;
        font-size: 1rem;
        background: var(--colour-bg);
        color: var(--colour-text);
        transition: border-color 0.15s;
    }

    input:focus {
        outline: none;
        border-color: var(--colour-primary);
    }

    input:disabled {
        opacity: 0.6;
    }

    button {
        background: var(--colour-primary);
        color: white;
        border: none;
        border-radius: var(--radius);
        padding: 0.625rem 1rem;
        font-size: 1rem;
        font-weight: 700;
        margin-top: 0.5rem;
        transition: background 0.15s;
    }

    button:hover:not(:disabled) {
        background: var(--colour-primary-hover);
    }

    button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .error {
        color: var(--colour-danger);
        font-size: 0.875rem;
    }
</style>
