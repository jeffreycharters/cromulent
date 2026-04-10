package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"fmt"
)

type MMAHandler struct{}

func NewMMAHandler() *MMAHandler {
	return &MMAHandler{}
}

// AddAnalyteToMMA creates a material_method_analyte row.
// display_order determines column order in the data entry grid.
func (h *MMAHandler) AddAnalyteToMMA(methodMaterialID, analyteID int64, displayOrder int) error {
	_, err := db.DB.Exec(
		`INSERT INTO material_method_analytes (method_material_id, analyte_id, display_order)
		 VALUES (?, ?, ?)`,
		methodMaterialID, analyteID, displayOrder,
	)
	return err
}

func (h *MMAHandler) EnsureMethodMaterial(methodID, materialID int64) (int64, error) {
	_, err := db.DB.Exec(
		`INSERT OR IGNORE INTO method_materials (method_id, material_id) VALUES (?, ?)`,
		methodID, materialID,
	)
	if err != nil {
		return 0, err
	}
	var id int64
	err = db.DB.QueryRow(
		`SELECT id FROM method_materials WHERE method_id = ? AND material_id = ?`,
		methodID, materialID,
	).Scan(&id)
	return id, err
}

// ListMMAsForMethod returns all material+analyte combos for a given method,
// grouped such that the frontend can build the sidebar and grid.
func (h *MMAHandler) ListMMAsForMethod(methodID int64) ([]models.MMAEntry, error) {
	rows, err := db.DB.Query(`
		SELECT
			mma.id,
			mm.id,
			mm.material_id,
			mat.name,
			mm.method_id,
			met.name,
			mma.analyte_id,
			a.name,
			a.unit,
			mma.display_order,
			mma.render_chart,
			mm.active
		FROM material_method_analytes mma
		JOIN method_materials mm ON mm.id = mma.method_material_id
		JOIN materials mat ON mat.id = mm.material_id
		JOIN methods met ON met.id = mm.method_id
		JOIN analytes a ON a.id = mma.analyte_id
		WHERE mm.method_id = ?
		ORDER BY mat.name, mma.display_order
	`, methodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.MMAEntry
	for rows.Next() {
		var e models.MMAEntry
		if err := rows.Scan(
			&e.ID, &e.MethodMaterialID, &e.MaterialID, &e.MaterialName,
			&e.MethodID, &e.MethodName,
			&e.AnalyteID, &e.AnalyteName, &e.Unit,
			&e.DisplayOrder, &e.RenderChart, &e.Active,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListAllMMAs returns everything — used to populate the sidebar.
func (h *MMAHandler) ListAllMMAs() ([]models.MMAEntry, error) {
	rows, err := db.DB.Query(`
    SELECT
        mma.id,
        mma.material_id,
        mat.name AS material_name,
        mma.method_id,
        met.name AS method_name,
        mma.analyte_id,
        a.name AS analyte_name,
        a.unit,
        mma.display_order,
        mma.active
    FROM material_method_analytes mma
    JOIN materials mat ON mat.id = mma.material_id
    JOIN methods met ON met.id = mma.method_id
    JOIN analytes a ON a.id = mma.analyte_id
    ORDER BY met.name, mat.name, mma.display_order
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.MMAEntry
	for rows.Next() {
		var e models.MMAEntry
		if err := rows.Scan(
			&e.ID, &e.MaterialID, &e.MaterialName,
			&e.MethodID, &e.MethodName,
			&e.AnalyteID, &e.AnalyteName, &e.Unit, &e.DisplayOrder, &e.Active,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// RemoveAnalyteFromMMA deletes a single material_method_analyte row by id.
func (h *MMAHandler) RemoveAnalyteFromMMA(id int64) error {
	res, err := db.DB.Exec(`DELETE FROM material_method_analytes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (h *MMAHandler) UpdateDisplayOrders(ids []int64, orders []int) error {
	if len(ids) != len(orders) {
		return fmt.Errorf("ids and orders length mismatch")
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		if _, err := tx.Exec(
			`UPDATE material_method_analytes SET display_order = ? WHERE id = ?`,
			orders[i], id,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (h *MMAHandler) ListUsedMMAIDs() ([]int64, error) {
	rows, err := db.DB.Query(`SELECT DISTINCT material_method_analyte_id FROM measurements`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (h *MMAHandler) DeactivateCombo(methodMaterialID int64) error {
	_, err := db.DB.Exec(
		`UPDATE method_materials SET active = 0 WHERE id = ?`,
		methodMaterialID,
	)
	return err
}

func (h *MMAHandler) ActivateCombo(methodMaterialID int64) error {
	_, err := db.DB.Exec(
		`UPDATE method_materials SET active = 1 WHERE id = ?`,
		methodMaterialID,
	)
	return err
}

func (h *MMAHandler) SetRenderChart(mmaID int64, render bool) error {
	val := 0
	if render {
		val = 1
	}
	_, err := db.DB.Exec(
		`UPDATE material_method_analytes SET render_chart = ? WHERE id = ?`,
		val, mmaID,
	)
	return err
}
