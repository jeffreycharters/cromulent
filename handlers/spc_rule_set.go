package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"database/sql"
	"errors"
	"fmt"
)

type SPCRuleSetHandler struct{}

func NewSPCRuleSetHandler() *SPCRuleSetHandler {
	return &SPCRuleSetHandler{}
}

func (h *SPCRuleSetHandler) GetGlobalRuleSet() (*models.SPCRuleSet, error) {
	return scanRuleSet(db.DB.QueryRow(`
		SELECT id, method_material_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by, created_at
		FROM spc_rule_sets
		WHERE method_material_id IS NULL
		ORDER BY id DESC
		LIMIT 1`,
	))
}

func (h *SPCRuleSetHandler) GetRuleSetsForMMA(methodMaterialID int64) ([]models.SPCRuleSet, error) {
	rows, err := db.DB.Query(`
		SELECT id, method_material_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by, created_at
		FROM spc_rule_sets
		WHERE method_material_id = ?
		ORDER BY effective_from_sequence ASC`,
		methodMaterialID,
	)
	if err != nil {
		return nil, fmt.Errorf("get rule sets for combo: %w", err)
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
		WHERE method_material_id IS NULL`,
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
	methodMaterialID int64,
	effectiveFromSequence int64,
	beyondLimitsEnabled bool,
	warningLimitsEnabled bool, warningConsecutiveCount, warningTriggerCount int,
	trendEnabled bool, trendConsecutiveCount int,
	oneSideEnabled bool, oneSideConsecutiveCount int,
	createdBy int64,
) error {
	_, err := db.DB.Exec(`
		INSERT INTO spc_rule_sets (
			method_material_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		methodMaterialID, effectiveFromSequence,
		beyondLimitsEnabled,
		warningLimitsEnabled, warningConsecutiveCount, warningTriggerCount,
		trendEnabled, trendConsecutiveCount,
		oneSideEnabled, oneSideConsecutiveCount,
		createdBy,
	)
	if err != nil {
		return fmt.Errorf("create combo rule set: %w", err)
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
		&rs.ID, &rs.MethodMaterialID, &rs.EffectiveFromSequence,
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

func (h *SPCRuleSetHandler) GetEffectiveRuleSetForCombo(methodMaterialID int64, sequence int64) (*models.SPCRuleSet, error) {
	row := db.DB.QueryRow(`
		SELECT id, method_material_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by, created_at
		FROM spc_rule_sets
		WHERE method_material_id = ?
		  AND effective_from_sequence <= ?
		ORDER BY effective_from_sequence DESC
		LIMIT 1`,
		methodMaterialID, sequence,
	)
	rs, err := scanRuleSetRow(row)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get effective rule set for combo: %w", err)
	}
	if err == nil {
		return &rs, nil
	}

	globalRow := db.DB.QueryRow(`
		SELECT id, method_material_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by, created_at
		FROM spc_rule_sets
		WHERE method_material_id IS NULL
		ORDER BY id DESC
		LIMIT 1`,
	)
	globalRS, err := scanRuleSetRow(globalRow)
	if err != nil {
		return nil, fmt.Errorf("get global rule set fallback: %w", err)
	}
	return &globalRS, nil
}
