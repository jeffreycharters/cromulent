package handlers

import (
	"fmt"
	"time"

	"cromulent/db"
	"cromulent/models"
)

type CommentsHandler struct{}

func NewCommentsHandler() *CommentsHandler {
	return &CommentsHandler{}
}

func (h *CommentsHandler) AddComment(chartID int64, measurementID *int64, text string, userID int64) error {
	if text == "" {
		return fmt.Errorf("comment text cannot be empty")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.DB.Exec(`
		INSERT INTO comments (control_chart_id, measurement_id, text, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, chartID, measurementID, text, userID, now)
	return err
}

func (h *CommentsHandler) GetCommentsForChart(chartID int64) ([]models.CommentResponse, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.control_chart_id, c.measurement_id, c.text, c.user_id, u.username, c.created_at
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.control_chart_id = ?
		ORDER BY c.created_at ASC
	`, chartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CommentResponse
	for rows.Next() {
		var r models.CommentResponse
		if err := rows.Scan(&r.ID, &r.ControlChartID, &r.MeasurementID, &r.Text, &r.UserID, &r.Username, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}

func (h *CommentsHandler) GetCommentsForCombo(methodMaterialID int64) ([]models.CommentResponse, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.control_chart_id, c.measurement_id, c.text, c.user_id, u.username, c.created_at
		FROM comments c
		JOIN users u ON u.id = c.user_id
		JOIN control_charts cc ON cc.id = c.control_chart_id
		WHERE cc.method_material_id = ?
		ORDER BY c.created_at ASC
	`, methodMaterialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CommentResponse
	for rows.Next() {
		var r models.CommentResponse
		if err := rows.Scan(&r.ID, &r.ControlChartID, &r.MeasurementID, &r.Text, &r.UserID, &r.Username, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
