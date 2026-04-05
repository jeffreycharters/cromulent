<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Chart, registerables } from "chart.js";
    import annotationPlugin from "chartjs-plugin-annotation";
    import type { models } from "../../wailsjs/go/models";
    import {
        ListMethodsWithMaterials,
        GetAnalytesForCombo,
    } from "../../wailsjs/go/handlers/DataEntryHandler";
    import { GetComboChartData } from "../../wailsjs/go/handlers/ChartReviewHandler";

    Chart.register(...registerables, annotationPlugin);

    // --- types ---

    type ComboAnalyte = models.ComboAnalyte;

    interface ChartPoint {
        sequence_number: number;
        value: number;
        mean: number | null;
        ucl: number | null;
        lcl: number | null;
        uwl: number | null;
        lwl: number | null;
        uil: number | null;
        lil: number | null;
    }

    interface ComboOption {
        methodID: number;
        methodName: string;
        materialID: number;
        materialName: string;
    }

    // --- state ---

    let combos: ComboOption[] = [];
    let selectedMethodID: number | null = null;
    let selectedMaterialID: number | null = null;

    let analytes: ComboAnalyte[] = [];
    let chartData: Record<string, ChartPoint[]> = {};
    let chartInstances: Record<string, Chart> = {};

    let pointLimit = 50;
    let chartsPerRow = 3;
    let loading = false;
    let error = "";
    let showRawData = false;

    // --- lifecycle ---

    onMount(async () => {
        const methods = (await ListMethodsWithMaterials()) ?? [];
        combos = methods.flatMap((m) =>
            (m.materials ?? []).map((mat) => ({
                methodID: m.id,
                methodName: m.name,
                materialID: mat.id,
                materialName: mat.name,
            })),
        );
    });

    onDestroy(() => {
        destroyCharts();
    });

    $: if (!loading && !showRawData && analytes.length > 0) {
        import("svelte").then(({ tick }) => tick().then(() => buildCharts()));
    }

    // --- actions ---

    async function selectCombo(combo: ComboOption) {
        selectedMethodID = combo.methodID;
        selectedMaterialID = combo.materialID;
        await loadCombo();
    }

    async function loadCombo() {
        if (!selectedMethodID || !selectedMaterialID) return;
        loading = true;
        error = "";
        destroyCharts();
        chartData = {};
        analytes = [];
        try {
            const [ana, data] = await Promise.all([
                GetAnalytesForCombo(selectedMethodID, selectedMaterialID),
                GetComboChartData(
                    selectedMethodID,
                    selectedMaterialID,
                    pointLimit,
                ),
            ]);
            analytes = (ana ?? []).sort(
                (a, b) => a.display_order - b.display_order,
            );
            chartData = data ?? {};
        } catch (e: any) {
            error = e?.toString() ?? "Failed to load chart data";
        } finally {
            loading = false;
        }
    }

    // --- chart management ---

    function destroyCharts() {
        for (const c of Object.values(chartInstances)) c.destroy();
        chartInstances = {};
    }

    function buildCharts() {
        destroyCharts();
        for (const analyte of analytes) {
            const points: ChartPoint[] =
                chartData[String(analyte.mma_id)] ?? [];
            if (points.length === 0) continue;

            const labels = points.map((p) => String(p.sequence_number));
            while (labels.length < 20) labels.push("");

            const mrValues = points.map((p, i) =>
                i === 0 ? null : Math.abs(p.value - points[i - 1].value),
            );

            const xCanvas = document.getElementById(
                `chart-x-${analyte.mma_id}`,
            ) as HTMLCanvasElement | null;
            if (xCanvas) {
                chartInstances[`${analyte.mma_id}-x`] = new Chart(xCanvas, {
                    type: "line",
                    data: {
                        labels,
                        datasets: [
                            {
                                label: "Value",
                                data: points.map((p) => p.value),
                                borderColor: "var(--colour-primary)",
                                backgroundColor: "transparent",
                                pointRadius: 3,
                                tension: 0,
                                spanGaps: false,
                            },
                        ],
                    },
                    options: xChartOptions(points),
                });
            }

            const mrCanvas = document.getElementById(
                `chart-mr-${analyte.mma_id}`,
            ) as HTMLCanvasElement | null;
            if (mrCanvas) {
                chartInstances[`${analyte.mma_id}-mr`] = new Chart(mrCanvas, {
                    type: "line",
                    data: {
                        labels,
                        datasets: [
                            {
                                label: "mR",
                                data: mrValues,
                                borderColor: "var(--colour-text-muted)",
                                backgroundColor: "transparent",
                                pointRadius: 3,
                                tension: 0,
                                spanGaps: false,
                            },
                        ],
                    },
                    options: mrChartOptions(mRucl(points), points),
                });
            }
        }
    }

    // --- chart options ---

    function sigFigs(value: number, n = 3): string {
        if (value === 0) return "0";
        const d = Math.ceil(Math.log10(Math.abs(value)));
        const power = n - d;
        const magnitude = Math.pow(10, power);
        return String(Math.round(value * magnitude) / magnitude);
    }

    function mRucl(points: ChartPoint[]): number | null {
        const first = points.find((p) => p.ucl != null && p.lcl != null);
        if (!first) return null;
        return (first.ucl! - first.lcl!) * 0.61;
    }

    function annotationsForX(points: ChartPoint[], yMin: number, yMax: number) {
        if (points.length === 0) return {};
        const last = points[points.length - 1];
        const lines: Record<string, any> = {};

        const addLine = (
            key: string,
            value: number | null,
            color: string,
            dash: number[],
        ) => {
            if (value == null) return;
            lines[key] = {
                type: "line",
                yMin: value,
                yMax: value,
                borderColor: color,
                borderWidth: 1.5,
                borderDash: dash,
                label: { display: false },
            };
        };
        addLine("mean", last.mean, "#888", []);
        addLine("ucl", last.ucl, "#e53e3e", []);
        addLine("lcl", last.lcl, "#e53e3e", []);
        addLine("uwl", last.uwl, "#dd6b20", [4, 4]);
        addLine("lwl", last.lwl, "#dd6b20", [4, 4]);
        addLine("uil", last.uil, "#d69e2e", [2, 4]);
        addLine("lil", last.lil, "#d69e2e", [2, 4]);

        points.forEach((p, i) => {
            if (p.value <= yMax && p.value >= yMin) return;
            const isHigh = p.value > yMax;
            lines[`outlier-${i}`] = {
                type: "label",
                xValue: String(p.sequence_number),
                yValue: isHigh ? yMax : yMin,
                content: [`${isHigh ? "▲" : "▼"} ${sigFigs(p.value)}`],
                font: { family: "var(--font-mono)", size: 10 },
                color: "var(--colour-danger)",
                backgroundColor: "rgba(255,255,255,0.85)",
                borderColor: "var(--colour-danger)",
                borderWidth: 1,
                borderRadius: 3,
                padding: 4,
                z: 100,
                yAdjust: isHigh ? 16 : -16,
            };
        });

        return lines;
    }

    function xChartOptions(points: ChartPoint[]) {
        if (points.length === 0) return {};

        const ucl = points.find((p) => p.ucl != null)?.ucl ?? null;
        const lcl = points.find((p) => p.lcl != null)?.lcl ?? null;

        let yMin: number;
        let yMax: number;

        if (ucl != null && lcl != null) {
            const range = ucl - lcl;
            yMin = lcl - range * 0.3;
            yMax = ucl + range * 0.3;
        } else {
            const values = points.map((p) => p.value);
            const min = Math.min(...values);
            const max = Math.max(...values);
            const pad = (max - min) * 0.2 || 1;
            yMin = min - pad;
            yMax = max + pad;
        }

        return {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 },
            plugins: {
                legend: { display: false },
                annotation: {
                    annotations: annotationsForX(points, yMin, yMax),
                },
            },
            scales: {
                x: {
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
                y: {
                    min: yMin,
                    max: yMax,
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
            },
        };
    }

    function mrChartOptions(ucl: number | null, points: ChartPoint[]) {
        const annotations: Record<string, any> = {};

        const first = points.find((p) => p.ucl != null && p.lcl != null);
        const yMax = first
            ? (first.ucl! - first.lcl!) * 1.3
            : ucl != null
              ? ucl * 1.3
              : undefined;

        if (ucl != null) {
            annotations["ucl"] = {
                type: "line",
                yMin: ucl,
                yMax: ucl,
                borderColor: "#e53e3e",
                borderWidth: 1.5,
                borderDash: [],
                label: { display: false },
            };
        }

        if (yMax != null) {
            const mrValues = points.map((p, i) =>
                i === 0 ? null : Math.abs(p.value - points[i - 1].value),
            );
            mrValues.forEach((mr, i) => {
                if (mr == null || mr <= yMax) return;
                annotations[`outlier-${i}`] = {
                    type: "label",
                    xValue: String(points[i].sequence_number),
                    yValue: yMax,
                    content: [`▲ ${sigFigs(mr)}`],
                    font: { family: "var(--font-mono)", size: 10 },
                    color: "var(--colour-danger)",
                    backgroundColor: "rgba(255,255,255,0.85)",
                    borderColor: "var(--colour-danger)",
                    borderWidth: 1,
                    borderRadius: 3,
                    padding: 4,
                    yAdjust: 16,
                    z: 100,
                };
            });
        }

        return {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 },
            plugins: {
                legend: { display: false },
                annotation: { annotations },
            },
            scales: {
                x: {
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
                y: {
                    min: 0,
                    ...(yMax != null ? { max: yMax } : {}),
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
            },
        };
    }
</script>

<!-- ── combo picker ───────────────────────────────────────────────────── -->
<div class="picker-bar">
    <div class="combo-list">
        {#each combos as combo}
            <button
                class="combo-card"
                class:active={selectedMethodID === combo.methodID &&
                    selectedMaterialID === combo.materialID}
                on:click={() => selectCombo(combo)}
            >
                <span class="combo-method">{combo.methodName}</span>
                <span class="combo-material">{combo.materialName}</span>
            </button>
        {/each}
    </div>
</div>

<!-- ── chart controls ─────────────────────────────────────────────────── -->
{#if selectedMethodID && selectedMaterialID}
    <div class="picker-bar secondary">
        <label class="limit-label">
            Points
            <input
                type="number"
                min="0"
                bind:value={pointLimit}
                class="limit-input"
            />
        </label>
        <label class="limit-label">
            Per row
            <select bind:value={chartsPerRow} class="limit-input">
                <option value={1}>1</option>
                <option value={2}>2</option>
                <option value={3}>3</option>
                <option value={4}>4</option>
                <option value={5}>5</option>
            </select>
        </label>
        <button class="btn-primary" disabled={loading} on:click={loadCombo}>
            {loading ? "Loading…" : "Reload"}
        </button>
        {#if analytes.length > 0}
            <button
                class="btn-secondary"
                on:click={() => (showRawData = !showRawData)}
            >
                {showRawData ? "Show charts" : "Show data"}
            </button>
        {/if}
    </div>
{/if}

{#if error}
    <p class="error">{error}</p>
{/if}

<!-- ── charts ─────────────────────────────────────────────────────────── -->
{#if !showRawData}
    <div class="charts-area" style="--per-row: {chartsPerRow}">
        {#each analytes as analyte (analyte.mma_id)}
            {@const points = chartData[String(analyte.mma_id)] ?? []}
            <section class="analyte-section">
                <h2 class="analyte-title">
                    {analyte.name} <span class="unit">({analyte.unit})</span>
                </h2>
                {#if points.length === 0}
                    <p class="no-data">No data</p>
                {:else}
                    <div class="chart-pair">
                        <div class="chart-wrap">
                            <span class="chart-label">X (Individuals)</span>
                            <canvas id="chart-x-{analyte.mma_id}" height="200"
                            ></canvas>
                        </div>
                        <div class="chart-wrap mr">
                            <span class="chart-label">mR (Moving Range)</span>
                            <canvas id="chart-mr-{analyte.mma_id}" height="120"
                            ></canvas>
                        </div>
                    </div>
                {/if}
            </section>
        {/each}
    </div>
{/if}

<!-- ── raw data ───────────────────────────────────────────────────────── -->
{#if showRawData && analytes.length > 0}
    <div class="raw-data-area">
        <table class="raw-table">
            <thead>
                <tr>
                    <th>#</th>
                    {#each analytes as analyte}
                        <th>{analyte.name} ({analyte.unit})</th>
                        <th>Mean</th>
                        <th>UCL</th>
                        <th>LCL</th>
                    {/each}
                </tr>
            </thead>
            <tbody>
                {#each Array.from( { length: Math.max(...analytes.map((a) => (chartData[String(a.mma_id)] ?? []).length)) }, ) as _, i}
                    <tr>
                        <td>{i + 1}</td>
                        {#each analytes as analyte}
                            {@const p = (chartData[String(analyte.mma_id)] ??
                                [])[i]}
                            <td>{p != null ? sigFigs(p.value) : "—"}</td>
                            <td>{p?.mean != null ? sigFigs(p.mean) : "—"}</td>
                            <td>{p?.ucl != null ? sigFigs(p.ucl) : "—"}</td>
                            <td>{p?.lcl != null ? sigFigs(p.lcl) : "—"}</td>
                        {/each}
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
{/if}

<style>
    /* ── picker bars ── */
    .picker-bar {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 1rem;
        background: var(--colour-surface);
        border-bottom: 1px solid var(--colour-border);
        flex-wrap: wrap;
    }
    .picker-bar.secondary {
        background: var(--colour-bg);
    }

    .combo-list {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }
    .combo-card {
        display: flex;
        flex-direction: column;
        padding: 0.4rem 0.75rem;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        background: var(--colour-bg);
        cursor: pointer;
        text-align: left;
        font-family: var(--font-sans);
        gap: 0.1rem;
    }
    .combo-card:hover {
        border-color: var(--colour-primary);
    }
    .combo-card.active {
        border-color: var(--colour-primary);
        background: var(--colour-primary);
        color: white;
    }
    .combo-card.active .combo-material {
        color: rgba(255, 255, 255, 0.8);
    }
    .combo-method {
        font-size: 0.8rem;
        font-weight: 600;
    }
    .combo-material {
        font-size: 0.75rem;
        color: var(--colour-text-muted);
    }

    select {
        font-family: var(--font-sans);
        font-size: 0.9rem;
        padding: 0.35rem 0.6rem;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        background: var(--colour-bg);
        color: var(--colour-text);
        min-width: 60px;
    }
    .limit-label {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        font-size: 0.875rem;
        color: var(--colour-text-muted);
    }
    .limit-input {
        width: 60px;
        font-family: var(--font-mono);
        font-size: 0.875rem;
        padding: 0.3rem 0.5rem;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        background: var(--colour-bg);
        color: var(--colour-text);
    }

    /* ── buttons ── */
    .btn-primary {
        padding: 0.35rem 1rem;
        background: var(--colour-primary);
        color: white;
        border: none;
        border-radius: var(--radius);
        cursor: pointer;
        font-family: var(--font-sans);
        font-size: 0.9rem;
    }
    .btn-primary:hover:not(:disabled) {
        background: var(--colour-primary-hover);
    }
    .btn-primary:disabled {
        opacity: 0.5;
        cursor: default;
    }

    .btn-secondary {
        padding: 0.35rem 1rem;
        background: transparent;
        color: var(--colour-primary);
        border: 1px solid var(--colour-primary);
        border-radius: var(--radius);
        cursor: pointer;
        font-family: var(--font-sans);
        font-size: 0.9rem;
    }
    .btn-secondary:hover {
        background: var(--colour-primary);
        color: white;
    }

    /* ── error ── */
    .error {
        color: var(--colour-danger);
        padding: 0.75rem 1rem;
        font-size: 0.9rem;
    }

    /* ── charts ── */
    .charts-area {
        padding: 1rem;
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        gap: 1rem;
    }
    .analyte-section {
        flex: 0 0 calc((100% - (var(--per-row) - 1) * 1rem) / var(--per-row));
        min-width: 0;
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius-lg);
        padding: 1rem 1.25rem;
    }
    .analyte-title {
        font-size: 1rem;
        font-weight: 600;
        margin: 0 0 0.75rem;
    }
    .unit {
        font-weight: 400;
        color: var(--colour-text-muted);
        font-size: 0.875rem;
    }
    .chart-pair {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .chart-wrap {
        height: 200px;
        position: relative;
    }
    .chart-wrap.mr {
        height: 120px;
    }
    .chart-label {
        font-size: 0.75rem;
        color: var(--colour-text-muted);
        font-family: var(--font-mono);
        display: block;
        margin-bottom: 0.25rem;
    }
    .no-data {
        color: var(--colour-text-muted);
        font-size: 0.875rem;
        font-style: italic;
    }

    /* ── raw data ── */
    .raw-data-area {
        padding: 1rem;
        overflow-x: auto;
    }
    .raw-table {
        width: 100%;
        border-collapse: collapse;
        font-family: var(--font-mono);
        font-size: 0.8rem;
    }
    .raw-table th,
    .raw-table td {
        text-align: right;
        padding: 0.25rem 0.5rem;
        border-bottom: 1px solid var(--colour-border);
    }
    .raw-table th {
        color: var(--colour-text-muted);
        font-weight: 600;
    }
    .raw-table tr:last-child td {
        border-bottom: none;
    }
</style>
