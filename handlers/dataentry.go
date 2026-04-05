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
			met.id AS method_id,
			met.name AS method_name,
			mat.id AS material_id,
			mat.name AS material_name
		FROM methods met
		JOIN material_method_analytes mma ON mma.method_id = met.id
		JOIN materials mat ON mat.id = mma.material_id
		WHERE mma.active = 1
		GROUP BY met.id, mat.id
		ORDER BY met.name, mat.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methodMap []models.MethodWithMaterials
	index := map[int64]int{}

	for rows.Next() {
		var methodID, materialID int64
		var methodName, materialName string
		if err := rows.Scan(&methodID, &methodName, &materialID, &materialName); err != nil {
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
			ID:   materialID,
			Name: materialName,
		})
	}
	return methodMap, rows.Err()
}

// GetAnalytesForCombo returns the ordered analytes for a method+material combo.
func (h *DataEntryHandler) GetAnalytesForCombo(methodID, materialID int64) ([]models.ComboAnalyte, error) {
	rows, err := db.DB.Query(`
		SELECT
			mma.id,
			a.name,
			a.unit,
			mma.display_order
		FROM material_method_analytes mma
		JOIN analytes a ON a.id = mma.analyte_id
		WHERE mma.method_id = ? AND mma.material_id = ? AND mma.active = 1
		ORDER BY mma.display_order
	`, methodID, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytes []models.ComboAnalyte
	for rows.Next() {
		var a models.ComboAnalyte
		if err := rows.Scan(&a.MMAID, &a.Name, &a.Unit, &a.DisplayOrder); err != nil {
			return nil, err
		}
		analytes = append(analytes, a)
	}
	return analytes, rows.Err()
}

// SaveChart creates a control_chart and its measurements in one transaction.
// Values map is mma_id -> value, missing keys mean no measurement for that analyte.
func (h *DataEntryHandler) SaveChart(methodID, materialID, technicianID int64, values map[string]float64) (int64, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	res, err := tx.Exec(`
		INSERT INTO control_charts (material_id, method_id, technician_id, created_at, locked_at)
		VALUES (?, ?, ?, ?, ?)
	`, materialID, methodID, technicianID, now, now)
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
		_, err := tx.Exec(`
			INSERT INTO measurements (control_chart_id, material_method_analyte_id, value, sequence_order, entered_by, entered_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, chartID, mmaID, value, seq, technicianID, now)
		if err != nil {
			return 0, fmt.Errorf("insert measurement: %w", err)
		}
		seq++
	}

	return chartID, tx.Commit()
}
