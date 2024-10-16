package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// A simple troubleshooting utility for logging multi or single
// line statements to aid in tracking execution.

// Add easy log for variables
// Add single line quick logs that do not require flush

type logger struct {
	// lines contains lines waiting to be flushed
	lines      []*line
	linesMutex sync.Mutex
}

type line struct {
	msg    string
	src    *source
	prefix string
}

type source struct {
	File     string
	Function string
	Line     int
}

func (s source) String() string {
	return fmt.Sprintf("%s:%d", s.File, s.Line)
}

var (
	defLogger = &logger{}
	defWriter io.Writer
	defPrefix = ""
	baseDir   string
)

func init() {
	defWriter = os.Stdout
	baseDir = filepath.Dir(os.Args[0])
}

func Writer(writer io.Writer) {
	defWriter = writer
}

func Prefix(prefix string) {
	defPrefix = prefix
}

// Append appends a line to the buffer that will only be
func Append(msg string) {
	appendLine(msg)
}

// Check will append a line to the logger that can be used
// to track that a partiular line was executed.
func Check(thing any) {
	var dataB []byte
	if thing != nil {
		dataB, _ = json.Marshal(thing)
	}
	if len(dataB) > 0 {
		appendLine(fmt.Sprintf("CHECK - %s", string(dataB)))
	} else {
		appendLine("CHECK")
	}
}

func appendLine(msg string) {
	l := &line{
		msg: msg,
		src: getSource(),
	}

	genPrefix(l)

	defLogger.linesMutex.Lock()
	defer defLogger.linesMutex.Unlock()

	defLogger.lines = append(defLogger.lines, l)
}

func Flush() {
	defLogger.linesMutex.Lock()
	defer defLogger.linesMutex.Unlock()

	// preprocess line prefix len so that all messages are aligned
	maxPrefixLen := -1
	for _, l := range defLogger.lines {
		maxPrefixLen = max(maxPrefixLen, len(l.prefix))
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s-----\n", defPrefix))

	lFmt := fmt.Sprintf("%%-%ds- %%s\n", maxPrefixLen)
	for _, l := range defLogger.lines {
		sb.WriteString(fmt.Sprintf(lFmt, l.prefix, l.msg))
	}

	sb.WriteString(fmt.Sprintf("%s=====\n", defPrefix))

	fmt.Fprintln(defWriter, sb.String())

	clear()
}

func genPrefix(l *line) {
	p := fmt.Sprintf("%s%s ", defPrefix, l.src)
	l.prefix = p
}
func getSource() *source {
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	file := strings.TrimPrefix(f.File, baseDir)
	file = strings.TrimPrefix(file, "/")

	return &source{
		File:     file,
		Function: f.Function,
		Line:     f.Line,
	}
}

func clear() {
	defLogger.lines = []*line{}
}
