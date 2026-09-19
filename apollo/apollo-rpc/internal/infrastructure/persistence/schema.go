package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ValidateSchema fails service startup when the configured database is older
// than the code. It only inspects metadata and never creates or alters objects.
func ValidateSchema(ctx context.Context, db *sql.DB, outboxTable string) error {
	if db == nil {
		return errors.New("nil database")
	}
	if outboxTable == "" {
		outboxTable = defaultOutboxTable
	}
	if !sqlIdentifier(outboxTable) {
		return ErrInvalidOutbox
	}
	required := map[string][]string{
		"user":                   {"id", "given_name", "middle_name", "family_name", "email", "login_email", "password", "password_update_time", "auth_version", "email_verified", "avatar", "birthday_year", "birthday_month", "birthday_day", "notification_email", "locate", "language", "last_login_time", "create_time", "update_time", "mark"},
		"passkey":                {"credential_id", "user_id", "display_name", "credential_json", "sign_count", "version", "created_at", "last_used_at", "is_enabled", "is_deleted"},
		"token":                  {"id", "user_id", "name", "value", "scope", "create_time", "expires_at", "is_enabled", "is_deleted"},
		"third_party":            {"id", "provider", "third_id", "user_id", "name", "create_time", "update_time"},
		"contact":                {"id", "user_id", "value", "type", "phone_region", "create_time", "is_enabled"},
		"authentication_session": {"id", "kind", "user_id", "payload", "expires_at"},
	}
	if _, reserved := required[outboxTable]; reserved {
		return errors.New("outbox table conflicts with an Apollo table")
	}
	required[outboxTable] = []string{"id", "event_id", "event_type", "schema_version", "payload", "occurred_at", "published_at", "attempt_count", "last_error"}
	rows, err := db.QueryContext(ctx, "SELECT TABLE_NAME,COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()")
	if err != nil {
		return fmt.Errorf("read columns: %w", err)
	}
	columns := make(map[string]map[string]bool)
	for rows.Next() {
		var table, column string
		if err = rows.Scan(&table, &column); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
			return fmt.Errorf("scan columns: %w", err)
		}
		if columns[table] == nil {
			columns[table] = make(map[string]bool)
		}
		columns[table][column] = true
	}
	err = rows.Err()
	if closeErr := rows.Close(); closeErr != nil {
		err = errors.Join(err, closeErr)
	}
	if err != nil {
		return fmt.Errorf("read columns: %w", err)
	}
	missing := make([]string, 0)
	for table, names := range required {
		for _, name := range names {
			if !columns[table][name] {
				missing = append(missing, table+"."+name)
			}
		}
	}
	indexRows, err := db.QueryContext(ctx, "SELECT TABLE_NAME,INDEX_NAME,COLUMN_NAME,SEQ_IN_INDEX,NON_UNIQUE FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() ORDER BY TABLE_NAME,INDEX_NAME,SEQ_IN_INDEX")
	if err != nil {
		return fmt.Errorf("read indexes: %w", err)
	}
	type indexDefinition struct {
		columns   []string
		nonUnique int
	}
	indexes := make(map[string]indexDefinition)
	for indexRows.Next() {
		var table, index, column string
		var sequence, nonUnique int
		if err = indexRows.Scan(&table, &index, &column, &sequence, &nonUnique); err != nil {
			if closeErr := indexRows.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
			return fmt.Errorf("scan indexes: %w", err)
		}
		key := table + "." + index
		definition := indexes[key]
		definition.columns = append(definition.columns, column)
		definition.nonUnique = nonUnique
		indexes[key] = definition
	}
	err = indexRows.Err()
	if closeErr := indexRows.Close(); closeErr != nil {
		err = errors.Join(err, closeErr)
	}
	if err != nil {
		return fmt.Errorf("read indexes: %w", err)
	}
	requiredIndexes := map[string]string{
		"user.user_email_unique":                "email",
		"contact.contact_value":                 "user_id,type,value",
		"third_party.provider_subject":          "provider,third_id",
		"third_party.owner_provider":            "user_id,provider",
		outboxTable + ".event_outbox_id_unique": "event_id",
	}
	for key, columns := range requiredIndexes {
		definition, ok := indexes[key]
		if !ok || definition.nonUnique != 0 || strings.Join(definition.columns, ",") != columns {
			missing = append(missing, key+"[unique]")
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	return nil
}
