<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Chart, registerables } from "chart.js";
    import type { models } from "../../wailsjs/go/models";

    type MethodWithMaterials = models.MethodWithMaterials;
    type ComboAnalyte = models.ComboAnalyte;

    import annotationPlugin from "chartjs-plugin-annotation";
    import {
        ListMethodsWithMaterials,
        GetAnalytesForCombo,
    } from "../../wailsjs/go/handlers/DataEntryHandler";
    import { GetComboChartData } from "../../wailsjs/go/handlers/ChartReviewHandler";

    Chart.register(...registerables, annotationPlugin);

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

    // --- state ---
    let methods: MethodWithMaterials[] = [];
    let selectedMethodID: number | null = null;
    let selectedMaterialID: number | null = null;
    let analytes: ComboAnalyte[] = [];
    let chartData: Record<string, ChartPoint[]> = {};
    let pointLimit = 50;
    let loading = false;
    let error = "";

    // chart instances: keyed by `${mmaid}-x` and `${mmaid}-mr`
    let chartInstances: Record<string, Chart> = {};

    onMount(async () => {
        methods = (await ListMethodsWithMaterials()) ?? [];
    });

    onDestroy(() => {
        destroyCharts();
    });

    function destroyCharts() {
        for (const c of Object.values(chartInstances)) c.destroy();
        chartInstances = {};
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

    // After DOM updates, build/rebuild charts
    $: if (!loading && analytes.length > 0) {
        // tick needed so canvas elements exist
        import("svelte").then(({ tick }) => tick().then(() => buildCharts()));
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

            console.log("points for", analyte.mma_id, points, chartData);

            // --- X chart ---
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

            // --- mR chart ---
            const mrCanvas = document.getElementById(
                `chart-mr-${analyte.mma_id}`,
            ) as HTMLCanvasElement | null;
            if (mrCanvas) {
                const mrUCL = mRucl(points);
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
                    options: mrChartOptions(mrUCL),
                });
            }
        }
    }

    function annotationsForX(points: ChartPoint[], yMin: number, yMax: number) {
        if (points.length === 0) return {};
        const last = points[points.length - 1];
        const lines: Record<string, any> = {};
        const add = (
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
        add("mean", last.mean, "#888", []);
        add("ucl", last.ucl, "#e53e3e", []);
        add("lcl", last.lcl, "#e53e3e", []);
        add("uwl", last.uwl, "#dd6b20", [4, 4]);
        add("lwl", last.lwl, "#dd6b20", [4, 4]);
        add("uil", last.uil, "#d69e2e", [2, 4]);
        add("lil", last.lil, "#d69e2e", [2, 4]);

        points.forEach((p, i) => {
            if (p.value > yMax || p.value < yMin) {
                const isHigh = p.value > yMax;
                lines[`outlier-${i}`] = {
                    type: "label",
                    xValue: String(p.sequence_number),
                    yValue: isHigh ? yMax : yMin,
                    content: [`${isHigh ? "▲" : "▼"} ${p.value.toFixed(3)}`],
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
            }
        });

        return lines;
    }

    function xChartOptions(points: ChartPoint[]) {
        if (points.length === 0) return {};

        const ucl = points.find((p) => p.ucl != null)?.ucl ?? null;
        const lcl = points.find((p) => p.lcl != null)?.lcl ?? null;
        const values = points.map((p) => p.value);

        let yMin: number;
        let yMax: number;

        if (ucl != null && lcl != null) {
            const range = ucl - lcl;
            yMin = lcl - range * 0.3;
            yMax = ucl + range * 0.3;
        } else {
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

    function mRucl(points: ChartPoint[]): number | null {
        // Standard XmR: UCL_mR = 3.267 * mean(mR)
        const mrs = points
            .slice(1)
            .map((p, i) => Math.abs(p.value - points[i].value));
        if (mrs.length === 0) return null;
        const meanMR = mrs.reduce((a, b) => a + b, 0) / mrs.length;
        return 3.267 * meanMR;
    }

    function mrChartOptions(ucl: number | null) {
        const annotations: Record<string, any> = {};
        if (ucl != null) {
            annotations["ucl"] = {
                type: "line",
                yMin: ucl,
                yMax: ucl,
                borderColor: "#e53e3e",
                borderWidth: 1.5,
                borderDash: [],
            };
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
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
            },
        };
    }
</script>

<!-- ── picker bar ─────────────────────────────────────────────────────── -->
<div class="picker-bar">
    <select
        bind:value={selectedMethodID}
        on:change={() => {
            selectedMaterialID = null;
            analytes = [];
            chartData = {};
        }}
    >
        <option value={null}>Select method…</option>
        {#each methods as m}
            <option value={m.id}>{m.name}</option>
        {/each}
    </select>

    {#if selectedMethodID}
        {@const mats =
            methods.find((m) => m.id === selectedMethodID)?.materials ?? []}
        <select bind:value={selectedMaterialID}>
            <option value={null}>Select material…</option>
            {#each mats as mat}
                <option value={mat.id}>{mat.name}</option>
            {/each}
        </select>
    {/if}

    <label class="limit-label">
        Points
        <input
            type="number"
            min="0"
            bind:value={pointLimit}
            class="limit-input"
        />
    </label>

    <button
        class="btn-primary"
        disabled={!selectedMethodID || !selectedMaterialID || loading}
        on:click={loadCombo}
    >
        {loading ? "Loading…" : "Load"}
    </button>
</div>

{#if error}
    <p class="error">{error}</p>
{/if}

<!-- ── charts ─────────────────────────────────────────────────────────── -->
<div class="charts-area">
    {#each analytes as analyte (analyte.mma_id)}
        {@const points = chartData[analyte.mma_id] ?? []}
        <section class="analyte-section">
            <h2 class="analyte-title">
                {analyte.name} <span class="unit">({analyte.unit})</span>
            </h2>

            {#if points.length === 0}
                <p class="no-data">No data</p>
            {:else}
                <div class="chart-pair">
                    <div class="chart-wrap" style="height: 200px">
                        <span class="chart-label">X (Individuals)</span>
                        <canvas
                            id="chart-x-{analyte.mma_id}"
                            height="200"
                            width="600"
                        ></canvas>
                    </div>
                    <div class="chart-wrap mr">
                        <span class="chart-label">mR (Moving Range)</span>
                        <canvas
                            id="chart-mr-{analyte.mma_id}"
                            height="120"
                            width="600"
                        ></canvas>
                    </div>
                </div>
            {/if}
        </section>
    {/each}
</div>

<style>
    .picker-bar {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 1rem;
        background: var(--colour-surface);
        border-bottom: 1px solid var(--colour-border);
        flex-wrap: wrap;
    }

    select {
        font-family: var(--font-sans);
        font-size: 0.9rem;
        padding: 0.35rem 0.6rem;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        background: var(--colour-bg);
        color: var(--colour-text);
        min-width: 160px;
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

    .error {
        color: var(--colour-danger);
        padding: 0.75rem 1rem;
        font-size: 0.9rem;
    }

    .charts-area {
        padding: 1rem;
        display: flex;
        flex-direction: column;
        gap: 2rem;
    }

    .analyte-section {
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
</style>
