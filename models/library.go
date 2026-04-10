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
    ID               int64  `json:"id"`
    MethodMaterialID int64  `json:"method_material_id"`
    MaterialID       int64  `json:"material_id"`
    MaterialName     string `json:"material_name"`
    MethodID         int64  `json:"method_id"`
    MethodName       string `json:"method_name"`
    AnalyteID        int64  `json:"analyte_id"`
    AnalyteName      string `json:"analyte_name"`
    Unit             string `json:"unit"`
    DisplayOrder     int    `json:"display_order"`
    RenderChart      bool   `json:"render_chart"`
    Active           bool   `json:"active"`
}

type MethodWithMaterials struct {
	ID        int64             `json:"id"`
	Name      string            `json:"name"`
	Materials []MaterialSummary `json:"materials"`
}

type MaterialSummary struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	MethodMaterialID int64  `json:"method_material_id"`
}

type ComboAnalyte struct {
    MMAID            int64  `json:"mma_id"`
    MethodMaterialID int64  `json:"method_material_id"`
    Name             string `json:"name"`
    Unit             string `json:"unit"`
    DisplayOrder     int    `json:"display_order"`
    RenderChart      bool   `json:"render_chart"`
}

type ControlLimitRegion struct {
	ID                    int64    `json:"id"`
	MMAID                 int64    `json:"mma_id"`
	Mean                  float64  `json:"mean"`
	UCL                   float64  `json:"ucl"`
	LCL                   float64  `json:"lcl"`
	UWL                   *float64 `json:"uwl"`
	LWL                   *float64 `json:"lwl"`
	UIL                   *float64 `json:"uil"`
	LIL                   *float64 `json:"lil"`
	EffectiveFromSequence int      `json:"effective_from_sequence"`
	CreatedBy             int64    `json:"created_by"`
	CreatedAt             string   `json:"created_at"`
}

type MeasurementResult struct {
	MMAID          int64    `json:"mma_id"`
	AnalyteName    string   `json:"analyte_name"`
	Unit           string   `json:"unit"`
	Value          float64  `json:"value"`
	SequenceNumber int      `json:"sequence_number"`
	UCL            *float64 `json:"ucl"`
	LCL            *float64 `json:"lcl"`
	Pass           bool     `json:"pass"`
	NoLimits       bool     `json:"no_limits"`
}

type ChartPoint struct {
    MeasurementID  int64	`json:"measurement_id"`
    ControlChartID int64	`json:"control_chart_id"`
    SequenceNumber int		`json:"sequence_number"`
    Value          float64	`json:"value"`
    Mean           *float64	`json:"mean"`
    UCL            *float64	`json:"ucl"`
    LCL            *float64	`json:"lcl"`
    UWL            *float64	`json:"uwl"`
    LWL            *float64	`json:"lwl"`
    UIL            *float64	`json:"uil"`
    LIL            *float64	`json:"lil"`
}
