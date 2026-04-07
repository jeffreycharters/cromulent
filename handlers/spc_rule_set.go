package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"fmt"
)

type SPCRuleSetHandler struct{}

func NewSPCRuleSetHandler() *SPCRuleSetHandler {
	return &SPCRuleSetHandler{}
}

func (h *SPCRuleSetHandler) GetGlobalRuleSet() (*models.SPCRuleSet, error) {
	return scanRuleSet(db.DB.QueryRow(`
		SELECT id, material_method_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by, created_at
		FROM spc_rule_sets
		WHERE material_method_id IS NULL
		ORDER BY id DESC
		LIMIT 1`,
	))
}

func (h *SPCRuleSetHandler) GetRuleSetsForMMA(mmaID int64) ([]models.SPCRuleSet, error) {
	rows, err := db.DB.Query(`
		SELECT id, material_method_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by, created_at
		FROM spc_rule_sets
		WHERE material_method_id = ?
		ORDER BY effective_from_sequence ASC`,
		mmaID,
	)
	if err != nil {
		return nil, fmt.Errorf("get rule sets for mma: %w", err)
	}
	defer rows.Close()

	var results []models.SPCRuleSet
	for rows.Next() {
		rs, err := scanRuleSetRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, rs)
	}
	return results, nil
}

func (h *SPCRuleSetHandler) UpdateGlobalRuleSet(
	beyondLimitsEnabled bool,
	warningLimitsEnabled bool, warningConsecutiveCount, warningTriggerCount int,
	trendEnabled bool, trendConsecutiveCount int,
	oneSideEnabled bool, oneSideConsecutiveCount int,
	createdBy int64,
) error {
	_, err := db.DB.Exec(`
		UPDATE spc_rule_sets SET
			beyond_limits_enabled = ?,
			warning_limits_enabled = ?, warning_consecutive_count = ?, warning_trigger_count = ?,
			trend_enabled = ?, trend_consecutive_count = ?,
			one_side_enabled = ?, one_side_consecutive_count = ?
		WHERE material_method_id IS NULL`,
		beyondLimitsEnabled,
		warningLimitsEnabled, warningConsecutiveCount, warningTriggerCount,
		trendEnabled, trendConsecutiveCount,
		oneSideEnabled, oneSideConsecutiveCount,
	)
	if err != nil {
		return fmt.Errorf("update global rule set: %w", err)
	}
	return nil
}

func (h *SPCRuleSetHandler) CreateMMARuleSet(
	mmaID int64,
	effectiveFromSequence int64,
	beyondLimitsEnabled bool,
	warningLimitsEnabled bool, warningConsecutiveCount, warningTriggerCount int,
	trendEnabled bool, trendConsecutiveCount int,
	oneSideEnabled bool, oneSideConsecutiveCount int,
	createdBy int64,
) error {
	_, err := db.DB.Exec(`
		INSERT INTO spc_rule_sets (
			material_method_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		mmaID, effectiveFromSequence,
		beyondLimitsEnabled,
		warningLimitsEnabled, warningConsecutiveCount, warningTriggerCount,
		trendEnabled, trendConsecutiveCount,
		oneSideEnabled, oneSideConsecutiveCount,
		createdBy,
	)
	if err != nil {
		return fmt.Errorf("create mma rule set: %w", err)
	}
	return nil
}

// --- scan helpers ---

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRuleSet(row rowScanner) (*models.SPCRuleSet, error) {
	rs, err := scanRuleSetRow(row)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

func scanRuleSetRow(row rowScanner) (models.SPCRuleSet, error) {
	var rs models.SPCRuleSet
	err := row.Scan(
		&rs.ID, &rs.MaterialMethodID, &rs.EffectiveFromSequence,
		&rs.BeyondLimitsEnabled,
		&rs.WarningLimitsEnabled, &rs.WarningConsecutiveCount, &rs.WarningTriggerCount,
		&rs.TrendEnabled, &rs.TrendConsecutiveCount,
		&rs.OneSideEnabled, &rs.OneSideConsecutiveCount,
		&rs.CreatedBy, &rs.CreatedAt,
	)
	if err != nil {
		return rs, fmt.Errorf("scan rule set: %w", err)
	}
	return rs, nil
}
