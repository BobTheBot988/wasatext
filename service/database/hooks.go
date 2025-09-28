package database

import (
	"database/sql"
	"fmt"
)

// CreateSanitizeHook creates a hook that validates specific tables/columns
func (db *appdbimpl) CreateSanitizeHook(validations map[string][]string) PreCommitHook {
	return func(tx *sql.Tx) error {
		for table, columns := range validations {
			for _, column := range columns {
				query := fmt.Sprintf("SELECT rowid, %s FROM %s WHERE %s IS NOT NULL", column, table, column)
				rows, err := tx.Query(query)
				if err != nil {
					return fmt.Errorf("failed to query %s.%s: %w", table, column, err)
				}
				if err = rows.Err(); err != nil {
					return err
				}
				defer rows.Close()

				for rows.Next() {
					var rowid int64
					var value string
					if err = rows.Scan(&rowid, &value); err != nil {
						return err
					}

					if _, err = db.SanitizeString(value); err != nil {
						return fmt.Errorf("invalid %s in %s (row %d): %w", column, table, rowid, err)
					}
				}
			}
		}
		return nil
	}
}
