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
        UpdateDisplayOrders,
        ListUsedMMAIDs,
        ListAllMMAs,
        AddAnalyteToMMA,
        RemoveAnalyteFromMMA,
        DeactivateCombo,
        ActivateCombo,
    } from "../../wailsjs/go/handlers/MMAHandler";
    import {
        GetCurrentSequencesForMMAs,
        ListControlLimitRegionsForCombo,
        SaveControlLimitRegions,
        DeleteControlLimitRegionSet,
    } from "../../wailsjs/go/handlers/LimitsHandler";
    import { dndzone } from "svelte-dnd-action";

    export let currentUser: any;

    let usedMMAIDs: Set<number> = new Set();
    let selectedMethodID: number | null = null;
    let selectedMaterialID: number | null = null;
    let addingAnalyteID: number | null = null;
    let draggableAnalytes: any[] = [];

    let showHidden = false;

    type Tab = "analytes" | "methods" | "materials" | "combos" | "limits";

    const tabs: { id: Tab; label: string }[] = [
        { id: "analytes", label: "Analytes" },
        { id: "methods", label: "Methods" },
        { id: "materials", label: "Materials" },
        { id: "combos", label: "Combos" },
        { id: "limits", label: "Limits" },
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

    const LIMIT_ROWS = [
        "UCL",
        "UWL",
        "UIL",
        "Mean",
        "LIL",
        "LWL",
        "LCL",
    ] as const;
    type LimitRow = (typeof LIMIT_ROWS)[number];

    interface ComboCard {
        methodID: number;
        methodName: string;
        materialID: number;
        materialName: string;
        mmaIDs: number[];
    }

    let limitComboCards: ComboCard[] = [];
    let selectedLimitCombo: ComboCard | null = null;
    let limitAnalytes: any[] = []; // ordered analytes for selected combo
    let existingRegions: any[] = [];
    let currentSequences: Record<number, number> = {};

    // New region paste grid: LIMIT_ROWS × analyte columns
    // grid[rowIndex][colIndex] = string value
    let newGrid: string[][] = [];
    let newEffectiveFrom: number = 0;
    let savingLimits = false;

    $: comboActive =
        draggableAnalytes.length > 0 ? draggableAnalytes[0].active : true;

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

    // Parse clipboard data into a 2D array of strings
    function parseClipboard(e: ClipboardEvent): string[][] {
        const html = e.clipboardData?.getData("text/html") ?? "";
        if (html) {
            const doc = new DOMParser().parseFromString(html, "text/html");
            const rows = doc.querySelectorAll("tr");
            if (rows.length > 0) {
                return Array.from(rows).map((row) =>
                    Array.from(row.querySelectorAll("td, th")).map(
                        (cell) => cell.textContent?.trim() ?? "",
                    ),
                );
            }
        }
        const text = e.clipboardData?.getData("text/plain") ?? "";
        return text
            .split(/\r?\n/)
            .filter((line) => line.trim() !== "")
            .map((line) => line.split("\t"));
    }

    function isNumeric(s: string): boolean {
        return /^-?\d+(\.\d+)?$/.test(s.trim());
    }

    $: if (limitAnalytes.length) {
        newGrid = LIMIT_ROWS.map(() => limitAnalytes.map(() => ""));
    }

    // --- Limits tab logic ---

    $: limitComboCards = Object.values(
        mmas
            .filter((m) => m.active)
            .reduce((acc: Record<string, ComboCard>, m: any) => {
                const key = `${m.method_id}_${m.material_id}`;
                if (!acc[key]) {
                    acc[key] = {
                        methodID: m.method_id,
                        methodName: m.method_name,
                        materialID: m.material_id,
                        materialName: m.material_name,
                        mmaIDs: [],
                    };
                }
                acc[key].mmaIDs.push(m.id);
                return acc;
            }, {}),
    );

    async function selectLimitCombo(card: ComboCard) {
        selectedLimitCombo = card;
        error = "";
        try {
            const ordered = mmas
                .filter(
                    (m) =>
                        m.method_id === card.methodID &&
                        m.material_id === card.materialID &&
                        m.active,
                )
                .sort((a, b) => a.display_order - b.display_order);
            limitAnalytes = ordered;

            const [regions, seqs] = await Promise.all([
                ListControlLimitRegionsForCombo(card.materialID, card.methodID),
                GetCurrentSequencesForMMAs(card.mmaIDs),
            ]);

            existingRegions = regions ?? [];
            currentSequences = seqs ?? {};

            // Default effective_from to max sequence across MMAs + 1, or 0
            const maxSeq = Math.max(0, ...Object.values(currentSequences));
            newEffectiveFrom = maxSeq;

            resetNewGrid();
        } catch (e: any) {
            error = e.toString();
        }
    }

    function resetNewGrid() {
        newGrid = LIMIT_ROWS.map(() => limitAnalytes.map(() => ""));
    }

    function handleLimitPaste(
        e: ClipboardEvent,
        fromRow: number,
        fromCol: number,
    ) {
        e.preventDefault();
        const parsed = parseClipboard(e);
        const next = newGrid.map((r) => [...r]);
        parsed.forEach((row, ri) => {
            const targetRow = fromRow + ri;
            if (targetRow >= LIMIT_ROWS.length) return;
            row.forEach((cell, ci) => {
                const targetCol = fromCol + ci;
                if (targetCol >= limitAnalytes.length) return;
                next[targetRow][targetCol] = isNumeric(cell) ? cell.trim() : "";
            });
        });
        newGrid = next;
    }

    // Group existing regions by effective_from_sequence
    $: groupedRegions = existingRegions.reduce(
        (acc: Record<number, any[]>, r: any) => {
            const k = r.effective_from_sequence;
            if (!acc[k]) acc[k] = [];
            acc[k].push(r);
            return acc;
        },
        {},
    );

    function getRegionValue(
        regions: unknown,
        mmaID: number,
        row: LimitRow,
    ): string {
        const list = regions as any[];
        if (!list) return "—";
        const r = list.find((x) => x.mma_id === mmaID);
        if (!r) return "—";
        const map: Record<LimitRow, number | null> = {
            UCL: r.ucl,
            UWL: r.uwl,
            UIL: r.uil,
            Mean: r.mean,
            LIL: r.lil,
            LWL: r.lwl,
            LCL: r.lcl,
        };
        const v = map[row];
        return v !== null && v !== undefined ? String(v) : "—";
    }

    async function deleteRegionSet(effectiveFromSequence: number) {
        if (!selectedLimitCombo) return;
        try {
            await DeleteControlLimitRegionSet(
                selectedLimitCombo.materialID,
                selectedLimitCombo.methodID,
                effectiveFromSequence,
                currentUser.id,
            );
            existingRegions =
                (await ListControlLimitRegionsForCombo(
                    selectedLimitCombo.materialID,
                    selectedLimitCombo.methodID,
                )) ?? [];
            flash("Limit set removed.");
        } catch (e: any) {
            error = e.toString();
        }
    }

    async function saveLimits() {
        if (!selectedLimitCombo) return;
        savingLimits = true;
        error = "";
        try {
          const regions = limitAnalytes
              .map((a, ci) => {
                  const get = (ri: number) => {
                      const v = newGrid[ri][ci];
                      return isNumeric(v) ? parseFloat(v) : null;
                  };
                  return {
                      id: 0,
                      mma_id: a.id,
                      ucl: get(0),
                      uwl: get(1),
                      uil: get(2),
                      mean: get(3),
                      lil: get(4),
                      lwl: get(5),
                      lcl: get(6),
                      effective_from_sequence: newEffectiveFrom,
                      created_by: currentUser.id,
                      created_at: "",
                  };
              })
              .filter((r) => r.ucl !== null && r.lcl !== null);

          if (regions.length === 0) {
              error = "At least one analyte must have UCL and LCL set.";
              savingLimits = false;
              return;
          }

            await SaveControlLimitRegions(regions);
            existingRegions =
                (await ListControlLimitRegionsForCombo(
                    selectedLimitCombo.materialID,
                    selectedLimitCombo.methodID,
                )) ?? [];
            resetNewGrid();
            flash("Limits saved.");
        } catch (e: any) {
            error = e.toString();
        } finally {
            savingLimits = false;
        }
    }

    // --- Existing tab logic unchanged ---

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

    function comboActiveForMat(matID: number): boolean {
        const combo = mmas.find(
            (m) => m.method_id === selectedMethodID && m.material_id === matID,
        );
        if (!combo) return true; // no combo exists yet — not inactive, just unlinked
        return combo.active;
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
            .map((m) => ({ ...m, id: m.id }));
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

    async function deactivateCombo() {
        if (!selectedMaterialID || !selectedMethodID) return;
        await DeactivateCombo(selectedMaterialID, selectedMethodID);
        mmas = await ListAllMMAs();
        flash("Combo deactivated.");
    }

    async function activateCombo() {
        if (!selectedMaterialID || !selectedMethodID) return;
        await ActivateCombo(selectedMaterialID, selectedMethodID);
        mmas = await ListAllMMAs();
        flash("Combo activated.");
    }

    $: materialsForMethod = selectedMethodID
        ? materials.filter((mat) => {
              const combo = mmas.find(
                  (m) =>
                      m.method_id === selectedMethodID &&
                      m.material_id === mat.id,
              );
              if (!combo) return true; // unlinked, always show
              if (combo.active) return true; // active, always show
              return showHidden; // inactive, only show if showHidden
          })
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
                class:active={activeTab === tab.id}
                on:click={() => {
                    activeTab = tab.id;
                    if (tab.id !== "limits") selectedLimitCombo = null;
                }}
            >
                {tab.label}
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
                <label class="show-hidden-label">
                    <input type="checkbox" bind:checked={showHidden} />
                    Show inactive
                </label>
                <div class="combos-fields">
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
                                <option value={null}>— Select material —</option
                                >
                                {#each materialsForMethod as mat}
                                    {@const linked = mmas.some(
                                        (m) =>
                                            m.method_id === selectedMethodID &&
                                            m.material_id === mat.id,
                                    )}
                                    <option value={mat.id}>
                                        {mat.name}{linked
                                            ? " ✓"
                                            : ""}{showHidden &&
                                        !comboActiveForMat(mat.id)
                                            ? " (inactive)"
                                            : ""}
                                    </option>
                                {/each}
                            </select>
                        </div>
                    {/if}
                </div>

                {#if selectedMethodID && selectedMaterialID}
                    <div class="combo-actions">
                        {#if comboActive}
                            <button
                                class="deactivate-btn"
                                on:click={deactivateCombo}
                                >Deactivate combo</button
                            >
                        {:else}
                            <button
                                class="activate-btn"
                                on:click={activateCombo}>Activate combo</button
                            >
                        {/if}
                    </div>
                    <div class="analyte-list">
                        {#if draggableAnalytes.length}
                            <p class="hint">
                                Drag to reorder — order determines column
                                position in the data entry grid.
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
                                                on:click={() =>
                                                    removeMMA(entry.id)}
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
                                    <option value={null}>— Add analyte —</option
                                    >
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
                            <p class="hint">
                                All analytes added for this combo.
                            </p>
                        {/if}
                    </div>
                {/if}
            </div>
        {:else if activeTab === "limits"}
            {#if !selectedLimitCombo}
                {#if limitComboCards.length === 0}
                    <p class="empty">
                        No active combos. Set them up in the Combos tab first.
                    </p>
                {:else}
                    <div class="combo-cards">
                        {#each limitComboCards as card}
                            <button
                                class="combo-card"
                                on:click={() => selectLimitCombo(card)}
                            >
                                <span class="card-method"
                                    >{card.methodName}</span
                                >
                                <span class="card-material"
                                    >{card.materialName}</span
                                >
                            </button>
                        {/each}
                    </div>
                {/if}
            {:else}
                <div class="limits-header">
                    <div>
                        <button
                            class="back-btn"
                            on:click={() => {
                                selectedLimitCombo = null;
                            }}>← Back</button
                        >
                        <h2 class="limits-title">
                            {selectedLimitCombo.methodName}
                        </h2>
                        <p class="limits-subtitle">
                            {selectedLimitCombo.materialName}
                        </p>
                    </div>
                </div>

                <!-- Existing region sets -->
                {#each Object.entries(groupedRegions).sort((a, b) => Number(a[0]) - Number(b[0])) as [seq, regions]}
                    <div class="region-set">
                        <div class="region-set-header">
                            <span class="region-seq"
                                >Effective from sequence {seq}</span
                            >
                            <button
                                class="remove-btn"
                                on:click={() => deleteRegionSet(Number(seq))}
                                >Delete set</button
                            >
                        </div>
                        <div class="limit-grid-wrap">
                            <table class="limit-table">
                                <thead>
                                    <tr>
                                        <th class="row-label-cell"></th>
                                        {#each limitAnalytes as a}
                                            <th>
                                                <div class="col-header">
                                                    <span>{a.analyte_name}</span
                                                    >
                                                    <span class="col-unit"
                                                        >{a.unit}</span
                                                    >
                                                </div>
                                            </th>
                                        {/each}
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each LIMIT_ROWS as row}
                                        <tr>
                                            <td class="row-label">{row}</td>
                                            {#each limitAnalytes as a}
                                                <td class="limit-cell">
                                                    {getRegionValue(
                                                        regions,
                                                        a.id,
                                                        row,
                                                    )}
                                                </td>
                                            {/each}
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    </div>
                {/each}

                <!-- New region paste grid -->
                <div class="region-set new-region">
                    <div class="region-set-header">
                        <span class="region-seq">
                            New limits — effective from sequence
                            <input
                                type="number"
                                class="seq-input"
                                bind:value={newEffectiveFrom}
                                min="0"
                            />
                        </span>
                        <button
                            class="save-btn"
                            disabled={savingLimits}
                            on:click={saveLimits}
                            >{savingLimits ? "Saving…" : "Save Limits"}</button
                        >
                    </div>
                    <p class="hint">
                        Paste from Excel starting at any cell — data fills right
                        and down from that position.
                    </p>
                    <div class="limit-grid-wrap">
                        <table class="limit-table">
                            <thead>
                                <tr>
                                    <th class="row-label-cell"></th>
                                    {#each limitAnalytes as a}
                                        <th>
                                            <div class="col-header">
                                                <span>{a.analyte_name}</span>
                                                <span class="col-unit"
                                                    >{a.unit}</span
                                                >
                                            </div>
                                        </th>
                                    {/each}
                                </tr>
                            </thead>
                            <tbody>
                                {#each LIMIT_ROWS as row, ri}
                                    <tr>
                                        <td class="row-label">{row}</td>
                                        {#each limitAnalytes as a, ci}
                                            <td class="limit-cell editable">
                                                <input
                                                    type="text"
                                                    class="limit-input"
                                                    class:filled={newGrid[ri]?.[
                                                        ci
                                                    ] !== ""}
                                                    bind:value={newGrid[ri][ci]}
                                                    placeholder="—"
                                                    on:paste={(e) =>
                                                        handleLimitPaste(
                                                            e,
                                                            ri,
                                                            ci,
                                                        )}
                                                />
                                            </td>
                                        {/each}
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                </div>
            {/if}
        {/if}
    </div>
</div>

<style>
    .library {
        max-width: 100%;
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
        cursor: pointer;
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
        flex-direction: column;
        gap: 1rem;
        margin-bottom: 1.5rem;
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

    /* Limits tab */
    .combo-cards {
        display: flex;
        flex-wrap: wrap;
        gap: 0.75rem;
    }
    .combo-card {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        padding: 0.75rem 1.25rem;
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius-lg);
        cursor: pointer;
        text-align: left;
        transition:
            border-color 0.15s,
            box-shadow 0.15s;
        min-width: 160px;
    }
    .combo-card:hover {
        border-color: var(--colour-primary);
        box-shadow: var(--shadow-sm);
    }
    .card-method {
        font-size: 0.875rem;
        font-weight: 600;
        color: var(--colour-text);
    }
    .card-material {
        font-size: 0.8rem;
        color: var(--colour-text-muted);
    }
    .limits-header {
        margin-bottom: 1.5rem;
    }
    .back-btn {
        background: none;
        border: none;
        color: var(--colour-text-muted);
        font-size: 0.85rem;
        cursor: pointer;
        padding: 0;
        margin-bottom: 0.5rem;
        display: block;
    }
    .back-btn:hover {
        color: var(--colour-text);
    }
    .limits-title {
        font-size: 1.25rem;
        font-weight: 700;
        margin-bottom: 0.125rem;
    }
    .limits-subtitle {
        font-size: 0.875rem;
        color: var(--colour-text-muted);
    }
    .region-set {
        margin-bottom: 2rem;
    }
    .region-set-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 0.75rem;
    }
    .region-seq {
        font-size: 0.85rem;
        font-weight: 600;
        color: var(--colour-text-muted);
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    .seq-input {
        width: 5rem;
        font-family: var(--font-mono);
        font-size: 0.85rem;
        padding: 0.2rem 0.4rem;
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
    .limit-grid-wrap {
        overflow-x: auto;
    }
    .limit-table {
        border-collapse: collapse;
        font-size: 0.875rem;
        width: auto;
    }
    .limit-table th {
        text-transform: none;
        letter-spacing: 0;
        font-size: 0.8rem;
        white-space: nowrap;
        min-width: 90px;
    }
    .col-header {
        display: flex;
        flex-direction: column;
        gap: 0.1rem;
    }
    .col-unit {
        font-size: 0.7rem;
        color: var(--colour-text-muted);
        font-weight: 400;
    }
    .row-label-cell {
        width: 3rem;
    }
    .row-label {
        font-size: 0.75rem;
        font-weight: 700;
        color: var(--colour-text-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
        white-space: nowrap;
        padding-right: 1rem;
    }
    .limit-cell {
        text-align: right;
        font-family: var(--font-mono);
        font-size: 0.875rem;
        padding: 0.375rem 0.75rem;
    }
    .limit-cell.editable {
        padding: 0.125rem 0.375rem;
    }
    .limit-input {
        width: 100%;
        border: none;
        background: none;
        font-family: var(--font-mono);
        font-size: 0.875rem;
        color: var(--colour-text);
        text-align: right;
        outline: none;
        padding: 0.25rem 0.375rem;
    }
    .limit-input.filled {
        color: var(--colour-primary);
    }
    .limit-input::placeholder {
        color: var(--colour-border);
    }
    .new-region {
        border-top: 2px dashed var(--colour-border);
        padding-top: 1.5rem;
    }
    .combo-actions {
        margin-bottom: 1rem;
    }
    .deactivate-btn {
        background: none;
        border: 1px solid var(--colour-danger);
        color: var(--colour-danger);
        border-radius: var(--radius);
        padding: 0.3rem 0.75rem;
        font-size: 0.8rem;
        cursor: pointer;
    }
    .deactivate-btn:hover {
        background: color-mix(in srgb, var(--colour-danger) 8%, transparent);
    }
    .activate-btn {
        background: none;
        border: 1px solid var(--colour-success);
        color: var(--colour-success);
        border-radius: var(--radius);
        padding: 0.3rem 0.75rem;
        font-size: 0.8rem;
        cursor: pointer;
    }
    .activate-btn:hover {
        background: color-mix(in srgb, var(--colour-success) 8%, transparent);
    }
    .show-hidden-label {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        font-size: 0.8rem;
        color: var(--colour-text-muted);
        margin-bottom: 1rem;
        cursor: pointer;
    }
    .combos-fields {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
    }
</style>
