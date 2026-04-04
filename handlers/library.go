package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"fmt"
)

type LibraryHandler struct{}

func NewLibraryHandler() *LibraryHandler {
	return &LibraryHandler{}
}

// --- Analytes ---

func (h *LibraryHandler) CreateAnalyte(name, unit string) error {
	if name == "" {
		return fmt.Errorf("name required")
	}
	_, err := db.DB.Exec(
		`INSERT INTO analytes (name, unit) VALUES (?, ?)`,
		name, unit,
	)
	return err
}

func (h *LibraryHandler) ListAnalytes() ([]models.Analyte, error) {
	rows, err := db.DB.Query(`SELECT id, name, unit FROM analytes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytes []models.Analyte
	for rows.Next() {
		var a models.Analyte
		if err := rows.Scan(&a.ID, &a.Name, &a.Unit); err != nil {
			return nil, err
		}
		analytes = append(analytes, a)
	}
	return analytes, rows.Err()
}

// --- Methods ---

func (h *LibraryHandler) CreateMethod(name, description string) error {
	if name == "" {
		return fmt.Errorf("name required")
	}
	_, err := db.DB.Exec(
		`INSERT INTO methods (name, description) VALUES (?, ?)`,
		name, description,
	)
	return err
}

func (h *LibraryHandler) ListMethods() ([]models.Method, error) {
	rows, err := db.DB.Query(`SELECT id, name, description FROM methods ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []models.Method
	for rows.Next() {
		var m models.Method
		if err := rows.Scan(&m.ID, &m.Name, &m.Description); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, rows.Err()
}

// --- Materials ---

func (h *LibraryHandler) CreateMaterial(name, description string) error {
	if name == "" {
		return fmt.Errorf("name required")
	}
	_, err := db.DB.Exec(
		`INSERT INTO materials (name, description) VALUES (?, ?)`,
		name, description,
	)
	return err
}

func (h *LibraryHandler) ListMaterials() ([]models.Material, error) {
	rows, err := db.DB.Query(`SELECT id, name, description FROM materials ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []models.Material
	for rows.Next() {
		var m models.Material
		if err := rows.Scan(&m.ID, &m.Name, &m.Description); err != nil {
			return nil, err
		}
		materials = append(materials, m)
	}
	return materials, rows.Err()
}
