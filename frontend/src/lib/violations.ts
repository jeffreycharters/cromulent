export type ViolationCode = "OOC" | "WRN" | "TRD" | "RUN";

export type ViolationMap = Map<number, ViolationCode[]>; // keyed by measurement_id

interface Point {
  measurement_id: number;
  value: number;
  mean: number | null;
  ucl: number | null;
  lcl: number | null;
  uwl: number | null;
  lwl: number | null;
}

function sorted(points: Point[]): Point[] {
  return [...points].sort((a, b) => a.measurement_id - b.measurement_id);
}

function addViolation(map: ViolationMap, id: number, code: ViolationCode) {
  const existing = map.get(id) ?? [];
  if (!existing.includes(code)) map.set(id, [...existing, code]);
}

export function computeViolations(raw: Point[]): ViolationMap {
  const points = sorted(raw);
  const map: ViolationMap = new Map();

  for (let i = 0; i < points.length; i++) {
    const p = points[i];

    // OOC — outside control limits
    if (
      (p.ucl != null && p.value > p.ucl) ||
      (p.lcl != null && p.value < p.lcl)
    ) {
      addViolation(map, p.measurement_id, "OOC");
    }

    // WRN — 2 of 3 consecutive outside warning limits (same side)
    if (i >= 2) {
      const window = [points[i - 2], points[i - 1], points[i]];
      const aboveWarn = window.filter(
        (w) => w.uwl != null && w.value > w.uwl,
      ).length;
      const belowWarn = window.filter(
        (w) => w.lwl != null && w.value < w.lwl,
      ).length;
      if (aboveWarn >= 2 || belowWarn >= 2) {
        for (const w of window) addViolation(map, w.measurement_id, "WRN");
      }
    }

    // TRD — 6 consecutive increasing or decreasing
    if (i >= 5) {
      const window = points.slice(i - 5, i + 1);
      let inc = true,
        dec = true;
      for (let j = 1; j < window.length; j++) {
        if (window[j].value <= window[j - 1].value) inc = false;
        if (window[j].value >= window[j - 1].value) dec = false;
      }
      if (inc || dec) addViolation(map, p.measurement_id, "TRD");
    }

    // RUN — 8 consecutive on same side of mean
    if (i >= 7) {
      const window = points.slice(i - 7, i + 1);
      const allAbove = window.every((w) => w.mean != null && w.value > w.mean);
      const allBelow = window.every((w) => w.mean != null && w.value < w.mean);
      if (allAbove || allBelow) addViolation(map, p.measurement_id, "RUN");
    }
  }

  return map;
}

export const VIOLATION_LABELS: Record<ViolationCode, string> = {
  OOC: "Outside control limits",
  WRN: "2 of 3 outside warning limits",
  TRD: "6-point trend",
  RUN: "8 points one side of mean",
};

// Highest severity wins for cell colour
export function worstViolation(codes: ViolationCode[]): ViolationCode | null {
  if (codes.includes("OOC")) return "OOC";
  if (codes.includes("WRN")) return "WRN";
  if (codes.includes("TRD")) return "TRD";
  if (codes.includes("RUN")) return "RUN";
  return null;
}
