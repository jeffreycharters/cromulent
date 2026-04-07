package models

type SPCRuleSet struct {
	ID                      int64  `json:"id"`
	MaterialMethodID        *int64 `json:"materialMethodId"`
	EffectiveFromSequence   *int64 `json:"effectiveFromSequence"`
	BeyondLimitsEnabled     bool   `json:"beyondLimitsEnabled"`
	WarningLimitsEnabled    bool   `json:"warningLimitsEnabled"`
	WarningConsecutiveCount int    `json:"warningConsecutiveCount"`
	WarningTriggerCount     int    `json:"warningTriggerCount"`
	TrendEnabled            bool   `json:"trendEnabled"`
	TrendConsecutiveCount   int    `json:"trendConsecutiveCount"`
	OneSideEnabled          bool   `json:"oneSideEnabled"`
	OneSideConsecutiveCount int    `json:"oneSideConsecutiveCount"`
	CreatedBy               int64  `json:"createdBy"`
	CreatedAt               string `json:"createdAt"`
}
