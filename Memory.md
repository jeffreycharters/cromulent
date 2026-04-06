Wanna develop an app?
2:19 p.m.

Cromulent! Let's go. Where do you want to pick up — fixing the Chart.js onClick, or something else?

Yeah let's do it to it.
2:20 p.m.

Let's see the current onClick setup in ChartReview.svelte. Can you paste the relevant section — specifically the xChartOptions definition and whatever click handler you have wired up?

xChartOptions(points: ChartPoint[])

and

 onClick: (_: any, elements: any[]) => {
                        if (elements.length === 0) return;
                        const idx = elements[0].index;
                        if (idx < points.length) openModal(points[idx]);
                    },
2:21 p.m.

That looks structurally fine. A few things that commonly kill Chart.js onClick:

    Options not being applied — if xChartOptions returns a new object but the chart is already instantiated, the onClick never gets registered unless the chart is destroyed and rebuilt.
    events array missing — Chart.js v3+ requires events: ['click'] (or includes it by default, but plugins can stomp it).
    The click is hitting the annotation layer, not the canvas — chartjs-plugin-annotation can intercept pointer events if annotations have click handlers or if interaction config is off.

Can you paste:

    How you're calling xChartOptions when building the chart (the new Chart(...) call)
    Your Chart.js + annotation plugin versions from package.json

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
                chart-x-${analyte.mma_id},
            ) as HTMLCanvasElement | null;
            if (xCanvas) {
                chartInstances[${analyte.mma_id}-x] = new Chart(xCanvas, {
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
                                pointBackgroundColor: points.map((p) =>
                                    commentsByMeasurement[p.measurement_id]?.length > 0
                                        ? "#d69e2e"
                                        : "var(--colour-primary)"
                                ),
                                pointRadius: points.map((p) =>
                                    commentsByMeasurement[p.measurement_id]?.length > 0 ? 5 : 3
                                ),
                            },
                        ],
                    },
                    options: xChartOptions(points),
                });
            }
            const mrCanvas = document.getElementById(
                chart-mr-${analyte.mma_id},
            ) as HTMLCanvasElement | null;
            if (mrCanvas) {
                chartInstances[${analyte.mma_id}-mr] = new Chart(mrCanvas, {
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

and

{
    "dependencies": {
        "chart.js": "^4.5.1",
        "chartjs-plugin-annotation": "^3.1.0",
        "svelte-dnd-action": "^0.9.69"
    }
}
2:23 p.m.

Those versions are fine together. The build looks correct too — options applied at construction time, so that's not it.

Most likely culprit: chartjs-plugin-annotation v3 intercepts click events on the canvas by default, and if your annotations don't have explicit click: null handlers, it can swallow the event before Chart.js processes it for data points.

Try adding this to your xChartOptions:
typescript

plugins: {
    annotation: {
        enter: undefined,
        click: undefined,  // don't let annotation plugin steal clicks
        annotations: { /* your existing annotations */ }
    }
}

Also, one thing worth confirming — paste your xChartOptions function. Specifically want to see the interaction and plugins config. If interaction.mode isn't set to 'nearest' or 'index', elements in the onClick can come back empty even when you click directly on a point.

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
                  enter: undefined,
                          click: undefined,
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
            onClick: (_: any, elements: any[]) => {
                        if (elements.length === 0) return;
                        const idx = elements[0].index;
                        if (idx < points.length) openModal(points[idx]);
                    },
        };
    }
2:26 p.m.

You already have enter: undefined, click: undefined in there — so that's not it.

The missing piece is almost certainly interaction. Without it, Chart.js defaults to mode: 'nearest' with intersect: true, which means you have to click exactly on the dot. Easy to miss.

Add this to the returned options object:
typescript

interaction: {
    mode: 'nearest' as const,
    intersect: false,
},

intersect: false means clicking anywhere along the vertical slice of a point registers it, not just the pixel-perfect dot. Should make onClick fire reliably.

If that still doesn't work, the next thing to check is whether openModal is actually being called but something inside it is failing silently — worth a console.log(points[idx]) at the top of onClick to confirm the handler is reaching it.

OK so it seems to be tracking the points better, but the modal just opens when I open the chart data only. So my modal is probably in the wrong part of the template?
2:28 p.m.

Yeah, classic — the modal is probably inside the {#if showRawData} block or some other conditional that's only true briefly on load, then hidden.

Paste the modal markup and the surrounding template structure so I can see where it's sitting.

That helped. Looks good now! Only concern now is that I can't comment when the point is off-figure. Can we make it so the annotation can be clicked?
2:30 p.m.

Makes sense — the outlier annotations are rendered as labels, not data points, so Chart.js onClick never sees them.

Paste your annotationsForX function so I can see how the outlier annotations are built. We'll add a click handler directly on each outlier annotation.

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
            lines[outlier-${i}] = {
                type: "label",
                xValue: String(p.sequence_number),
                yValue: isHigh ? yMax : yMin,
                content: [${isHigh ? "▲" : "▼"} ${sigFigs(p.value)}],
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
2:30 p.m.

Easy fix — add a click handler to each outlier label annotation, passing the point through:
typescript

lines[`outlier-${i}`] = {
    // ... existing properties ...
    click: ({ chart }: { chart: any }) => {
        openModal(p);
    },
};

The annotation plugin calls its own click with an event context object, not the Chart.js elements array, so you ignore the argument and just call openModal(p) directly — p is already in scope from the forEach.

You'll also want to add cursor: 'pointer' so it's obvious it's clickable:
typescript

cursor: 'pointer',

That's it — no changes needed to xChartOptions since you already have click: undefined at the plugin level, which only blocks clicks on line annotations. Label annotations with their own click handler override that.

Not working still?

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
            lines[outlier-${i}] = {
                type: "label",
                xValue: String(p.sequence_number),
                yValue: isHigh ? yMax : yMin,
                content: [${isHigh ? "▲" : "▼"} ${sigFigs(p.value)}],
                font: { family: "var(--font-mono)", size: 10 },
                color: "var(--colour-danger)",
                backgroundColor: "rgba(255,255,255,0.85)",
                borderColor: "var(--colour-danger)",
                borderWidth: 1,
                borderRadius: 3,
                padding: 4,
                z: 100,
                yAdjust: isHigh ? 16 : -16,
                cursor: 'pointer',
                click: ({ chart }: { chart: any }) => {
                       openModal(p);
                   },
            };
        });
        return lines;
    }
2:34 p.m.

The problem is that click: undefined in your plugin-level annotation config is nuking the per-annotation click handlers too. Change it to remove that line:
typescript

annotation: {
    enter: undefined,
    // remove click: undefined entirely
    annotations: annotationsForX(points, yMin, yMax),
},

The click: undefined at the plugin level overrides individual annotation click handlers. You don't need it there — the line annotations don't have click handlers so they won't interfere.

Still not working - doesn't show cursor as pointer, either.
2:49 p.m.

Let me look this up — annotation plugin click handling changed between versions and I want to give you the right API for 3.1.0.
Searched the web

Good find. In v1+, the events array was removed and event hooks are now supplied with a single context parameter containing chart and element properties.
Chart.js

So the hook name in v3 is click, not onClick — which is what you have. That part is right.

But look at this — events are managed using the beforeEvent hook which is only invoked for events generated inside the chart area. If an annotation is drawn partially outside the chart area, clicks on the outside portion won't fire.
GitHub

Your outlier labels are rendered at yMax or yMin with a yAdjust of ±16, which pushes them outside the chart area. That's why the click never fires.

Two options:

Option A — remove yAdjust and keep the label inside the chart area. The label sits right at the boundary but stays clickable.

Option B — add clip: false to the annotation to allow drawing outside, but per that issue this is a known limitation and clicks still won't work outside the area.

Realistically, Option A is the fix. Try removing yAdjust from the outlier annotations and see if the click fires. You may want to adjust yMin/yMax to leave a little headroom for the label instead.

The annotations are showing on the chart thought - do you have the +/- reversed?
2:50 p.m.

You're right, I do. yAdjust: 16 on a high outlier pushes the label down into the chart, and yAdjust: -16 on a low outlier pushes it up into the chart. So they're already inside the chart area.

Back to square one on that theory. Let me think about what's actually different between your setup and a working annotation click...

Can you add a console.log inside the click handler on the outlier annotation and tell me if it fires at all? Just want to confirm whether it's a hit-detection problem or the handler genuinely isn't being called.

Is the normal annotation just catching every click when it is visible?
2:52 p.m.

Oh — yes, very likely. Your line annotations (ucl, lcl, etc.) span the full width of the chart. Even though they're thin lines, the annotation plugin's hit detection has a tolerance area around them. If one of those lines is near where you're clicking, it catches the event first.

Those line annotations don't have a click handler, so the click gets swallowed silently.

Add a no-op click handler to each line annotation in addLine:
typescript

lines[key] = {
    type: "line",
    yMin: value,
    yMax: value,
    borderColor: color,
    borderWidth: 1.5,
    borderDash: dash,
    label: { display: false },
    click: () => {},  // don't swallow clicks
};

Actually wait — a no-op still swallows it. What you actually want is for the line annotation click to fall through to the canvas. Try returning false from the handler, which in some versions tells the plugin not to consume the event. Or just test first by temporarily removing all line annotations and see if the outlier label click works — that'll confirm this is the culprit.

If I remove the onClick returned from xChartOptions it works, but still no cursor:pointer.
2:57 p.m.

Good diagnostic. So the Chart.js onClick is still winning over the annotation click when both are present. The annotation fires when Chart.js onClick is out of the picture.

For the conflict — move the click handling out of xChartOptions entirely and handle it on the canvas directly. In buildCharts, after creating the chart instance, add a canvas click listener that checks annotation hits first, falls back to data point:

Actually, the cleaner fix is to keep onClick in the options but check if the click landed on an outlier annotation first, and if so bail out. But that requires querying annotation elements which is messy.

Simplest real fix: remove onClick from xChartOptions and instead attach a click listener to the canvas itself:
typescript

xCanvas.addEventListener('click', (e) => {
    const chart = chartInstances[`${analyte.mma_id}-x`];
    if (!chart) return;
    const elements = chart.getElementsAtEventForMode(e, 'nearest', { intersect: false }, false);
    if (elements.length === 0) return;
    const idx = elements[0].index;
    if (idx < points.length) openModal(points[idx]);
});

The annotation plugin handles its own clicks independently on the canvas, so these won't conflict.

For cursor: pointer — that's a separate issue. The annotation plugin sets it via enter/leave hooks. Add these to your outlier annotations:
typescript

enter: () => { xCanvas.style.cursor = 'pointer'; },
leave: () => { xCanvas.style.cursor = 'default'; },

But you don't have xCanvas in scope inside annotationsForX. You'll need to either pass the canvas in, or set cursor via the chart element reference in the hook context:
typescript

enter: ({ chart }: { chart: any }) => { chart.canvas.style.cursor = 'pointer'; },
leave: ({ chart }: { chart: any }) => { chart.canvas.style.cursor = 'default'; },

Can I just have a class and a style do this for me?
3:00 p.m.

No — the canvas is a single element, you can't target individual drawn annotations with CSS. The cursor style applies to the whole canvas, so it has to be set dynamically via the enter/leave hooks.

OK, better. Does the annotation have a margin the size of texas? It grabs everything now.
3:02 p.m.

Yeah, the hit detection tolerance on label annotations is generous by default. You can tighten it with hitTolerance:
typescript

hitTolerance: 0,

Add that to your outlier annotation config. 0 means it only registers a hit when the pointer is actually over the rendered pixels of the label, no padding around it.

Still is huge, have a look:
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
            lines[outlier-${i}] = {
                type: "label",
                xValue: String(p.sequence_number),
                yValue: isHigh ? yMax : yMin,
                content: [${isHigh ? "▲" : "▼"} ${sigFigs(p.value)}],
                font: { family: "var(--font-mono)", size: 10 },
                color: "var(--colour-danger)",
                backgroundColor: "rgba(255,255,255,0.85)",
                borderColor: "var(--colour-danger)",
                borderWidth: 1,
                borderRadius: 3,
                padding: 4,
                hitTolerance: 0,
                z: 100,
                yAdjust: isHigh ? 16 : -16,
                cursor: 'pointer',
                click: ({ chart }: { chart: any }) => {
                       openModal(p);
                   },
                   enter: ({ chart }: { chart: any }) => { chart.canvas.style.cursor = 'pointer'; },
                   leave: ({ chart }: { chart: any }) => { chart.canvas.style.cursor = 'default'; },
            };
        });
        return lines;
    }
3:05 p.m.

The line annotations are the culprit again — they span the full chart width and their hit area is tall enough to catch clicks meant for the outlier labels. They don't have click handlers so they're not doing anything visible, but they're triggering enter on your outlier and making the cursor change everywhere.

Add enter/leave no-ops to the line annotations in addLine to stop them interfering with the cursor:
typescript

lines[key] = {
    type: "line",
    yMin: value,
    yMax: value,
    borderColor: color,
    borderWidth: 1.5,
    borderDash: dash,
    label: { display: false },
    enter: () => {},
    leave: () => {},
};

This gives the line annotations their own enter/leave handlers so they don't bubble up or bleed into the outlier label hit zones.

No dice, same thing.
3:07 p.m.

Let me think about this differently. The enter hook on the outlier label is firing way too broadly — so it's probably not the line annotations at all, it's the label annotation's own hit area being huge regardless of hitTolerance.

Can you temporarily add a console.log(element) to the outlier enter hook and tell me what element looks like — specifically its x, y, width, height properties? Want to see what bounding box the plugin thinks the label has.

Where I add this?
3:09 p.m.

In the enter handler:
typescript

enter: ({ chart, element }: { chart: any, element: any }) => { 
    console.log(element.x, element.y, element.width, element.height);
    chart.canvas.style.cursor = 'pointer'; 
},

44.286934621710515 146.29999999999998 41.138671875 21
3:09 p.m.

So the label itself is only 41×21 pixels — that's not huge. The hit area should be reasonable.

The real question is: where is your cursor triggering the pointer? If it's changing across a wide horizontal band, the line annotations are still the suspect despite the no-op handlers. If it's only near that label position but a bit too tall/wide, it's the label's own hit area.

Can you check — does the cursor change in a horizontal band across the whole chart width, or only in a localized area around where the label is?

There's a horizontal band and a vertical band that are quite large, but not one single rectangle.
3:12 p.m.

That confirms it — the line annotations are causing the horizontal band, and the vertical band is likely the x-axis category hit area or a vertical slice through the label's x position.

Since the no-op enter/leave on line annotations didn't help, try setting display: false equivalent for events — add enter: undefined, leave: undefined won't work either. The real fix is to disable interaction on the line annotations entirely with:
typescript

display: true,
interaction: false,

Wait — actually the annotation plugin doesn't have an interaction: false option per annotation.

The nuclear option that should actually work: don't use annotation plugin enter/leave for the cursor at all. Instead, attach a mousemove listener to the canvas and manually check if the pointer is within the label's bounding box:
typescript

xCanvas.addEventListener('mousemove', (e) => {
    const chart = chartInstances[`${analyte.mma_id}-x`];
    if (!chart) return;
    const annotations = (chart.options.plugins as any).annotation.annotations;
    let overLabel = false;
    for (const key of Object.keys(annotations)) {
        if (!key.startsWith('outlier-')) continue;
        const el = chart.getDatasetMeta(0); // won't work
    }
});

Actually that won't easily give you the rendered element. Let me ask — is the cursor behaviour annoying enough to keep fighting, or is it acceptable to just always show pointer on the whole canvas when an outlier is present?

More so just making sure the right point is being selected for commenting. Right now I click on a point on the chart and the annotation is still hijacking it even though I am nowhere close to it.
3:14 p.m.

So the cursor thing is a red herring — the real problem is the annotation is stealing clicks from data points. Back to the horizontal band issue.

The line annotations spanning the full width are the most likely culprit for click hijacking. Since no-op handlers didn't stop them, try setting display: false on all line annotations temporarily to confirm they're the ones stealing clicks. If point clicking works perfectly without them, we know for sure and can find a way to make them non-interactive.

Yes, works perfectly without them aside from can't select points off-chart.
3:16 p.m.

OK so the line annotations need to be non-interactive. Try adding adjustScaleRange: false — actually that's not it.

The annotation plugin doesn't have a per-annotation "non-interactive" flag, but you can fake it by not registering any event hooks on them, which you've tried. The issue is the plugin still does hit detection on all annotations regardless.

One option: move the line annotations out of the annotation plugin entirely and draw them manually using Chart.js's beforeDraw plugin hook on the chart instance. Lines are trivial to draw on canvas — just ctx.beginPath(), moveTo, lineTo, stroke. Then they're completely invisible to the annotation plugin's hit detection.

Want to go that route? It's a bit more code but gives you full control and cleans up the annotation config to just outlier labels.

Sure - want me to paste in the whole file so you can tidy it a bit and remove extra nonsense I may have left in by accident? Well here it is anyway.
3:17 p.m.

Looking at the full file — the core issue is clear. Your canvas click listener and the annotation plugin are both fighting over the same canvas clicks. The annotation plugin intercepts clicks on line annotations (which span the full width) before your canvas listener even sees them.

The cleanest fix: ditch the annotation plugin for the limit lines entirely and draw them as a custom Chart.js plugin. Here's the replacement for addLine logic — add this once at the top of your script where you register plugins:
typescript

const limitLinesPlugin = {
    id: 'limitLines',
    afterDatasetsDraw(chart: any) {
        const { ctx, chartArea, scales } = chart;
        const lines = chart.options.plugins?.limitLines?.lines ?? [];
        if (!lines.length) return;
        ctx.save();
        for (const line of lines) {
            const y = scales.y.getPixelForValue(line.value);
            if (y < chartArea.top || y > chartArea.bottom) continue;
            ctx.beginPath();
            ctx.moveTo(chartArea.left, y);
            ctx.lineTo(chartArea.right, y);
            ctx.strokeStyle = line.color;
            ctx.lineWidth = 1.5;
            ctx.setLineDash(line.dash);
            ctx.stroke();
        }
        ctx.setLineDash([]);
        ctx.restore();
    }
};

Chart.register(...registerables, annotationPlugin, limitLinesPlugin);

Then in xChartOptions, replace the annotation lines config with this in plugins:
typescript

limitLines: {
    lines: [
        last.mean != null ? { value: last.mean, color: '#888', dash: [] } : null,
        last.ucl != null ? { value: last.ucl, color: '#e53e3e', dash: [] } : null,
        last.lcl != null ? { value: last.lcl, color: '#e53e3e', dash: [] } : null,
        last.uwl != null ? { value: last.uwl, color: '#dd6b20', dash: [4,4] } : null,
        last.lwl != null ? { value: last.lwl, color: '#dd6b20', dash: [4,4] } : null,
        last.uil != null ? { value: last.uil, color: '#d69e2e', dash: [2,4] } : null,
        last.lil != null ? { value: last.lil, color: '#d69e2e', dash: [2,4] } : null,
    ].filter(Boolean),
},
annotation: {
    annotations: annotationsForX(points, yMin, yMax), // outlier labels only now
},

And remove the addLine function and all the line entries from annotationsForX entirely — it only needs to build outlier labels now.

You'll also need to declare limitLines on the Chart.js plugin options type to avoid TypeScript complaints — add this somewhere at the top level:
typescript

declare module 'chart.js' {
    interface PluginOptionsByType<TType extends ChartType> {
        limitLines?: { lines: any[] };
    }
}

Same treatment for mrChartOptions — move the UCL line there to limitLines too.

There's no last in xChartOptions and it won't let me use declare.
3:24 p.m.

Right — last lives in annotationsForX, not xChartOptions. You need to pass the limit values in separately, or just extract them in xChartOptions the same way annotationsForX does:
typescript

const last = points[points.length - 1];

Add that line at the top of xChartOptions alongside the existing ucl/lcl extraction.

For the declare issue — Wails or the tsconfig is probably too strict. Just cast it instead:
typescript

limitLines: {
    lines: [ ... ]
} as any,

And on the plugin definition, type the parameter as any to avoid the property access errors:
typescript

const limitLinesPlugin = {
    id: 'limitLines',
    afterDatasetsDraw(chart: any) {
        const lines = (chart.options.plugins?.limitLines as any)?.lines ?? [];

No declaration needed that way.

TS doesn't really seem to care, what am I doing wrong?
3:28 p.m.

You still have display: false on the outlier annotations — that's hiding them from rendering AND from hit detection. Remove that line from the outlier annotation config:
typescript

// remove this:
display: false,

That was a leftover from earlier debugging and it's killing the whole thing.

Thank you - still not grabbing properly unfortunately.
3:28 p.m.

The canvas click listener is still firing and competing with the annotation click. When you click the outlier label, both handlers race — the canvas listener wins because it uses intersect: false which grabs the nearest point regardless of where you click.

In the canvas click listener, you need to check if the click landed on an outlier annotation first and bail if so. The annotation plugin exposes this via getAnnotationAtEvent — but that's not a real API in v3.

Simpler approach: track whether an annotation click is in flight with a flag.

Add a variable at the top:
typescript

let annotationClickInProgress = false;

In each outlier annotation's click handler, set it before opening the modal:
typescript

click: ({ chart }: { chart: any }) => {
    annotationClickInProgress = true;
    openModal(p);
},

In the canvas click listener, check and reset it:
typescript

xCanvas.addEventListener('click', (e) => {
    if (annotationClickInProgress) {
        annotationClickInProgress = false;
        return;
    }
    const chart = chartInstances[`${analyte.mma_id}-x`];
    if (!chart) return;
    const elements = chart.getElementsAtEventForMode(e, 'nearest', { intersect: false }, false);
    if (elements.length === 0) return;
    const idx = elements[0].index;
    if (idx < points.length) openModal(points[idx]);
});

The annotation plugin fires its click handler before the canvas click event bubbles, so the flag will be set by the time the canvas listener runs.

Still nope. Weird. Frustrating. Why is the click area so huge?!?!
3:30 p.m.

Let me look at this more carefully. You said there's a horizontal AND vertical band. That's not annotation hit detection — that's the intersect: false on your canvas listener grabbing the nearest point to wherever you click. The annotation never actually fires at all, the canvas listener just finds the nearest data point to your click regardless of distance.

So the real problem is: how do you know if a click was "close enough" to an outlier label to open it vs a data point?

The annotation click isn't firing reliably, so ditch it entirely. Instead, in the canvas click listener, manually check if the click coordinates are within any outlier label's bounding box:
typescript

xCanvas.addEventListener('click', (e) => {
    const chart = chartInstances[`${analyte.mma_id}-x`];
    if (!chart) return;

    const rect = xCanvas.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const clickY = e.clientY - rect.top;

    // check outlier annotations first
    const anns = (chart.options.plugins as any)?.annotation?.annotations ?? {};
    for (const key of Object.keys(anns)) {
        if (!key.startsWith('outlier-')) continue;
        const el = (chart as any).getDatasetMeta(0); // wrong, need annotation element
    }
});

Actually — does the annotation plugin expose rendered elements anywhere on the chart instance? Check (chart as any).annotation?.elements in the console and tell me what's there.

How about this - is there just a way to be able to select data that is off-chart? Is this annotation overkill and already present in chartjs?
3:43 p.m.

Good instinct. Chart.js clamps data points that are outside the scale range — they don't render, so you can't click them. But you can change that behaviour by setting clip: false on the dataset, which lets points render outside the chart area. Combined with a bigger pointRadius or just leaving the scale unclamped, you might not need the annotation labels at all.

But the real question is: what do you actually need? You need to be able to click a point that's beyond the UCL/LCL. The simplest fix is just don't clamp the y-axis. Remove the fixed yMin/yMax from the scale, let Chart.js auto-scale to fit all points, and keep the limit lines drawn via limitLinesPlugin. Outlier points render normally, clicks work normally, no annotations needed for them.

The tradeoff: the chart scale changes based on data, so a single extreme outlier can squash everything else. For SPC that's arguably fine — you want to see the outlier in context.

Want to go that route and drop the outlier label annotations entirely? You could still show the value in the tooltip or modal title so the reviewer knows the exact number.

How about this - render them off-screen same as now. Have a toggle on the UI - show outliers. Everyone wins? So keep the clamping how it is for now, if people want to see outliers we can just let the y scale naturally?
3:45 p.m.

Perfect. Simple, clean, no fighting the annotation plugin.

Add a toggle to state:
typescript

let showOutliers = false;

Add a button in the controls bar alongside the existing ones:
svelte

<button class="btn-secondary" on:click={() => { showOutliers = !showOutliers; buildCharts(); }}>
    {showOutliers ? "Clamp chart" : "Show outliers"}
</button>

In xChartOptions, conditionally set the y scale:
typescript

y: {
    ...(showOutliers ? {} : { min: yMin, max: yMax }),
    ticks: { font: { family: "var(--font-mono)", size: 11 } },
},

And when showOutliers is true, skip building outlier annotations in annotationsForX — they're unnecessary if the points are visible:
typescript

if (!showOutliers) {
    points.forEach((p, i) => {
        // existing outlier label logic
    });
}

That's it. Clamped by default, one click to see everything.

I feel like there's a lot of loose ends in this file now, can you clean it up and add in what you just showed me, please?
3:46 p.m.

I also notice you have a missing closing brace on the if (xCanvas) block in buildCharts — the mrCanvas block ended up inside it. I'll fix that too.

Here's the cleaned up file:
Created a file, read a file

Key changes from the messy version: removed annotationClickInProgress, all the dead enter/leave/click handlers on outlier annotations, the leftover last variable in annotationsForX, and fixed the missing closing brace on the xCanvas block. Also moved the mR UCL line to limitLinesPlugin to keep it consistent. The showOutliers toggle is wired up and ready.
Chartreview
SVELTE 

That works. Sometimes you gotta rethink it. Can you please update the project memory, I think this is in a workable place for now!
3:51 p.m.

Here's the updated memory document:
Created a file, read a file

Updated. Key additions: limitLinesPlugin design decision, chart click handling approach, outlier toggle behaviour, and comment system marked done. Trend detection and audit log are up next.
Cromulent memory
Document · MD 
You are out of free messages until 7:00 PM
Upgrade


Claude is AI and can make mistakes. Please double-check responses.
Cromulent memory · MD
Cromulent — project memory document
What is this

A reset document for continuing development of Cromulent with Claude. Paste this at the start of a new conversation to restore context.
Github repository

https://github.com/jeffreycharters/jcc
Communication style

    Concise answers preferred — no rambling
    Considered pushback welcome, don't just agree
    Technically capable — Go developer with reasonable experience
    Hobby project, wants to understand everything under the hood
    Likes "if you get it, you get it" humour (Simpsons-brained)
    Work one file at a time, one change at a time

Project overview

Cromulent — a desktop SPC (statistical process control) charting and review application for an ISO 17025 accredited laboratory. Named after the Simpsons word. Tagline: "perfectly cromulent data, every time."

Replaces Northwest Analytical (NWA) Quality Analyst, which is slow, ugly, has no chart annotations, and doesn't enforce data locking.
Core use case

    Technicians enter measurement data
    Reviewer (user) opens charts, checks for trends, adds traceable comments/rationale
    Data is locked (append-only) once entered
    Full audit trail for ISO 17025 compliance

Tech stack
Layer	Choice
Framework	Wails v2 (Go backend, web frontend)
Language	Go 1.26
Frontend	Svelte 5 + TypeScript (plain Svelte, NOT SvelteKit — single view variable for routing)
Database	SQLite via modernc.org/sqlite (pure Go, no CGO)
Auth	Username + password (bcrypt via golang.org/x/crypto/bcrypt)
Charts	Chart.js + chartjs-plugin-annotation
Font	Atkinson Hyperlegible Next + Mono (woff2, bundled in frontend assets)
Package manager	pnpm
Target OS	Windows only (dev on Linux, cross-compile with -platform windows/amd64)
Dev command	wails dev -tags webkit2_41
DnD	svelte-dnd-action
SQLite notes

    WAL mode enabled (PRAGMA journal_mode=WAL)
    Multi-user via network share (2-3 users per lab, low write contention)
    Network share path stored in local config file per workstation
    Append-only measurements table enforces data integrity architecturally
    Migrations via PRAGMA user_version — versioned migration slice in db/migrate.go
    Concurrent writes solved by using transactions rather than Promise.all parallel calls

Go project structure

db/        — SQLite setup and migrations
models/    — data structs
handlers/  — functions exposed to Svelte via Wails bindings
frontend/  — Svelte app

Authentication

    Username + password stored in SQLite users table
    Passwords bcrypt hashed
    Minimum 6 character password (threat model is audit trail integrity, not internet exposure)
    Session timeout: re-enter password after 30 minutes idle
    No magic links, no device-bound auth — credentials must work on any workstation
    Inactive users (active = 0) cannot log in — soft delete preserves audit trail

Roles
Role	Description
technician	Enters data, lands on data entry view
reviewer	Reviews charts, lands on chart review view
supervisor	Reviews charts, lands on chart review view
admin	Full access including user management

All roles can create methods, materials, analytes, and MMA combos. Trust model is audit trail integrity, not access control — user_id on everything means bad actors are traceable.
Database schema (migration v4)
Core lookup tables

users — id, username, password_hash, role, active (default 1), created_at

materials — id, name, description

methods — id, name, description

analytes — id, name, unit — UNIQUE(name, unit) so molybdenum/ppm and molybdenum/ppb are distinct
Relationships

material_method_analytes — id, material_id FK, method_id FK, analyte_id FK, display_order, active (default 1)

    UNIQUE(material_id, method_id, analyte_id)
    The unique combination that owns its own control limits and measurement history
    active allows soft-inactivation of retired combos — operates at combo level (all analytes), not per-analyte
    No deletes once referenced by measurements (FK enforcement)

control_limit_regions — id, material_method_analyte_id FK, mean, ucl (NOT NULL), lcl (NOT NULL), uwl, lwl, uil, lil, effective_from_sequence, created_by FK, created_at, deleted_at, deleted_by FK

    Soft delete only — deleted_at/deleted_by set on deletion, never hard deleted
    This preserves audit trail; audit log view can query deleted_at IS NOT NULL later

Three pairs of limits:

    ucl/lcl — control limits (required, typically ±3 SD)
    uwl/lwl — warning limits (optional, typically ±2 SD)
    uil/lil — inner limits (optional, typically ±1 SD)

Limits versioned by sequence number. To find limits for sequence N: query where effective_from_sequence <= N ORDER BY effective_from_sequence DESC LIMIT 1.
Chart / run tables

control_charts — id, material_id FK, method_id FK, batch_id, technician_id FK, created_at, locked_at

    locked_at is set immediately on save — all charts are immutable from creation

chart_metadata_fields — id, name, required, display_order

chart_metadata_values — id, control_chart_id FK, field_id FK, value

measurements — id, control_chart_id FK, material_method_analyte_id FK, value, sequence_order, sequence_number, entered_by FK, entered_at

    sequence_order — column position within a chart (instrument order)
    sequence_number — immutable per-MMA run number, computed at insert time via COALESCE(MAX(sequence_number), 0) + 1 inside the SaveChart transaction
    Append-only. No UPDATE statements ever issued against this table.

comments — id, control_chart_id FK, measurement_id FK (nullable), text, user_id FK, created_at

    measurement_id IS NULL → chart-level comment (e.g. bad dilution affecting whole run)
    measurement_id IS NOT NULL → point-level comment (single dodgy analyte)
    Append-only, no edits or deletes — full comment history preserved for audit trail
    Multiple comments per point/chart allowed, sorted by created_at

spc_rule_sets — id, beyond_sigma_enabled, beyond_sigma_n, run_trend_enabled, run_trend_n, one_side_enabled, one_side_n, effective_from_date, created_by FK, created_at
Key design decisions

    No SvelteKit — plain Svelte with top-level let view = 'login' for navigation
    Append-only measurements — data integrity enforced architecturally, not by policy
    sequence_number — per-MMA immutable run number on measurements, used as x-axis on charts and paperwork reference
    sequence_order — instrument column order within a chart, separate from sequence_number
    Hybrid metadata — technician_id is a proper FK, everything else flexible via metadata tables
    Control limit regions — versioned per material_method_analyte by sequence number, soft-deleted only
    No pass/fail shown live during data entry — deliberate, prevents pressure to fudge values before saving; shown after save only
    SQLite over Postgres — IT won't support a Postgres server; WAL mode acceptable at 2-3 users
    Atkinson Hyperlegible Next — bundled font for readability; Mono variant for data display
    No Jet/sqlx — plain database/sql; schema is simple enough, overhead not worth it
    Wails v2 not v3 — v3 still alpha, no multi-window needed anyway
    UserResponse not User — Wails bindings can't handle time.Time; all frontend-bound returns use UserResponse with string timestamps
    Light theme — better for well-lit lab environment; backed by readability research
    Go slices serialize as null — empty slices from Go come back as null in JS, always use ?? [] on list results
    Clipboard parsing — LibreOffice Calc puts data as text/html, not text/plain. Parse HTML table cells via DOMParser first, fall back to tab-split plain text. Excel also uses HTML format so this handles both.
    Decimal separator — regex currently only accepts . as decimal separator. Locale issues deferred until if/when needed.
    No LATERAL joins — SQLite doesn't support them; use correlated subqueries instead
    Combo-level deactivation — DeactivateCombo/ActivateCombo operate on all analytes for a method+material at once
    Chart data limit — default 50 points, user-adjustable input; limit <= 0 means no limit in handler
    Wails JSON serialization — all bound handler return values serialize to camelCase JSON. Struct fields like MMAID, UCL, LCL come back as mmaID, ucl, lcl. Always check actual runtime keys, not struct field names. models.ChartPoint is not generated by Wails (returned inside a map) — define it manually in the frontend.
    mR UCL calculation — derived from X chart limits: (UCL - LCL) * 0.61, not computed from data. This keeps the mR scale stable when outliers are present.
    mR chart scale — based on (UCL - LCL) * 1.3, independent of actual mR values. Outliers shown as annotations.
    X chart scale — based on UCL/LCL with 30% padding each side. Outliers shown as inline label annotations with ▲/▼ and value.
    Sig figs — hardcoded at 3 for now via sigFigs(value, n=3). Planned as a per-user setting.
    Passthrough analytes — analytes with no control_limit_regions row are included in the combo for paste convenience but not control charted. Frontend filters them from chart rendering. saveLimits filters out analytes where UCL and LCL are both blank before building payload.
    limitLinesPlugin — custom Chart.js plugin that draws limit lines (mean, UCL/LCL, UWL/LWL, UIL/LIL) directly on canvas via afterDatasetsDraw. Replaces annotation plugin line annotations which have uncontrollable hit detection that intercepts point clicks.
    Chart.js point clicking — handled via canvas addEventListener('click', ...) using getElementsAtEventForMode with mode: 'nearest', intersect: false. Do NOT use Chart.js onClick option — it conflicts with annotation plugin click handling.
    Outlier annotations — chartjs-plugin-annotation label annotations are kept only for ▲/▼ outlier indicators when chart is clamped. They are display-only (no click handlers) to avoid hit detection conflicts.
    Show outliers toggle — button in controls bar toggles showOutliers state. When true, y-axis is unclamped (no min/max) and outlier label annotations are suppressed. When false, y-axis is clamped to UCL/LCL ± 30% and outlier labels are shown. Rebuilds charts on toggle.

Models
models/user.go
go

type User struct {
    ID           int64
    Username     string
    PasswordHash string
    Role         string
    Active       bool
    CreatedAt    time.Time
}

type UserResponse struct {
    ID        int64  `json:"id"`
    Username  string `json:"username"`
    Role      string `json:"role"`
    Active    bool   `json:"active"`
    CreatedAt string `json:"created_at"`
}

type Role string

const (
    RoleTechnician Role = "technician"
    RoleReviewer   Role = "reviewer"
    RoleSupervisor Role = "supervisor"
    RoleAdmin      Role = "admin"
)

models/library.go
go

type Analyte struct { ID int64; Name string; Unit string }
type Method struct { ID int64; Name string; Description string }
type Material struct { ID int64; Name string; Description string }

type MMAEntry struct {
    ID int64; MaterialID int64; MaterialName string
    MethodID int64; MethodName string
    AnalyteID int64; AnalyteName string
    Unit string; DisplayOrder int; Active bool
}

type MethodWithMaterials struct { ID int64; Name string; Materials []MaterialSummary }
type MaterialSummary struct { ID int64; Name string }
type ComboAnalyte struct { MMAID int64; Name string; Unit string; DisplayOrder int }

type ControlLimitRegion struct {
    ID int64; MMAID int64
    Mean float64; UCL float64; LCL float64
    UWL *float64; LWL *float64; UIL *float64; LIL *float64
    EffectiveFromSequence int; CreatedBy int64; CreatedAt string
}

type MeasurementResult struct {
    MMAID int64; AnalyteName string; Unit string
    Value float64; SequenceNumber int
    UCL *float64; LCL *float64
    Pass bool; NoLimits bool
}

type ChartPoint struct {
    MeasurementID  int64
    ControlChartID int64
    SequenceNumber int
    Value          float64
    Mean           *float64
    UCL            *float64
    LCL            *float64
    UWL            *float64
    LWL            *float64
    UIL            *float64
    LIL            *float64
}

models/comment.go
go

type CommentResponse struct {
    ID             int64  `json:"id"`
    ControlChartID int64  `json:"control_chart_id"`
    MeasurementID  *int64 `json:"measurement_id"`
    Text           string `json:"text"`
    UserID         int64  `json:"user_id"`
    Username       string `json:"username"`
    CreatedAt      string `json:"created_at"`
}

Handlers (Wails-bound)
handlers/auth.go — app.Auth

    Login(username, password string) (*models.UserResponse, error)
    Logout()
    CurrentUser() *models.UserResponse
    IsAuthenticated() bool
    ListUsers() ([]models.UserResponse, error)
    DeactivateUser(id int64) error
    ActivateUser(id int64) error
    CreateUser(username, password, role string) error

handlers/setup.go — app.Setup

    NeedsSetup() bool
    CreateAdminUser(username, password string) error
    UserExists(username string) bool

handlers/config.go — app.Config

    GetDBPath() (string, error)
    InitDB() error
    OpenDBFilePicker() (string, error)
    OpenDBFolderPicker() (string, error)
    SetDBPath(path string) error
    SetContext(ctx context.Context)

handlers/library.go — app.Library

    CreateAnalyte(name, unit string) error
    ListAnalytes() ([]models.Analyte, error)
    CreateMethod(name, description string) error
    ListMethods() ([]models.Method, error)
    CreateMaterial(name, description string) error
    ListMaterials() ([]models.Material, error)

handlers/mma.go — app.MMA

    AddAnalyteToMMA(materialID, methodID, analyteID int64, displayOrder int) error
    ListMMAsForMethod(methodID int64) ([]models.MMAEntry, error)
    ListAllMMAs() ([]models.MMAEntry, error)
    RemoveAnalyteFromMMA(id int64) error
    UpdateDisplayOrders(ids []int64, orders []int) error
    ListUsedMMAIDs() ([]int64, error)
    DeactivateCombo(materialID, methodID int64) error
    ActivateCombo(materialID, methodID int64) error

handlers/dataentry.go — app.DataEntry

    ListMethodsWithMaterials() ([]models.MethodWithMaterials, error) — active combos only
    GetAnalytesForCombo(methodID, materialID int64) ([]models.ComboAnalyte, error)
    SaveChart(methodID, materialID, technicianID int64, values map[string]float64) (int64, error) — single transaction, sets locked_at immediately, computes sequence_number per MMA
    GetChartResults(chartID int64) ([]models.MeasurementResult, error) — pass/fail per analyte after save

handlers/limits.go — app.Limits

    GetCurrentSequencesForMMAs(ids []int64) (map[int64]int, error)
    SaveControlLimitRegions(regions []models.ControlLimitRegion) error
    ListControlLimitRegionsForCombo(materialID, methodID int64) ([]models.ControlLimitRegion, error)
    DeleteControlLimitRegionSet(materialID, methodID int64, effectiveFromSequence int, userID int64) error

handlers/chartreview.go — app.ChartReview

    GetComboChartData(materialID, methodID int64, limit int) (map[int64][]models.ChartPoint, error) — correlated subqueries for limit lookup per point, limit <= 0 means no limit. Each ChartPoint includes MeasurementID and ControlChartID for comment linking.

handlers/comments.go — app.Comments

    AddComment(chartID int64, measurementID *int64, text string, userID int64) error
    GetCommentsForChart(chartID int64) ([]models.CommentResponse, error)
    GetCommentsForCombo(materialID, methodID int64) ([]models.CommentResponse, error) — all comments for all charts in this combo, joined with username

Frontend structure

frontend/src/
├── App.svelte              — top-level view router, owns `let view` and `currentUser`
├── main.ts
├── style.css               — global styles, CSS vars, font-face declarations
├── assets/fonts/           — Atkinson Hyperlegible Next + Mono woff2 files
└── lib/
    ├── Login.svelte
    ├── Setup.svelte
    ├── Shell.svelte        — navbar + content slot, resets scroll on view change via bind:this
    ├── Admin.svelte
    ├── DBPicker.svelte
    ├── Settings.svelte
    ├── Library.svelte      — tabbed: Analytes, Methods, Materials, Combos, Limits
    ├── DataEntry.svelte    — sidebar + grid + paste + save + post-save pass/fail + chart-level comment
    └── ChartReview.svelte  — XmR individuals + mR charts, combo card picker, raw data toggle, outlier toggle, comment modal

View routing in App.svelte

    loading → checks NeedsSetup on mount
    db-pick → first launch or no config
    setup → first run only
    login → standard login
    data-entry → technician/all roles
    chart-review → reviewer/supervisor/admin
    library → all roles
    admin → admin only
    settings → all roles

CSS variables (in style.css)
css

--font-sans, --font-mono
--colour-bg, --colour-surface, --colour-border
--colour-text, --colour-text-muted
--colour-primary, --colour-primary-hover
--colour-danger, --colour-success
--radius, --radius-lg
--shadow-sm, --shadow-md

Library.svelte notes

    Tabs: Analytes, Methods, Materials, Combos, Limits
    All Go list calls return null for empty slices — always ?? []
    Combos tab: method selector → material selector (all materials shown, filtered by showHidden for inactive)
    comboActiveForMat(matID) — returns true if no combo exists yet (unlinked), or if combo is active
    showHidden checkbox shows inactive combos in material dropdown with (inactive) label
    Deactivate/Activate operates at combo level via DeactivateCombo/ActivateCombo
    Analyte list is drag-to-reorder via svelte-dnd-action
    Reorder fires UpdateDisplayOrders in a single transaction
    Remove button hidden if MMA has any measurements (usedMMAIDs set)
    Limits tab: combo cards → select combo → shows existing region sets + new paste grid
    Existing regions grouped by effective_from_sequence, each deletable (soft delete)
    New region: 7-row × N-analyte paste grid, 2D anchor (fills right and down from paste cell)
    newGrid initialized reactively when limitAnalytes.length > 0
    Object.entries(groupedRegions) types values as unknown — cast inside getRegionValue(regions: unknown)
    getRegionValue must guard both !list AND !r before accessing properties — list.find() returns undefined if no match
    saveLimits filters out analytes where UCL and LCL are both blank — passthrough analytes silently dropped from payload

DataEntry.svelte notes

    Sidebar: methods as labels, materials as buttons, active combos only (from ListMethodsWithMaterials)
    Grid: wrapping flex of analyte cards
    Paste: 1D, fills right from anchor cell using HTML-first clipboard parsing
    After save: cards show result value + Pass/Fail/No limits badge, colored border
    Results cleared on combo change or new save
    results typed as Record<string, MeasurementResult> (Object.fromEntries always gives string keys)
    chartID hoisted to component state (not just local in save()), set after SaveChart
    After save, comment section appears below grid — chart-level comment only (no point-level in DataEntry)
    AddComment(chartID, null, text, userID) — measurement_id is null for chart-level comments
    Comment input and chartID cleared on combo change

ChartReview.svelte notes

    Combo picker: flat list of method+material cards, selecting one loads immediately
    Controls bar (points limit, per-row, reload, show outliers, show data) only shown once a combo is selected
    Per-analyte X (individuals) chart + mR (moving range) chart stacked
    limitLinesPlugin — custom plugin draws mean/UCL/LCL/UWL/LWL/UIL/LIL as canvas lines, completely bypassing annotation plugin hit detection. Configured via plugins.limitLines.lines array in chart options.
    Point clicking — canvas addEventListener('click') with getElementsAtEventForMode('nearest', {intersect: false}). No Chart.js onClick option used.
    Outlier toggle — showOutliers boolean. When false: y-axis clamped to UCL/LCL ±30%, ▲/▼ label annotations shown for out-of-range points (display only, no click handlers). When true: y-axis unclamped, no outlier annotations, all points visible and clickable.
    Minimum 20 x-axis labels even if fewer data points
    Charts per row: user-selectable 1–5, responsive flex layout
    Raw data toggle: shows single table of all analytes for the selected combo, replaces charts
    Sig figs: hardcoded at 3 via sigFigs() helper — planned as per-user setting
    ChartPoint not generated by Wails bindings (inside map value) — defined manually in frontend
    Chart instances tracked in chartInstances record, destroyed before rebuild and on component destroy
    Reactive statement guards on !loading && !showRawData && analytes.length > 0 to trigger rebuild on toggle back to charts
    Comments loaded via GetCommentsForCombo on combo select, reloaded after adding a comment
    commentsByMeasurement: Record<number, CommentResponse[]> — indexed by measurement_id
    commentsByChart: Record<number, CommentResponse[]> — indexed by control_chart_id
    Points with comments shown as larger yellow dots via pointBackgroundColor/pointRadius arrays in dataset
    Clicking a data point opens comment modal — shows sequence number, value, limit context, existing comments, new comment textarea
    limitContext(point) — returns human-readable string for where value sits relative to limits
    TODO: Raw data table needs violation indicators (symbol beside value) with click-to-comment
    TODO: comment_type — deferred. May use to distinguish point violations from chart-level observations. Can add with a default value migration when needed.

Shell.svelte notes

    bind:this={contentEl} on .content div
    $: if (view && contentEl) contentEl.scrollTop = 0 resets scroll on navigate

Migrations

    v1 — full initial schema
    v2 — ALTER TABLE material_method_analytes ADD COLUMN active INTEGER NOT NULL DEFAULT 1
    v3 — ALTER TABLE measurements ADD COLUMN sequence_number INTEGER
    v4 — ALTER TABLE control_limit_regions ADD COLUMN deleted_at DATETIME; ALTER TABLE control_limit_regions ADD COLUMN deleted_by INTEGER REFERENCES users(id)

Plan: merge all into single v1 migration before release — easy cleanup task.
What's done

    Wails + Svelte + TypeScript project scaffold
    pnpm configured in wails.json
    Go 1.26, modernc.org/sqlite, golang.org/x/crypto/bcrypt
    DB init with WAL mode + foreign keys + busy timeout
    Versioned migrations via PRAGMA user_version (v4)
    Full schema
    Auth handler with session timeout (30 min)
    Role-based post-login routing
    First-run setup screen
    Login screen
    App shell with role-filtered navbar, scroll reset on navigate
    Admin user management screen
    config package
    ConfigHandler
    DBPicker.svelte
    Settings.svelte
    Library.svelte — analytes, methods, materials CRUD + MMA combo wiring + drag-to-reorder + deactivate/activate combo + show hidden toggle + control limits tab with paste grid + soft delete
    DataEntry.svelte — sidebar, analyte card grid, clipboard paste, save, post-save pass/fail, chart-level comment input
    ChartReview.svelte — XmR individuals + mR charts, combo card picker, outlier toggle, raw data toggle, responsive n-per-row layout, comment loading, yellow dot indicators for commented points, comment modal with existing comments + new comment input
    CommentsHandler — AddComment, GetCommentsForChart, GetCommentsForCombo

What's next

    Raw data table violation indicators + click-to-comment
    Trend detection against spc_rule_sets
    Audit log view

Planned build order (revised)

    Wails init ✅
    SQLite setup + migrations ✅
    Auth ✅
    Library setup ✅
    Data entry ✅
    Control limits setup ✅
    XmR control charts ✅
    Comment/annotation system ✅
    Trend detection (SPC rule sets)
    Audit log view
