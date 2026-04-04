<script lang="ts">
    import { Login } from "../../wailsjs/go/handlers/AuthHandler";

    let username = "";
    let password = "";
    let error = "";
    let loading = false;

    async function handleSubmit() {
        error = "";
        loading = true;
        try {
            const result = await Login(username, password);
            dispatch("success", result);
        } catch (e: any) {
            error = e;
        } finally {
            loading = false;
        }
    }

    import { createEventDispatcher } from "svelte";
    const dispatch = createEventDispatcher();
</script>

<div class="login-wrap">
    <div class="login-card">
        <h1>Cromulent</h1>
        <p class="tagline">perfectly cromulent data, every time</p>

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
                    on:keydown={(e) => e.key === "Enter" && handleSubmit()}
                />
            </div>
            {#if error}
                <p class="error">{error}</p>
            {/if}
            <button on:click={handleSubmit} disabled={loading}>
                {loading ? "Signing in…" : "Sign in"}
            </button>
        </div>
    </div>
</div>

<style>
    .login-wrap {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--colour-bg);
    }

    .login-card {
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
        font-style: italic;
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
