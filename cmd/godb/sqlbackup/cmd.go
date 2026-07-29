package sqlbackup

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// CmdSQLBackup is the cobra command for creating a logical database backup.
var CmdSQLBackup = &cobra.Command{
	Use:   "sqlbackup",
	Short: "Create a logical database backup",
	Long:  "Create a logical PostgreSQL or MySQL backup containing schema and data.",
	Args:  cobra.NoArgs,
	RunE:  Run,
}

var (
	db     string
	dsn    string
	output string
	force  bool
)

type runOptions struct {
	db     string
	dsn    string
	output string
	force  bool
}

// init registers sqlbackup command flags.
//
//nolint:gochecknoinits
func init() {
	CmdSQLBackup.Flags().StringVarP(&db, "db", "d", "", "database type: postgres or mysql")
	CmdSQLBackup.Flags().StringVarP(&dsn, "dsn", "s", "", "database connection string")
	CmdSQLBackup.Flags().StringVarP(&output, "output", "o", "", "backup output file")
	CmdSQLBackup.Flags().BoolVarP(&force, "force", "f", false, "replace an existing output file")
}

// Run executes the sqlbackup command.
func Run(cmd *cobra.Command, _ []string) error {
	outputPath, err := runWithOptions(commandContext(cmd), snapshotRunOptions())
	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), outputPath)
	return err
}

func snapshotRunOptions() runOptions {
	return runOptions{
		db:     db,
		dsn:    dsn,
		output: output,
		force:  force,
	}
}

func runWithOptions(ctx context.Context, opts runOptions) (string, error) {
	opts = opts.normalize()
	if err := opts.validate(); err != nil {
		return "", err
	}

	backup := newSQLBackup(opts)
	if err := backup.run(ctx); err != nil {
		return "", err
	}
	return backup.output, nil
}

func (o runOptions) normalize() runOptions {
	o.db = strings.ToLower(strings.TrimSpace(o.db))
	o.dsn = strings.TrimSpace(o.dsn)
	o.output = strings.TrimSpace(o.output)
	return o
}

func (o runOptions) validate() error {
	if o.db == "" {
		return fmt.Errorf("db cannot be empty")
	}
	if o.db != databasePostgres && o.db != databaseMySQL {
		return fmt.Errorf("unknown database type: %s", o.db)
	}
	if o.dsn == "" {
		return fmt.Errorf("dsn cannot be empty")
	}
	if o.output == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	return nil
}

func commandContext(cmd *cobra.Command) context.Context {
	if cmd == nil || cmd.Context() == nil {
		return context.Background()
	}
	return cmd.Context()
}
