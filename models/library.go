package models

type Analyte struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type Method struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Material struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MMAEntry struct {
	ID           int64  `json:"id"`
	MaterialID   int64  `json:"material_id"`
	MaterialName string `json:"material_name"`
	MethodID     int64  `json:"method_id"`
	MethodName   string `json:"method_name"`
	AnalyteID    int64  `json:"analyte_id"`
	AnalyteName  string `json:"analyte_name"`
	Unit         string `json:"unit"`
	DisplayOrder int    `json:"display_order"`
}
