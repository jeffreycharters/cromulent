package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"fmt"
	"time"
)

type DataEntryHandler struct{}

func NewDataEntryHandler() *DataEntryHandler {
	return &DataEntryHandler{}
}

// ListMethodsWithMaterials returns the sidebar structure —
// methods, each with their active material+analyte combos.
func (h *DataEntryHandler) ListMethodsWithMaterials() ([]models.MethodWithMaterials, error) {
	rows, err := db.DB.Query(`
		SELECT
			met.id,
			met.name,
			mat.id,
			mat.name,
			mm.id
		FROM methods met
		JOIN method_materials mm ON mm.method_id = met.id
		JOIN materials mat ON mat.id = mm.material_id
		WHERE mm.active = 1
		ORDER BY met.name, mat.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methodMap []models.MethodWithMaterials
	index := map[int64]int{}

	for rows.Next() {
		var methodID, materialID, mmID int64
		var methodName, materialName string
		if err := rows.Scan(&methodID, &methodName, &materialID, &materialName, &mmID); err != nil {

			return nil, err
		}
		if _, ok := index[methodID]; !ok {
			methodMap = append(methodMap, models.MethodWithMaterials{
				ID:        methodID,
				Name:      methodName,
				Materials: []models.MaterialSummary{},
			})
			index[methodID] = len(methodMap) - 1
		}
		i := index[methodID]
		methodMap[i].Materials = append(methodMap[i].Materials, models.MaterialSummary{
			ID:               materialID,
			Name:             materialName,
			MethodMaterialID: mmID,
		})
	}
	return methodMap, rows.Err()
}

// GetAnalytesForCombo returns the ordered analytes for a method+material combo.
func (h *DataEntryHandler) GetAnalytesForCombo(methodMaterialID int64) ([]models.ComboAnalyte, error) {
	rows, err := db.DB.Query(`
		SELECT mma.id, a.name, a.unit, mma.display_order, mma.render_chart
		FROM material_method_analytes mma
		JOIN analytes a ON a.id = mma.analyte_id
		WHERE mma.method_material_id = ?
		ORDER BY mma.display_order
	`, methodMaterialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytes []models.ComboAnalyte
	for rows.Next() {
		var a models.ComboAnalyte
		if err := rows.Scan(&a.MMAID, &a.Name, &a.Unit, &a.DisplayOrder, &a.RenderChart); err != nil {
			return nil, err
		}
		analytes = append(analytes, a)
	}
	return analytes, rows.Err()
}

// SaveChart creates a control_chart and its measurements in one transaction.
// Values map is mma_id -> value, missing keys mean no measurement for that analyte.
func (h *DataEntryHandler) SaveChart(methodMaterialID, technicianID int64, values map[string]float64) (int64, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	res, err := tx.Exec(`
    INSERT INTO control_charts (method_material_id, technician_id, created_at, locked_at)
    VALUES (?, ?, ?, ?)
`, methodMaterialID, technicianID, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert chart: %w", err)
	}

	chartID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	seq := 0
	for mmaIDStr, value := range values {
		var mmaID int64
		fmt.Sscanf(mmaIDStr, "%d", &mmaID)

		var seqNum int
		err := tx.QueryRow(`
			SELECT COALESCE(MAX(sequence_number), 0) + 1
			FROM measurements
			WHERE material_method_analyte_id = ?
		`, mmaID).Scan(&seqNum)
		if err != nil {
			return 0, fmt.Errorf("compute sequence_number: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO measurements (control_chart_id, material_method_analyte_id, value, sequence_order, sequence_number, entered_by, entered_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, chartID, mmaID, value, seq, seqNum, technicianID, now)
		if err != nil {
			return 0, fmt.Errorf("insert measurement: %w", err)
		}
		seq++
	}

	return chartID, tx.Commit()
}

// GetChartResults returns pass/fail results for all measurements in a chart,
// evaluated against the applicable control limit region for each MMA.
func (h *DataEntryHandler) GetChartResults(chartID int64) ([]models.MeasurementResult, error) {
	rows, err := db.DB.Query(`
		SELECT
			m.material_method_analyte_id,
			a.name,
			a.unit,
			m.value,
			m.sequence_number
		FROM measurements m
		JOIN material_method_analytes mma ON mma.id = m.material_method_analyte_id
		JOIN analytes a ON a.id = mma.analyte_id
		WHERE m.control_chart_id = ?
		ORDER BY mma.display_order
	`, chartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.MeasurementResult
	for rows.Next() {
		var r models.MeasurementResult
		if err := rows.Scan(&r.MMAID, &r.AnalyteName, &r.Unit, &r.Value, &r.SequenceNumber); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// For each measurement, find the applicable control limit region
	for i, r := range results {
		var ucl, lcl float64
		err := db.DB.QueryRow(`
			SELECT ucl, lcl
			FROM control_limit_regions
			WHERE material_method_analyte_id = ?
			  AND effective_from_sequence <= ?
			  AND deleted_at IS NULL
			ORDER BY effective_from_sequence DESC
			LIMIT 1
		`, r.MMAID, r.SequenceNumber).Scan(&ucl, &lcl)
		if err != nil {
			// No limits found — mark as no limits, not a fail
			results[i].NoLimits = true
			continue
		}
		results[i].UCL = &ucl
		results[i].LCL = &lcl
		results[i].Pass = r.Value >= lcl && r.Value <= ucl
	}

	return results, nil
}
