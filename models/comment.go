package models

type CommentResponse struct {
	ID             int64  `json:"id"`
	ControlChartID int64  `json:"control_chart_id"`
	MeasurementID  *int64 `json:"measurement_id"`
	Text           string `json:"text"`
	UserID         int64  `json:"user_id"`
	Username       string `json:"username"`
	CreatedAt      string `json:"created_at"`
}
