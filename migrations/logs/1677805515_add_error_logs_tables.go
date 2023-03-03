package logs

import (
	"github.com/pocketbase/dbx"
)

func init() {
	LogsMigrations.Register(func(db dbx.Builder) (err error) {
		for _, query := range []string{
			`
			CREATE TABLE {{_errors}} (
				[[id]]        TEXT PRIMARY KEY NOT NULL,
				[[file]]      TEXT DEFAULT "" NOT NULL,
				[[line]]      INTEGER DEFAULT 0 NOT NULL,
				[[error]]     TEXT DEFAULT "" NOT NULL,
				[[fatal]]     BOOL DEFAULT 0 NOT NULL,
				[[meta]]      JSON DEFAULT "{}" NOT NULL,
				[[created]]   TEXT DEFAULT "" NOT NULL,
				[[updated]]   TEXT DEFAULT "" NOT NULL
			);

			CREATE INDEX _errors_file_idx on {{_errors}} ([[file]]);
			CREATE INDEX _errors_created_hour_idx on {{_errors}} (strftime('%Y-%m-%d %H:00:00', [[created]]));
		`, `
			CREATE TABLE {{_logs}} (
				[[id]]        TEXT PRIMARY KEY NOT NULL,
				[[level]]     TEXT DEFAULT "" NOT NULL,
				[[message]]   TEXT DEFAULT "get" NOT NULL,
				[[meta]]      JSON DEFAULT "{}" NOT NULL,
				[[created]]   TEXT DEFAULT "" NOT NULL,
				[[updated]]   TEXT DEFAULT "" NOT NULL
			);

			CREATE INDEX _logs_level_idx on {{_logs}} ([[level]]);
			CREATE INDEX _logs_created_hour_idx on {{_logs}} (strftime('%Y-%m-%d %H:00:00', [[created]]));
		`,
		} {
			if _, err = db.NewQuery(query).Execute(); err != nil {
				return err
			}

		}

		return nil
	}, func(db dbx.Builder) error {
		for _, table := range []string{
			"_logs",
			"_errors",
		} {
			if _, err := db.DropTable(table).Execute(); err != nil {
				return err
			}
		}

		return nil
	})
}
