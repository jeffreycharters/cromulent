package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"fmt"
	"time"
)

type LimitsHandler struct{}

func NewLimitsHandler() *LimitsHandler {
	return &LimitsHandler{}
}

// GetCurrentSequencesForMMAs returns the current max sequence_number
// for each MMA id, keyed by MMA id. Missing MMAs (no measurements yet) return 0.
func (h *LimitsHandler) GetCurrentSequencesForMMAs(ids []int64) (map[int64]int, error) {
	result := make(map[int64]int, len(ids))
	for _, id := range ids {
		var seq int
		err := db.DB.QueryRow(`
			SELECT COALESCE(MAX(sequence_number), 0)
			FROM measurements
			WHERE material_method_analyte_id = ?
		`, id).Scan(&seq)
		if err != nil {
			return nil, fmt.Errorf("sequence query for mma %d: %w", id, err)
		}
		result[id] = seq
	}
	return result, nil
}

// SaveControlLimitRegions saves a batch of control limit regions in one transaction.
func (h *LimitsHandler) SaveControlLimitRegions(regions []models.ControlLimitRegion) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	for _, r := range regions {
		_, err := tx.Exec(`
			INSERT INTO control_limit_regions
				(material_method_analyte_id, mean, ucl, lcl, uwl, lwl, uil, lil, effective_from_sequence, created_by, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, r.MMAID, r.Mean, r.UCL, r.LCL, r.UWL, r.LWL, r.UIL, r.LIL, r.EffectiveFromSequence, r.CreatedBy, now)
		if err != nil {
			return fmt.Errorf("insert limit region for mma %d: %w", r.MMAID, err)
		}
	}

	return tx.Commit()
}

// ListControlLimitRegionsForCombo returns all non-deleted regions for a method+material combo.
func (h *LimitsHandler) ListControlLimitRegionsForCombo(methodMaterialID int64) ([]models.ControlLimitRegion, error) {
	rows, err := db.DB.Query(`
        SELECT
            clr.id,
            clr.material_method_analyte_id,
            clr.mean,
            clr.ucl,
            clr.lcl,
            clr.uwl,
            clr.lwl,
            clr.uil,
            clr.lil,
            clr.effective_from_sequence,
            clr.created_by,
            clr.created_at
        FROM control_limit_regions clr
        JOIN material_method_analytes mma ON mma.id = clr.material_method_analyte_id
        WHERE mma.method_material_id = ? AND clr.deleted_at IS NULL
        ORDER BY clr.effective_from_sequence, clr.material_method_analyte_id
    `, methodMaterialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []models.ControlLimitRegion
	for rows.Next() {
		var r models.ControlLimitRegion
		if err := rows.Scan(
			&r.ID, &r.MMAID, &r.Mean, &r.UCL, &r.LCL,
			&r.UWL, &r.LWL, &r.UIL, &r.LIL,
			&r.EffectiveFromSequence, &r.CreatedBy, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, rows.Err()
}

func (h *LimitsHandler) DeleteControlLimitRegionSet(methodMaterialID int64, effectiveFromSequence int, userID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.DB.Exec(`
        UPDATE control_limit_regions
        SET deleted_at = ?, deleted_by = ?
        WHERE effective_from_sequence = ?
          AND material_method_analyte_id IN (
              SELECT id FROM material_method_analytes
              WHERE method_material_id = ?
          )
          AND deleted_at IS NULL
    `, now, userID, effectiveFromSequence, methodMaterialID)
	return err
}
