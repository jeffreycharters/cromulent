<script lang="ts">
    import { onMount } from "svelte";
    import {
        ListAnalytes,
        CreateAnalyte,
        ListMethods,
        CreateMethod,
        ListMaterials,
        CreateMaterial,
    } from "../../wailsjs/go/handlers/LibraryHandler";
    import {
        ListAllMMAs,
        AddAnalyteToMMA,
        RemoveAnalyteFromMMA,
    } from "../../wailsjs/go/handlers/MMAHandler";

    import { dndzone } from "svelte-dnd-action";
    import {
        UpdateDisplayOrders,
        ListUsedMMAIDs,
    } from "../../wailsjs/go/handlers/MMAHandler";

    let usedMMAIDs: Set<number> = new Set();

    let selectedMethodID: number | null = null;
    let selectedMaterialID: number | null = null;
    let addingAnalyteID: number | null = null;
    let draggableAnalytes: any[] = [];

    type Tab = "analytes" | "methods" | "materials" | "combos";

    const tabs: { id: Tab; label: string }[] = [
        { id: "analytes", label: "Analytes" },
        { id: "methods", label: "Methods" },
        { id: "materials", label: "Materials" },
        { id: "combos", label: "Combos" },
    ];

    let activeTab: Tab = "analytes";

    let analytes: any[] = [];
    let methods: any[] = [];
    let materials: any[] = [];
    let mmas: any[] = [];

    let error = "";
    let success = "";

    // New item forms
    let newAnalyteName = "";
    let newAnalyteUnit = "";
    let newMethodName = "";
    let newMethodDescription = "";
    let newMaterialName = "";
    let newMaterialDescription = "";

    // MMA wiring form
    let mmaMethodID: number | null = null;
    let mmaMaterialID: number | null = null;
    let mmaAnalyteID: number | null = null;
    let mmaDisplayOrder = 0;

    onMount(loadAll);

    async function loadAll() {
        try {
            const [a, me, mat, mm, used] = await Promise.all([
                ListAnalytes(),
                ListMethods(),
                ListMaterials(),
                ListAllMMAs(),
                ListUsedMMAIDs(),
            ]);
            analytes = a ?? [];
            methods = me ?? [];
            materials = mat ?? [];
            mmas = mm ?? [];
            usedMMAIDs = new Set((used ?? []).map(Number));
        } catch (e: any) {
            error = e.toString();
        }
    }

    function flash(msg: string) {
        success = msg;
        setTimeout(() => (success = ""), 3000);
    }

    async function createAnalyte() {
        if (!newAnalyteName.trim()) return;
        try {
            await CreateAnalyte(newAnalyteName.trim(), newAnalyteUnit.trim());
            newAnalyteName = "";
            newAnalyteUnit = "";
            analytes = await ListAnalytes();
            flash("Analyte added.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    async function createMethod() {
        if (!newMethodName.trim()) return;
        try {
            await CreateMethod(
                newMethodName.trim(),
                newMethodDescription.trim(),
            );
            newMethodName = "";
            newMethodDescription = "";
            methods = await ListMethods();
            flash("Method added.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    async function createMaterial() {
        if (!newMaterialName.trim()) return;
        try {
            await CreateMaterial(
                newMaterialName.trim(),
                newMaterialDescription.trim(),
            );
            newMaterialName = "";
            newMaterialDescription = "";
            materials = await ListMaterials();
            flash("Material added.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    async function addMMA() {
        if (!mmaMethodID || !mmaMaterialID || !mmaAnalyteID) {
            error = "Select a method, material, and analyte.";
            return;
        }
        try {
            await AddAnalyteToMMA(
                mmaMaterialID,
                mmaMethodID,
                mmaAnalyteID,
                mmaDisplayOrder,
            );
            mmaAnalyteID = null;
            mmaDisplayOrder = 0;
            mmas = await ListAllMMAs();
            flash("Analyte added to combo.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    async function removeMMA(id: number) {
        try {
            await RemoveAnalyteFromMMA(id);
            mmas = await ListAllMMAs();
            flash("Removed.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    async function handleDndFinalize(e: CustomEvent) {
        draggableAnalytes = e.detail.items;
        try {
            const ids = draggableAnalytes.map((a) => a.id);
            const orders = draggableAnalytes.map((_, i) => i);
            await UpdateDisplayOrders(ids, orders);
            mmas = await ListAllMMAs();
        } catch (e: any) {
            error = e.toString();
        }
    }

    interface MMAMaterialGroup {
        materialName: string;
        analytes: any[];
    }
    interface MMAMethodGroup {
        methodName: string;
        materials: Record<string, MMAMaterialGroup>;
    }

    // Group MMAs by method → material for display

    let mmaGrouped: Record<string, MMAMethodGroup> = {};

    $: mmaGrouped = mmas.reduce(
        (acc: Record<string, MMAMethodGroup>, entry: any) => {
            const mk = `${entry.method_id}`;
            if (!acc[mk])
                acc[mk] = { methodName: entry.method_name, materials: {} };
            const matk = `${entry.material_id}`;
            if (!acc[mk].materials[matk])
                acc[mk].materials[matk] = {
                    materialName: entry.material_name,
                    analytes: [],
                };
            acc[mk].materials[matk].analytes.push(entry);
            return acc;
        },
        {} as Record<string, MMAMethodGroup>,
    );

    $: {
        if (mmaMethodID && mmaMaterialID) {
            const existing = mmas.filter(
                (m) =>
                    m.method_id === mmaMethodID &&
                    m.material_id === mmaMaterialID,
            );
            mmaDisplayOrder = existing.length
                ? Math.max(...existing.map((m) => m.display_order)) + 1
                : 0;
        }
    }

    $: orderConflict =
        mmaMethodID && mmaMaterialID && mmaAnalyteID
            ? mmas.some(
                  (m) =>
                      m.method_id === mmaMethodID &&
                      m.material_id === mmaMaterialID &&
                      m.display_order === mmaDisplayOrder,
              )
            : false;

    $: if (selectedMethodID && selectedMaterialID) {
        draggableAnalytes = mmas
            .filter(
                (m) =>
                    m.method_id === selectedMethodID &&
                    m.material_id === selectedMaterialID,
            )
            .sort((a, b) => a.display_order - b.display_order)
            .map((m, i) => ({ ...m, id: m.id }));
    }

    function handleDndConsider(e: CustomEvent) {
        draggableAnalytes = e.detail.items;
    }
    async function addAnalyteToCurrentCombo() {
        if (!selectedMethodID || !selectedMaterialID || !addingAnalyteID)
            return;
        const nextOrder = draggableAnalytes.length;
        try {
            await AddAnalyteToMMA(
                selectedMaterialID,
                selectedMethodID,
                addingAnalyteID,
                nextOrder,
            );
            addingAnalyteID = null;
            mmas = await ListAllMMAs();
            flash("Analyte added.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    $: materialsForMethod = selectedMethodID
        ? [
              ...materials.filter((mat) =>
                  mmas.some(
                      (m) =>
                          m.method_id === selectedMethodID &&
                          m.material_id === mat.id,
                  ),
              ),
              ...materials.filter(
                  (mat) =>
                      !mmas.some(
                          (m) =>
                              m.method_id === selectedMethodID &&
                              m.material_id === mat.id,
                      ),
              ),
          ]
        : materials;

    $: analytesNotInCombo =
        selectedMethodID && selectedMaterialID
            ? analytes.filter(
                  (a) =>
                      !mmas.some(
                          (m) =>
                              m.method_id === selectedMethodID &&
                              m.material_id === selectedMaterialID &&
                              m.analyte_id === a.id,
                      ),
              )
            : [];
</script>

<div class="library">
    {#if error}
        <div class="banner error" role="alert">
            {error}
            <button on:click={() => (error = "")}>✕</button>
        </div>
    {/if}
    {#if success}
        <div class="banner success" role="status">{success}</div>
    {/if}

    <div class="tabs">
        {#each tabs as tab}
            <button
                class="tab"
                class:active={activeTab === tab.label}
                on:click={() => (activeTab = tab.id)}
            >
                {tab.label.charAt(0).toUpperCase() + tab.label.slice(1)}
            </button>
        {/each}
    </div>

    <div class="tab-content">
        {#if activeTab === "analytes"}
            <div class="add-form">
                <input
                    bind:value={newAnalyteName}
                    placeholder="Analyte name (e.g. Lead)"
                />
                <input
                    bind:value={newAnalyteUnit}
                    placeholder="Unit (e.g. mg/kg)"
                />
                <button on:click={createAnalyte}>Add</button>
            </div>
            <table>
                <thead><tr><th>Name</th><th>Unit</th></tr></thead>
                <tbody>
                    {#each analytes as a}
                        <tr><td>{a.name}</td><td>{a.unit}</td></tr>
                    {:else}
                        <tr
                            ><td colspan="2" class="empty">No analytes yet.</td
                            ></tr
                        >
                    {/each}
                </tbody>
            </table>
        {:else if activeTab === "methods"}
            <div class="add-form">
                <input
                    bind:value={newMethodName}
                    placeholder="Method name (e.g. EPA 200.8)"
                />
                <input
                    bind:value={newMethodDescription}
                    placeholder="Description (optional)"
                />
                <button on:click={createMethod}>Add</button>
            </div>
            <table>
                <thead><tr><th>Name</th><th>Description</th></tr></thead>
                <tbody>
                    {#each methods as m}
                        <tr><td>{m.name}</td><td>{m.description}</td></tr>
                    {:else}
                        <tr
                            ><td colspan="2" class="empty">No methods yet.</td
                            ></tr
                        >
                    {/each}
                </tbody>
            </table>
        {:else if activeTab === "materials"}
            <div class="add-form">
                <input
                    bind:value={newMaterialName}
                    placeholder="Material name (e.g. Bovine Liver CRM)"
                />
                <input
                    bind:value={newMaterialDescription}
                    placeholder="Description (optional)"
                />
                <button on:click={createMaterial}>Add</button>
            </div>
            <table>
                <thead><tr><th>Name</th><th>Description</th></tr></thead>
                <tbody>
                    {#each materials as m}
                        <tr><td>{m.name}</td><td>{m.description}</td></tr>
                    {:else}
                        <tr
                            ><td colspan="2" class="empty">No materials yet.</td
                            ></tr
                        >
                    {/each}
                </tbody>
            </table>
        {:else if activeTab === "combos"}
            <div class="combos-selectors">
                <div class="field">
                    <label>Method</label>
                    <select
                        bind:value={selectedMethodID}
                        on:change={() => {
                            selectedMaterialID = null;
                            addingAnalyteID = null;
                        }}
                    >
                        <option value={null}>— Select method —</option>
                        {#each methods as m}
                            <option value={m.id}>{m.name}</option>
                        {/each}
                    </select>
                </div>

                {#if selectedMethodID}
                    <div class="field">
                        <label>Material</label>
                        <select
                            bind:value={selectedMaterialID}
                            on:change={() => {
                                addingAnalyteID = null;
                            }}
                        >
                            <option value={null}>— Select material —</option>
                            {#each materialsForMethod as mat}
                                {@const linked = mmas.some(
                                    (m) =>
                                        m.method_id === selectedMethodID &&
                                        m.material_id === mat.id,
                                )}
                                <option value={mat.id}
                                    >{mat.name}{linked ? " ✓" : ""}</option
                                >
                            {/each}
                        </select>
                    </div>
                {/if}
            </div>

            {#if selectedMethodID && selectedMaterialID}
                <div class="analyte-list">
                    {#if draggableAnalytes.length}
                        <p class="hint">
                            Drag to reorder — order determines column position
                            in the data entry grid.
                        </p>
                        <ul
                            use:dndzone={{
                                items: draggableAnalytes,
                                flipDurationMs: 150,
                            }}
                            on:consider={handleDndConsider}
                            on:finalize={handleDndFinalize}
                            class="dnd-list"
                        >
                            {#each draggableAnalytes as entry (entry.id)}
                                <li class="dnd-item">
                                    <span class="drag-handle">⠿</span>
                                    <span class="analyte-name"
                                        >{entry.analyte_name}</span
                                    >
                                    <span class="analyte-unit"
                                        >{entry.unit}</span
                                    >
                                    {#if !usedMMAIDs.has(entry.id)}
                                        <button
                                            class="remove-btn"
                                            on:click={() => removeMMA(entry.id)}
                                            >Remove</button
                                        >
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    {:else}
                        <p class="empty">No analytes added yet.</p>
                    {/if}

                    {#if analytesNotInCombo.length}
                        <div class="add-analyte-row">
                            <select bind:value={addingAnalyteID}>
                                <option value={null}>— Add analyte —</option>
                                {#each analytesNotInCombo as a}
                                    <option value={a.id}
                                        >{a.name} ({a.unit})</option
                                    >
                                {/each}
                            </select>
                            <button
                                on:click={addAnalyteToCurrentCombo}
                                disabled={!addingAnalyteID}>Add</button
                            >
                        </div>
                    {:else}
                        <p class="hint">All analytes added for this combo.</p>
                    {/if}
                </div>
            {/if}
        {/if}
    </div>
</div>

<style>
    .library {
        max-width: 800px;
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
    .banner.success {
        background: color-mix(in srgb, var(--colour-success) 12%, transparent);
        color: var(--colour-success);
        border: 1px solid
            color-mix(in srgb, var(--colour-success) 30%, transparent);
    }
    .tabs {
        display: flex;
        gap: 0.25rem;
        border-bottom: 1px solid var(--colour-border);
        margin-bottom: 1.5rem;
    }
    .tab {
        background: none;
        border: none;
        border-bottom: 2px solid transparent;
        padding: 0.5rem 1rem;
        font-size: 0.9rem;
        color: var(--colour-text-muted);
        margin-bottom: -1px;
        border-radius: 0;
        transition:
            color 0.15s,
            border-color 0.15s;
    }
    .tab:hover {
        color: var(--colour-text);
    }
    .tab.active {
        color: var(--colour-primary);
        border-bottom-color: var(--colour-primary);
        font-weight: 600;
    }
    .add-form {
        display: flex;
        gap: 0.5rem;
        margin-bottom: 1rem;
        flex-wrap: wrap;
        align-items: flex-start;
    }
    .add-form input {
        min-width: 10rem;
    }
    .add-form button {
        flex-shrink: 0;
    }
    table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.9rem;
    }
    th {
        text-align: left;
        padding: 0.5rem 0.75rem;
        border-bottom: 1px solid var(--colour-border);
        color: var(--colour-text-muted);
        font-weight: 600;
        font-size: 0.8rem;
        text-transform: uppercase;
        letter-spacing: 0.04em;
    }
    td {
        padding: 0.5rem 0.75rem;
        border-bottom: 1px solid var(--colour-border);
    }
    tr:last-child td {
        border-bottom: none;
    }
    .empty {
        color: var(--colour-text-muted);
        font-style: italic;
        text-align: center;
        padding: 1.5rem;
    }
    .hint {
        font-size: 0.875rem;
        color: var(--colour-text-muted);
        margin-bottom: 1rem;
    }
    .remove-btn {
        background: none;
        border: none;
        color: var(--colour-danger);
        font-size: 0.8rem;
        padding: 0.125rem 0.375rem;
        cursor: pointer;
    }
    .remove-btn:hover {
        text-decoration: underline;
    }
    .combos-selectors {
        display: flex;
        gap: 1rem;
        margin-bottom: 1.5rem;
        flex-wrap: wrap;
    }
    .field {
        display: flex;
        flex-direction: column;
        gap: 0.375rem;
        flex: 1;
        min-width: 12rem;
    }
    .field label {
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--colour-text-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
    }
    .dnd-list {
        list-style: none;
        padding: 0;
        margin: 0 0 1rem 0;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }
    .dnd-item {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.5rem 0.75rem;
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        cursor: grab;
    }
    .drag-handle {
        color: var(--colour-text-muted);
        font-size: 1.1rem;
        cursor: grab;
    }
    .analyte-name {
        flex: 1;
        font-size: 0.9rem;
    }
    .analyte-unit {
        font-size: 0.8rem;
        color: var(--colour-text-muted);
    }
    .add-analyte-row {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
    .analyte-list {
        max-width: 480px;
    }
</style>
