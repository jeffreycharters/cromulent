package db

import "fmt"

func EnsureGlobalRuleSet(createdBy int64) error {
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM spc_rule_sets WHERE method_material_id IS NULL`).Scan(&count); err != nil {
		return fmt.Errorf("check global rule set: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err := DB.Exec(`
		INSERT INTO spc_rule_sets (
			method_material_id, effective_from_sequence,
			beyond_limits_enabled,
			warning_limits_enabled, warning_consecutive_count, warning_trigger_count,
			trend_enabled, trend_consecutive_count,
			one_side_enabled, one_side_consecutive_count,
			created_by
		) VALUES (NULL, NULL, 1, 1, 3, 2, 1, 6, 1, 8, ?)`,
		createdBy,
	)
	if err != nil {
		return fmt.Errorf("seed global rule set: %w", err)
	}
	return nil
}
