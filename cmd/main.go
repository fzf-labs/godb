package main

import (
	"log"

	"github.com/fzf-labs/godb/cmd/ormgen"
	"github.com/fzf-labs/godb/cmd/sqldump"
	"github.com/fzf-labs/godb/cmd/sqltopb"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "godb",
	Short:   "godb: an db toolkit",
	Long:    `godb: an db toolkit`,
	Version: "v0.0.1",
}

func init() {
	rootCmd.AddCommand(ormgen.CmdOrmGen)
	rootCmd.AddCommand(sqldump.CmdSQLDump)
	rootCmd.AddCommand(sqltopb.CmdSQLToPb)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
