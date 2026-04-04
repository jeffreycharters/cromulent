package db

import "fmt"

func migrate() error {
	current, err := getUserVersion()
	if err != nil {
		return fmt.Errorf("get user_version: %w", err)
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		if _, err := DB.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %d: %w", m.version, err)
		}
		if err := setUserVersion(m.version); err != nil {
			return fmt.Errorf("set user_version %d: %w", m.version, err)
		}
	}
	return nil
}

func getUserVersion() (int, error) {
	var v int
	err := DB.QueryRow(`PRAGMA user_version`).Scan(&v)
	return v, err
}

func setUserVersion(v int) error {
	// user_version doesn't support prepared statement placeholders
	_, err := DB.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, v))
	return err
}

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'technician',
    active          INTEGER NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS materials (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS methods (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS analytes (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    TEXT NOT NULL,
    unit    TEXT NOT NULL,
    UNIQUE(name, unit)
);

CREATE TABLE IF NOT EXISTS material_method_analytes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    material_id   INTEGER NOT NULL REFERENCES materials(id),
    method_id     INTEGER NOT NULL REFERENCES methods(id),
    analyte_id    INTEGER NOT NULL REFERENCES analytes(id),
    display_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(material_id, method_id, analyte_id)
);

CREATE TABLE IF NOT EXISTS control_limit_regions (
    id                          INTEGER PRIMARY KEY AUTOINCREMENT,
    material_method_analyte_id  INTEGER NOT NULL REFERENCES material_method_analytes(id),
    mean                        REAL NOT NULL,
    ucl                         REAL NOT NULL,
    lcl                         REAL NOT NULL,
    uwl                         REAL,
    lwl                         REAL,
    uil                         REAL,
    lil                         REAL,
    effective_from_sequence     INTEGER NOT NULL,
    created_by                  INTEGER NOT NULL REFERENCES users(id),
    created_at                  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS control_charts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    material_id   INTEGER NOT NULL REFERENCES materials(id),
    method_id     INTEGER NOT NULL REFERENCES methods(id),
    batch_id      TEXT,
    technician_id INTEGER NOT NULL REFERENCES users(id),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    locked_at     DATETIME
);

CREATE TABLE IF NOT EXISTS chart_metadata_fields (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL UNIQUE,
    required      INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS chart_metadata_values (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    control_chart_id INTEGER NOT NULL REFERENCES control_charts(id),
    field_id         INTEGER NOT NULL REFERENCES chart_metadata_fields(id),
    value            TEXT NOT NULL,
    UNIQUE(control_chart_id, field_id)
);

CREATE TABLE IF NOT EXISTS measurements (
    id                         INTEGER PRIMARY KEY AUTOINCREMENT,
    control_chart_id           INTEGER NOT NULL REFERENCES control_charts(id),
    material_method_analyte_id INTEGER NOT NULL REFERENCES material_method_analytes(id),
    value                      REAL NOT NULL,
    sequence_order             INTEGER NOT NULL,
    entered_by                 INTEGER NOT NULL REFERENCES users(id),
    entered_at                 DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS comments (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    control_chart_id INTEGER NOT NULL REFERENCES control_charts(id),
    measurement_id   INTEGER REFERENCES measurements(id),
    comment_type     TEXT NOT NULL,
    text             TEXT NOT NULL,
    user_id          INTEGER NOT NULL REFERENCES users(id),
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS spc_rule_sets (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    beyond_sigma_enabled  INTEGER NOT NULL DEFAULT 1,
    beyond_sigma_n        INTEGER NOT NULL DEFAULT 3,
    run_trend_enabled     INTEGER NOT NULL DEFAULT 1,
    run_trend_n           INTEGER NOT NULL DEFAULT 6,
    one_side_enabled      INTEGER NOT NULL DEFAULT 1,
    one_side_n            INTEGER NOT NULL DEFAULT 7,
    effective_from_date   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by            INTEGER NOT NULL REFERENCES users(id),
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,
	},
}
