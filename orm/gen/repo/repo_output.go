package repo

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"
)

var formatErrorLinePattern = regexp.MustCompile(`:(\d+)(?::\d+)?:`)

// output 导出文件
func (r *Repo) output(fileName string, content []byte) error {
	result, err := imports.Process(fileName, content, nil)
	if err != nil {
		if errLine, ok := extractFormatErrorLine(err); ok {
			return fmt.Errorf("cannot format file at line %d: %w\n%s", errLine, err, formatErrorContext(content, errLine))
		}
		return fmt.Errorf("cannot format file: %w", err)
	}
	return os.WriteFile(fileName, result, 0600)
}

func extractFormatErrorLine(err error) (int, bool) {
	if err == nil {
		return 0, false
	}
	match := formatErrorLinePattern.FindStringSubmatch(err.Error())
	if len(match) < 2 {
		return 0, false
	}
	errLine, convErr := strconv.Atoi(match[1])
	if convErr != nil || errLine <= 0 {
		return 0, false
	}
	return errLine, true
}

func formatErrorContext(content []byte, errLine int) string {
	lines := strings.Split(string(content), "\n")
	startLine, endLine := errLine-5, errLine+5
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	var builder strings.Builder
	for lineNo := startLine; lineNo <= endLine; lineNo++ {
		builder.WriteString(fmt.Sprintf("%d %s\n", lineNo, lines[lineNo-1]))
	}
	return strings.TrimRight(builder.String(), "\n")
}
