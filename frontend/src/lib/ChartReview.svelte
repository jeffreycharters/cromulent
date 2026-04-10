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
    import {
        GetCommentsForCombo,
        AddComment,
    } from "../../wailsjs/go/handlers/CommentsHandler";

    import {
        computeViolations,
        worstViolation,
        VIOLATION_LABELS,
    } from "./violations";
    import type { ViolationMap } from "./violations";

    import {
        GetRawDataColWidth,
        SetRawDataColWidth,
    } from "../../wailsjs/go/handlers/ConfigHandler";

    import { GetEffectiveRuleSetForCombo } from "../../wailsjs/go/handlers/SPCRuleSetHandler";
    import type { RuleSet } from "./violations";

    let violationsByAnalyte: Record<number, ViolationMap> = {};

    let rawDataColWidth = 80;
    let colWidthDebounce: ReturnType<typeof setTimeout>;

    let ruleSetsByMMA: Record<number, RuleSet[]> = {};

    // --- custom plugin: draws limit lines without annotation plugin hit detection ---

    const limitLinesPlugin = {
        id: "limitLines",
        afterDatasetsDraw(chart: any) {
            const { ctx, chartArea, scales } = chart;
            const points: ChartPoint[] =
                (chart.options.plugins?.limitLines as any)?.points ?? [];
            if (!points.length) return;

            const limitKeys = [
                "mean",
                "ucl",
                "lcl",
                "uwl",
                "lwl",
                "uil",
                "lil",
            ] as const;
            const lineStyles: Record<
                string,
                { color: string; dash: number[] }
            > = {
                mean: { color: "#888", dash: [] },
                ucl: { color: "#e53e3e", dash: [] },
                lcl: { color: "#e53e3e", dash: [] },
                uwl: { color: "#dd6b20", dash: [4, 4] },
                lwl: { color: "#dd6b20", dash: [4, 4] },
                uil: { color: "#d69e2e", dash: [2, 4] },
                lil: { color: "#d69e2e", dash: [2, 4] },
            };

            ctx.save();

            for (const key of limitKeys) {
                const style = lineStyles[key];
                let segStart = 0;

                for (let i = 1; i <= points.length; i++) {
                    const prevVal = (points[i - 1] as any)[key];
                    const currVal =
                        i < points.length ? (points[i] as any)[key] : null;

                    // end segment when value changes or we reach the end
                    if (i === points.length || currVal !== prevVal) {
                        if (prevVal == null) {
                            segStart = i;
                            continue;
                        }

                        const y = scales.y.getPixelForValue(prevVal);
                        if (y < chartArea.top || y > chartArea.bottom) {
                            segStart = i;
                            continue;
                        }

                        const xStart =
                            segStart === 0
                                ? chartArea.left
                                : scales.x.getPixelForValue(segStart);
                        const xEnd =
                            i === points.length
                                ? chartArea.right
                                : scales.x.getPixelForValue(i);

                        ctx.beginPath();
                        ctx.moveTo(xStart, y);
                        ctx.lineTo(xEnd, y);
                        ctx.strokeStyle = style.color;
                        ctx.lineWidth = 1.5;
                        ctx.setLineDash(style.dash);
                        ctx.stroke();

                        // connect to next segment with a vertical line
                        if (i < points.length) {
                            const nextVal = (points[i] as any)[key];
                            if (nextVal != null) {
                                const yNext =
                                    scales.y.getPixelForValue(nextVal);
                                const xBoundary = scales.x.getPixelForValue(i);
                                ctx.beginPath();
                                ctx.moveTo(xBoundary, y);
                                ctx.lineTo(xBoundary, yNext);
                                ctx.strokeStyle = style.color;
                                ctx.lineWidth = 1;
                                ctx.setLineDash([2, 2]);
                                ctx.stroke();
                            }
                        }

                        segStart = i;
                    }
                }
            }

            ctx.setLineDash([]);
            ctx.restore();
        },
    };

    Chart.register(...registerables, annotationPlugin, limitLinesPlugin);

    // --- types ---

    type ComboAnalyte = models.ComboAnalyte;
    type CommentResponse = models.CommentResponse;

    interface ChartPoint {
        measurement_id: number;
        control_chart_id: number;
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
        methodMaterialID: number;
        methodName: string;
        materialName: string;
    }

    // --- state ---

    let combos: ComboOption[] = [];
    let selectedMethodMaterialID: number | null = null;

    let analytes: ComboAnalyte[] = [];
    let chartData: Record<string, ChartPoint[]> = {};
    let chartInstances: Record<string, Chart> = {};

    let pointLimit = 50;
    let chartsPerRow = 3;
    let loading = false;
    let error = "";
    let showRawData = false;
    let showOutliers = false;

    export let currentUser: any;

    let comments: CommentResponse[] = [];
    let commentsByMeasurement: Record<number, CommentResponse[]> = {};
    let commentsByChart: Record<number, CommentResponse[]> = {};

    let modalOpen = false;
    let modalPoint: ChartPoint | null = null;
    let modalComment = "";
    let savingComment = false;

    // --- lifecycle ---

    onMount(async () => {
        const methods = (await ListMethodsWithMaterials()) ?? [];
        combos = methods.flatMap((m) =>
            (m.materials ?? []).map((mat) => ({
                methodMaterialID: mat.method_material_id,
                methodName: m.name,
                materialName: mat.name,
            })),
        );
        rawDataColWidth = await GetRawDataColWidth();
    });

    onDestroy(() => {
        destroyCharts();
    });

    $: if (!loading && !showRawData && analytes.length > 0) {
        import("svelte").then(({ tick }) => tick().then(() => buildCharts()));
    }

    $: chartableAnalytes = analytes.filter((a) => {
        const pts = chartData[String(a.mma_id)] ?? [];
        return pts.some((p) => p.ucl != null);
    });

    function onColWidthChange() {
        clearTimeout(colWidthDebounce);
        colWidthDebounce = setTimeout(() => {
            SetRawDataColWidth(rawDataColWidth);
        }, 500);
    }

    // --- actions ---

    async function selectCombo(combo: ComboOption) {
      if (!selectedMethodMaterialID) return;
        await loadCombo();
    }

    async function loadCombo() {
      if (!selectedMethodMaterialID) return;
        loading = true;
        error = "";
        destroyCharts();
        chartData = {};
        analytes = [];
        comments = [];
        try {
          const [ana, data, cmts] = await Promise.all([
                 GetAnalytesForCombo(selectedMethodMaterialID),
                 GetComboChartData(selectedMethodMaterialID, pointLimit),
                 GetCommentsForCombo(selectedMethodMaterialID),
             ]);
            analytes = (ana ?? []).sort(
                (a, b) => a.display_order - b.display_order,
            );

            ruleSetsByMMA = {};
            for (const analyte of analytes) {
                const pts = (data ?? {})[String(analyte.mma_id)] ?? [];
                const maxSeq =
                    pts.length > 0
                        ? Math.max(...pts.map((p: any) => p.sequence_number))
                        : 0;
                const rs = await GetEffectiveRuleSetForCombo(
                    analyte.mma_id,
                    maxSeq,
                );
                ruleSetsByMMA[analyte.mma_id] = rs
                    ? [
                          {
                              effectiveFromSequence:
                                  rs.effectiveFromSequence ?? null,
                              beyondLimitsEnabled: rs.beyondLimitsEnabled,
                              warningLimitsEnabled: rs.warningLimitsEnabled,
                              warningConsecutiveCount:
                                  rs.warningConsecutiveCount,
                              warningTriggerCount: rs.warningTriggerCount,
                              trendEnabled: rs.trendEnabled,
                              trendConsecutiveCount: rs.trendConsecutiveCount,
                              oneSideEnabled: rs.oneSideEnabled,
                              oneSideConsecutiveCount:
                                  rs.oneSideConsecutiveCount,
                          },
                      ]
                    : [];
            }
            chartData = data ?? {};
            comments = cmts ?? [];
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
        for (const analyte of chartableAnalytes) {
            const points: ChartPoint[] =
                chartData[String(analyte.mma_id)] ?? [];
            if (points.length === 0) continue;

            const labels = points.map((p) => String(p.sequence_number));
            while (labels.length < 20) labels.push("");

            const mrValues = points.map((p, i) =>
                i === 0 ? null : Math.abs(p.value - points[i - 1].value),
            );

            const violations = violationsByAnalyte[analyte.mma_id];

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
                                tension: 0,
                                spanGaps: false,
                                pointBackgroundColor: points.map((p) => {
                                    const codes =
                                        violations?.get(p.measurement_id) ?? [];
                                    const worst = worstViolation(codes);
                                    if (worst === "OOC") return "#e53e3e";
                                    if (worst === "WRN") return "#dd6b20";
                                    if (
                                        commentsByMeasurement[p.measurement_id]
                                            ?.length > 0
                                    )
                                        return "#d69e2e";
                                    return "var(--colour-primary)";
                                }),
                                pointRadius: points.map((p) =>
                                    commentsByMeasurement[p.measurement_id]
                                        ?.length > 0
                                        ? 5
                                        : 3,
                                ),
                                pointStyle: points.map((p) => {
                                    const codes =
                                        violations?.get(p.measurement_id) ?? [];
                                    return codes.includes("WRN")
                                        ? "triangle"
                                        : "circle";
                                }),
                            },
                        ],
                    },
                    options: xChartOptions(points),
                });

                xCanvas.addEventListener("click", (e) => {
                    const chart = chartInstances[`${analyte.mma_id}-x`];
                    if (!chart) return;
                    const elements = chart.getElementsAtEventForMode(
                        e,
                        "nearest",
                        { intersect: false },
                        false,
                    );
                    if (elements.length === 0) return;
                    const idx = elements[0].index;
                    if (idx < points.length) openModal(points[idx]);
                });
            }

            const mrCanvas = document.getElementById(
                `chart-mr-${analyte.mma_id}`,
            ) as HTMLCanvasElement | null;

            if (mrCanvas) {
                const mrUcl = mRucl(points);
                // Synthesise a points-shaped array carrying only the mR UCL so
                // limitLinesPlugin can draw it using the same segment logic.
                const mrPoints = points.map((p) => ({
                    ...p,
                    mean: null,
                    ucl: mrUcl,
                    lcl: null,
                    uwl: null,
                    lwl: null,
                    uil: null,
                    lil: null,
                }));

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
                    options: mrChartOptions(mrUcl, points, mrPoints),
                });
            }
        }
    }

    // --- helpers ---

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

    // --- chart options ---

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

        // outlier label annotations — only shown when clamped
        const annotations: Record<string, any> = {};
        if (!showOutliers) {
            points.forEach((p, i) => {
                if (p.value <= yMax && p.value >= yMin) return;
                const isHigh = p.value > yMax;
                annotations[`outlier-${i}`] = {
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
        }

        return {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 },
            plugins: {
                legend: { display: false },
                limitLines: { points } as any,
                annotation: { annotations },
            },
            scales: {
                x: {
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
                y: {
                    ...(showOutliers ? {} : { min: yMin, max: yMax }),
                    ticks: { font: { family: "var(--font-mono)", size: 11 } },
                },
            },
            interaction: {
                mode: "nearest" as const,
                intersect: false,
            },
        };
    }

    function mrChartOptions(
        ucl: number | null,
        points: ChartPoint[],
        mrPoints: Omit<ChartPoint, "value">[],
    ) {
        const annotations: Record<string, any> = {};

        const first = points.find((p) => p.ucl != null && p.lcl != null);
        const yMax = first
            ? (first.ucl! - first.lcl!) * 1.3
            : ucl != null
              ? ucl * 1.3
              : undefined;

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
                limitLines: { points: mrPoints } as any,
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

    // --- modal ---

    let modalAnalyteName: string | null = null;
    let modalAnalyteUnit: string | null = null;

    function openModal(
        point: ChartPoint,
        analyteName?: string,
        analyteUnit?: string,
    ) {
        modalPoint = point;
        modalAnalyteName = analyteName ?? null;
        modalAnalyteUnit = analyteUnit ?? null;
        modalComment = "";
        modalOpen = true;
    }

    function closeModal() {
        modalOpen = false;
        modalPoint = null;
        modalComment = "";
    }

    async function submitComment() {
        if (!modalPoint || !modalComment.trim()) return;
        savingComment = true;
        try {
            await AddComment(
                modalPoint.control_chart_id,
                modalPoint.measurement_id,
                modalComment.trim(),
                currentUser.id,
            );
            comments = (await GetCommentsForCombo(selectedMethodMaterialID!)) ?? [];
            modalComment = "";
        } catch (e: any) {
            error = e?.toString() ?? "Failed to save comment";
        } finally {
            savingComment = false;
        }
    }

    function limitContext(point: ChartPoint): string {
        const { value, ucl, lcl, uwl, lwl, uil, lil, mean } = point;
        if (ucl != null && value > ucl) return "Above UCL";
        if (lcl != null && value < lcl) return "Below LCL";
        if (uwl != null && value > uwl) return "Above UWL";
        if (lwl != null && value < lwl) return "Below LWL";
        if (uil != null && value > uil) return "Above UIL";
        if (lil != null && value < lil) return "Below LIL";
        if (mean != null && value > mean) return "Above mean";
        if (mean != null && value < mean) return "Below mean";
        return "At mean";
    }

    // --- reactive: index comments ---

    $: {
        commentsByMeasurement = {};
        commentsByChart = {};
        for (const c of comments) {
            if (c.measurement_id != null) {
                commentsByMeasurement[c.measurement_id] ??= [];
                commentsByMeasurement[c.measurement_id].push(c);
            } else {
                commentsByChart[c.control_chart_id] ??= [];
                commentsByChart[c.control_chart_id].push(c);
            }
        }
    }

    $: {
        violationsByAnalyte = {};
        for (const analyte of analytes) {
            const points = chartData[String(analyte.mma_id)] ?? [];
            const ruleSets = ruleSetsByMMA[analyte.mma_id] ?? [];
            violationsByAnalyte[analyte.mma_id] = computeViolations(
                points,
                ruleSets,
            );
        }
    }

    $: sequenceNumbers = Array.from(
        new Set(
            analytes.flatMap((a) =>
                (chartData[String(a.mma_id)] ?? []).map(
                    (p) => p.sequence_number,
                ),
            ),
        ),
    ).sort((a, b) => a - b);
</script>

<!-- ── combo picker ───────────────────────────────────────────────────── -->
<div class="picker-bar">
    <div class="combo-list">
        {#each combos as combo}
            <button
                class="combo-card"
                class:active={selectedMethodMaterialID === combo.methodMaterialID}
                on:click={() => selectCombo(combo)}
            >
                <span class="combo-method">{combo.methodName}</span>
                <span class="combo-material">{combo.materialName}</span>
            </button>
        {/each}
    </div>
</div>

<!-- ── chart controls ─────────────────────────────────────────────────── -->
{#if selectedMethodMaterialID}
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
                on:click={() => {
                    showOutliers = !showOutliers;
                    buildCharts();
                }}
            >
                {showOutliers ? "Clamp chart" : "Show outliers"}
            </button>
            <button
                class="btn-secondary"
                on:click={() => (showRawData = !showRawData)}
            >
                {showRawData ? "Show charts" : "Show data"}
            </button>
            {#if showRawData}
                <label class="limit-label">
                    Col width
                    <input
                        type="range"
                        min="40"
                        max="160"
                        step="10"
                        bind:value={rawDataColWidth}
                        on:input={onColWidthChange}
                        class="col-width-slider"
                    />
                    <span class="limit-label">{rawDataColWidth}px</span>
                </label>
            {/if}
        {/if}
    </div>
{/if}

{#if error}
    <p class="error">{error}</p>
{/if}

<!-- ── charts ─────────────────────────────────────────────────────────── -->
{#if !showRawData}
    <div class="charts-area" style="--per-row: {chartsPerRow}">
        {#each chartableAnalytes as analyte (analyte.mma_id)}
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
        <table class="raw-table" style="--col-w: {rawDataColWidth}px">
            <thead>
                <tr>
                    <th class="seq-col">#</th>
                    {#each analytes as analyte}
                        <th>
                            <span class="th-name">{analyte.name}</span>
                            <span class="th-unit">{analyte.unit}</span>
                        </th>
                    {/each}
                </tr>
            </thead>
            <tbody>
                {#each sequenceNumbers as seq}
                    <tr>
                        <td class="seq-col">{seq}</td>
                        {#each analytes as analyte}
                            {@const points =
                                chartData[String(analyte.mma_id)] ?? []}
                            {@const p = points.find(
                                (pt) => pt.sequence_number === seq,
                            )}
                            {@const violations = p
                                ? (violationsByAnalyte[analyte.mma_id]?.get(
                                      p.measurement_id,
                                  ) ?? [])
                                : []}
                            {@const worst = worstViolation(violations)}
                            <td
                                class="value-cell"
                                class:viol-ooc={worst === "OOC"}
                                class:viol-wrn={worst === "WRN"}
                                class:viol-trd={worst === "TRD"}
                                class:viol-run={worst === "RUN"}
                                title={p != null
                                    ? [
                                          p.mean != null
                                              ? `Mean: ${sigFigs(p.mean)}`
                                              : null,
                                          p.ucl != null
                                              ? `UCL: ${sigFigs(p.ucl)}`
                                              : null,
                                          p.lcl != null
                                              ? `LCL: ${sigFigs(p.lcl)}`
                                              : null,
                                          p.uwl != null
                                              ? `UWL: ${sigFigs(p.uwl)}`
                                              : null,
                                          p.lwl != null
                                              ? `LWL: ${sigFigs(p.lwl)}`
                                              : null,
                                          p.uil != null
                                              ? `UIL: ${sigFigs(p.uil)}`
                                              : null,
                                          p.lil != null
                                              ? `LIL: ${sigFigs(p.lil)}`
                                              : null,
                                          ...violations.map(
                                              (v) => VIOLATION_LABELS[v],
                                          ),
                                      ]
                                          .filter(Boolean)
                                          .join("\n")
                                    : ""}
                                on:click={() =>
                                    p &&
                                    openModal(p, analyte.name, analyte.unit)}
                                role="button"
                                tabindex="0"
                                on:keydown={(e) =>
                                    e.key === "Enter" &&
                                    p &&
                                    openModal(p, analyte.name, analyte.unit)}
                            >
                                {#if p != null}
                                    <span class="cell-value"
                                        >{sigFigs(p.value)}</span
                                    >
                                    {#each violations as code}
                                        <span
                                            class="badge badge-{code.toLowerCase()}"
                                            >{code}</span
                                        >
                                    {/each}
                                {:else}
                                    <span class="cell-empty">—</span>
                                {/if}
                            </td>
                        {/each}
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
{/if}

<!-- ── comment modal ──────────────────────────────────────────────────── -->
{#if modalOpen && modalPoint}
    <div class="modal-backdrop" on:click={closeModal}>
        <div class="modal" on:click|stopPropagation>
            <div class="modal-header">
                <h3>
                    Sequence #{modalPoint.sequence_number}
                    {#if modalAnalyteName}
                        — {modalAnalyteName}{modalAnalyteUnit
                            ? ` (${modalAnalyteUnit})`
                            : ""}
                    {/if}
                </h3>
                <button class="modal-close" on:click={closeModal}>✕</button>
            </div>

            <div class="modal-meta">
                <span class="meta-value">{sigFigs(modalPoint.value)}</span>
                <span class="meta-context">{limitContext(modalPoint)}</span>
            </div>

            {#if modalPoint.ucl != null}
                <div class="modal-limits">
                    <span>UCL {sigFigs(modalPoint.ucl)}</span>
                    {#if modalPoint.uwl != null}<span
                            >UWL {sigFigs(modalPoint.uwl)}</span
                        >{/if}
                    {#if modalPoint.uil != null}<span
                            >UIL {sigFigs(modalPoint.uil)}</span
                        >{/if}
                    <span>Mean {sigFigs(modalPoint.mean ?? 0)}</span>
                    {#if modalPoint.lil != null}<span
                            >LIL {sigFigs(modalPoint.lil)}</span
                        >{/if}
                    {#if modalPoint.lwl != null}<span
                            >LWL {sigFigs(modalPoint.lwl)}</span
                        >{/if}
                    <span>LCL {sigFigs(modalPoint.lcl)}</span>
                </div>
            {/if}

            {#if (commentsByMeasurement[modalPoint.measurement_id] ?? []).length > 0}
                <div class="modal-comments">
                    <h4>Comments</h4>
                    {#each commentsByMeasurement[modalPoint.measurement_id] as c}
                        <div class="comment">
                            <span class="comment-meta"
                                >{c.username} · {new Date(
                                    c.created_at,
                                ).toLocaleString()}</span
                            >
                            <p class="comment-text">{c.text}</p>
                        </div>
                    {/each}
                </div>
            {/if}

            <div class="modal-new-comment">
                <textarea
                    class="comment-input"
                    bind:value={modalComment}
                    placeholder="Add a comment…"
                    rows="3"
                />
                <button
                    class="btn-primary"
                    disabled={!modalComment.trim() || savingComment}
                    on:click={submitComment}
                >
                    {savingComment ? "Saving…" : "Add Comment"}
                </button>
            </div>
        </div>
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

    /* ── controls ── */
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
        border-collapse: collapse;
        font-family: var(--font-mono);
        font-size: 0.8rem;
        text-align: center;
    }
    .raw-table th,
    .raw-table td {
        width: var(--col-w);
        min-width: var(--col-w);
        max-width: var(--col-w);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        text-align: right;
        padding: 0.2rem 0.4rem;
        border: 1px solid var(--colour-border);
        box-sizing: border-box;
    }
    .raw-table th {
        font-weight: 600;
        background: var(--colour-surface);
        position: sticky;
        top: 0;
        z-index: 1;
    }
    .col-width-slider {
        width: 80px;
        cursor: pointer;
    }
    .raw-table tr:last-child td {
        border-bottom: none;
    }

    /* ── modal ── */
    .modal-backdrop {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.4);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }
    .modal {
        background: var(--colour-surface);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius-lg);
        padding: 1.5rem;
        width: 420px;
        max-width: 90vw;
        display: flex;
        flex-direction: column;
        gap: 1rem;
        box-shadow: var(--shadow-md);
    }
    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .modal-header h3 {
        font-size: 1rem;
        font-weight: 700;
        margin: 0;
    }
    .modal-close {
        background: none;
        border: none;
        font-size: 1rem;
        cursor: pointer;
        color: var(--colour-text-muted);
        padding: 0.25rem;
    }
    .modal-close:hover {
        color: var(--colour-text);
    }
    .modal-meta {
        display: flex;
        align-items: baseline;
        gap: 0.75rem;
    }
    .meta-value {
        font-family: var(--font-mono);
        font-size: 1.5rem;
        font-weight: 700;
    }
    .meta-context {
        font-size: 0.875rem;
        color: var(--colour-text-muted);
    }
    .modal-limits {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
    }
    .modal-limits span {
        font-family: var(--font-mono);
        font-size: 0.75rem;
        background: var(--colour-bg);
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.2rem 0.5rem;
        color: var(--colour-text-muted);
    }
    .modal-comments h4 {
        font-size: 0.8rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: var(--colour-text-muted);
        margin: 0 0 0.5rem;
    }
    .comment {
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
        padding: 0.5rem 0;
        border-bottom: 1px solid var(--colour-border);
    }
    .comment:last-child {
        border-bottom: none;
    }
    .comment-meta {
        font-size: 0.75rem;
        color: var(--colour-text-muted);
        font-family: var(--font-mono);
    }
    .comment-text {
        font-size: 0.875rem;
        margin: 0;
    }
    .modal-new-comment {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .comment-input {
        width: 100%;
        border: 1px solid var(--colour-border);
        border-radius: var(--radius);
        padding: 0.5rem 0.75rem;
        font-family: var(--font-sans);
        font-size: 0.875rem;
        color: var(--colour-text);
        background: var(--colour-bg);
        resize: vertical;
        box-sizing: border-box;
    }
    .comment-input:focus {
        outline: none;
        border-color: var(--colour-primary);
    }
    /* ── violation cell colours ── */
    .value-cell {
        cursor: pointer;
        position: relative;
        white-space: nowrap;
    }
    .value-cell:hover {
        background: var(--colour-surface);
    }
    .viol-ooc {
        background: #fed7d7;
        color: #822727;
    }
    .viol-wrn {
        background: #feebc8;
        color: #7b341e;
    }
    .viol-trd {
        background: #fefcbf;
        color: #744210;
    }
    .viol-run {
        background: #fefcbf;
        color: #744210;
    }

    /* ── badges ── */
    .badge {
        display: inline-block;
        font-size: 0.6rem;
        font-weight: 700;
        padding: 0.1rem 0.3rem;
        border-radius: 3px;
        margin-left: 0.25rem;
        vertical-align: middle;
        font-family: var(--font-mono);
    }
    .badge-ooc {
        background: #e53e3e;
        color: white;
    }
    .badge-wrn {
        background: #dd6b20;
        color: white;
    }
    .badge-trd {
        background: #d69e2e;
        color: white;
    }
    .badge-run {
        background: #d69e2e;
        color: white;
    }

    /* ── seq column ── */
    .seq-col {
        position: sticky;
        left: 0;
        z-index: 2;
        background: var(--colour-surface);
        text-align: left;
        color: var(--colour-text-muted);
    }
    .raw-table th.seq-col {
        z-index: 3; /* above both sticky axes */
    }

    /* ── empty cell ── */
    .cell-empty {
        color: var(--colour-text-muted);
    }

    .th-name {
        display: block;
        font-weight: 600;
    }
    .th-unit {
        display: block;
        font-weight: 400;
        color: var(--colour-text-muted);
        font-size: 0.7rem;
    }
</style>
