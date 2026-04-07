export type ViolationCode = "OOC" | "WRN" | "TRD" | "RUN";
export type ViolationMap = Map<number, ViolationCode[]>; // keyed by measurement_id

export interface RuleSet {
  effectiveFromSequence: number | null; // null = applies from the beginning
  beyondLimitsEnabled: boolean;
  warningLimitsEnabled: boolean;
  warningConsecutiveCount: number;
  warningTriggerCount: number;
  trendEnabled: boolean;
  trendConsecutiveCount: number;
  oneSideEnabled: boolean;
  oneSideConsecutiveCount: number;
}

export interface Point {
  measurement_id: number;
  sequence_number: number;
  value: number;
  mean: number | null;
  ucl: number | null;
  lcl: number | null;
  uwl: number | null;
  lwl: number | null;
}

// Returns the rule set active at a given sequence number.
// Rule sets must be sorted by effectiveFromSequence ascending, global default last.
function ruleSetAt(ruleSets: RuleSet[], sequence: number): RuleSet {
  let active = ruleSets[ruleSets.length - 1]; // global default fallback
  for (const rs of ruleSets) {
    if (rs.effectiveFromSequence == null) continue;
    if (sequence >= rs.effectiveFromSequence) active = rs;
  }
  return active;
}

// Returns the index in points[] where the rule set changes, or -1 if none.
// Used to reset windows at boundaries.
function ruleSetBoundaryBefore(
  ruleSets: RuleSet[],
  points: Point[],
  start: number,
  end: number, // inclusive
): number {
  if (ruleSets.length <= 1) return -1;
  const endRS = ruleSetAt(ruleSets, points[end].sequence_number);
  for (let i = end - 1; i >= start; i--) {
    if (ruleSetAt(ruleSets, points[i].sequence_number) !== endRS) return i + 1;
  }
  return -1;
}

function sorted(points: Point[]): Point[] {
  return [...points].sort((a, b) => a.sequence_number - b.sequence_number);
}

function addViolation(map: ViolationMap, id: number, code: ViolationCode) {
  const existing = map.get(id) ?? [];
  if (!existing.includes(code)) map.set(id, [...existing, code]);
}

// ruleSets: MMA-specific regions sorted by effectiveFromSequence ASC,
// with the global default appended at the end (effectiveFromSequence: null).
export function computeViolations(
  raw: Point[],
  ruleSets: RuleSet[],
): ViolationMap {
  const points = sorted(raw);
  const map: ViolationMap = new Map();

  for (let i = 0; i < points.length; i++) {
    const p = points[i];
    const rs = ruleSetAt(ruleSets, p.sequence_number);

    // OOC — outside control limits
    if (rs.beyondLimitsEnabled) {
      if (
        (p.ucl != null && p.value > p.ucl) ||
        (p.lcl != null && p.value < p.lcl)
      ) {
        addViolation(map, p.measurement_id, "OOC");
      }
    }

    // WRN — warningTriggerCount of warningConsecutiveCount outside warning limits (same side)
    if (rs.warningLimitsEnabled) {
      const wrnWindow = rs.warningConsecutiveCount;
      if (i >= wrnWindow - 1) {
        const boundary = ruleSetBoundaryBefore(
          ruleSets,
          points,
          i - wrnWindow + 1,
          i,
        );
        if (boundary === -1) {
          const window = points.slice(i - wrnWindow + 1, i + 1);
          const aboveWarn = window.filter(
            (w) => w.uwl != null && w.value > w.uwl,
          ).length;
          const belowWarn = window.filter(
            (w) => w.lwl != null && w.value < w.lwl,
          ).length;
          if (
            aboveWarn >= rs.warningTriggerCount ||
            belowWarn >= rs.warningTriggerCount
          ) {
            for (const w of window) addViolation(map, w.measurement_id, "WRN");
          }
        }
      }
    }

    // TRD — trendConsecutiveCount consecutive increasing or decreasing
    if (rs.trendEnabled) {
      const trdWindow = rs.trendConsecutiveCount;
      if (i >= trdWindow - 1) {
        const boundary = ruleSetBoundaryBefore(
          ruleSets,
          points,
          i - trdWindow + 1,
          i,
        );
        if (boundary === -1) {
          const window = points.slice(i - trdWindow + 1, i + 1);
          let inc = true,
            dec = true;
          for (let j = 1; j < window.length; j++) {
            if (window[j].value <= window[j - 1].value) inc = false;
            if (window[j].value >= window[j - 1].value) dec = false;
          }
          if (inc || dec) addViolation(map, p.measurement_id, "TRD");
        }
      }
    }

    // RUN — oneSideConsecutiveCount consecutive on same side of mean
    if (rs.oneSideEnabled) {
      const runWindow = rs.oneSideConsecutiveCount;
      if (i >= runWindow - 1) {
        const boundary = ruleSetBoundaryBefore(
          ruleSets,
          points,
          i - runWindow + 1,
          i,
        );
        if (boundary === -1) {
          const window = points.slice(i - runWindow + 1, i + 1);
          const allAbove = window.every(
            (w) => w.mean != null && w.value > w.mean,
          );
          const allBelow = window.every(
            (w) => w.mean != null && w.value < w.mean,
          );
          if (allAbove || allBelow) addViolation(map, p.measurement_id, "RUN");
        }
      }
    }
  }

  return map;
}

// VIOLATION_LABELS are dynamic now but these are reasonable defaults for display.
// Callers that have the active rule set can build their own labels if needed.
export const VIOLATION_LABELS: Record<ViolationCode, string> = {
  OOC: "Outside control limits",
  WRN: "Outside warning limits",
  TRD: "Consecutive trend",
  RUN: "Points one side of mean",
};

// Highest severity wins for cell colour
export function worstViolation(codes: ViolationCode[]): ViolationCode | null {
  if (codes.includes("OOC")) return "OOC";
  if (codes.includes("WRN")) return "WRN";
  if (codes.includes("TRD")) return "TRD";
  if (codes.includes("RUN")) return "RUN";
  return null;
}
