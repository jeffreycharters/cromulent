package handlers

import (
	"cromulent/db"
	"cromulent/models"
)

type ChartReviewHandler struct{}

func NewChartReviewHandler() *ChartReviewHandler {
	return &ChartReviewHandler{}
}

// GetComboChartData returns chart points for all analytes in a method+material combo,
// keyed by mma_id, each ordered by sequence_number ascending.
// limit <= 0 means no limit.
func (h *ChartReviewHandler) GetComboChartData(methodMaterialID int64, limit int) (map[int64][]models.ChartPoint, error) {
	q := `
    SELECT
        m.material_method_analyte_id,
        m.id,
        m.control_chart_id,
        m.sequence_number,
        m.value,
        (SELECT mean FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS mean,
        (SELECT ucl FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS ucl,
        (SELECT lcl FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS lcl,
        (SELECT uwl FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS uwl,
        (SELECT lwl FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS lwl,
        (SELECT uil FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS uil,
        (SELECT lil FROM control_limit_regions
         WHERE material_method_analyte_id = m.material_method_analyte_id
           AND effective_from_sequence <= m.sequence_number
           AND deleted_at IS NULL
         ORDER BY effective_from_sequence DESC LIMIT 1) AS lil
    FROM (
        SELECT id, control_chart_id, material_method_analyte_id, sequence_number, value
        FROM measurements
        WHERE material_method_analyte_id IN (
            SELECT id FROM material_method_analytes
            WHERE method_material_id = ?
        )
        ORDER BY sequence_number DESC
        LIMIT ?
    ) m
    ORDER BY m.sequence_number ASC
`
	if limit <= 0 {
		limit = -1
	}

	rows, err := db.DB.Query(q, methodMaterialID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]models.ChartPoint)
	for rows.Next() {
		var mmaID int64
		var p models.ChartPoint
		if err := rows.Scan(
			&mmaID, &p.MeasurementID, &p.ControlChartID, &p.SequenceNumber, &p.Value,
			&p.Mean, &p.UCL, &p.LCL,
			&p.UWL, &p.LWL, &p.UIL, &p.LIL,
		); err != nil {
			return nil, err
		}
		result[mmaID] = append(result[mmaID], p)
	}
	return result, rows.Err()
}
