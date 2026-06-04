package main

import (
	"log"
	"runtime/debug"

	"github.com/fzf-labs/godb/cmd/godb/ormgen"
	"github.com/fzf-labs/godb/cmd/godb/sqldump"
	"github.com/fzf-labs/godb/cmd/godb/sqltopb"
	"github.com/spf13/cobra"
)

// version 可在发布构建时通过 -ldflags "-X main.version=vX.Y.Z" 注入。
var version = "dev"

var logFatal = log.Fatal

var rootCmd = &cobra.Command{
	Use:     "godb",
	Short:   "godb: an db toolkit",
	Long:    `godb: an db toolkit`,
	Version: commandVersion(),
}

func init() {
	rootCmd.AddCommand(ormgen.CmdOrmGen)
	rootCmd.AddCommand(sqldump.CmdSQLDump)
	rootCmd.AddCommand(sqltopb.CmdSQLToPb)
}

func commandVersion() string {
	if version != "" && version != "dev" {
		return version
	}
	// go install module@version 时优先使用 Go build info 中记录的模块版本。
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	runMain(rootCmd.Execute)
}

func runMain(execute func() error) {
	if err := execute(); err != nil {
		logFatal(err)
	}
}
