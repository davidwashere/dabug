package dabug

import (
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
	defLogger  = &logger{}
	defWriter  io.Writer
	defPrefix  = ""
	autoflush  = true
	sectionBeg = "-----"
	sectionEnd = "====="
	// baseDir   string
)

func init() {
	defWriter = os.Stdout
	// baseDir = filepath.Dir(os.Args[0])
}

func Writer(writer io.Writer) {
	defWriter = writer
}

func Prefix(prefix string) {
	defPrefix = prefix
}

func AutoFlush(flush bool) {
	autoflush = flush
	if len(defLogger.lines) > 0 {
		Flush()
	}
}

func Msg(msg string) {
	appendMsg(msg)
}

func Here() {
	appendEmpty()
}

// Check will append a line to the logger that can be used
// to track that a partiular line was executed.
func Objs(things ...any) {
	var msgs []string
	for i, t := range things {
		msg := fmt.Sprintf("[%d] %#v", i, t)
		msgs = append(msgs, msg)
	}
	appendMsg(strings.Join(msgs, ", "))
}

func Flush() {
	defLogger.linesMutex.Lock()
	defer defLogger.linesMutex.Unlock()

	if len(defLogger.lines) == 0 {
		// Nothing to do
		return
	}

	// preprocess line prefix len so that all messages are aligned
	maxPrefixLen := -1
	for _, l := range defLogger.lines {
		maxPrefixLen = max(maxPrefixLen, len(l.prefix))
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s%s\n", defPrefix, sectionBeg))

	lFmt := fmt.Sprintf("%%-%ds", maxPrefixLen)
	for _, l := range defLogger.lines {
		sb.WriteString(lineStr(lFmt, l) + "\n")
	}

	sb.WriteString(fmt.Sprintf("%s%s\n", defPrefix, sectionEnd))

	fmt.Fprint(defWriter, sb.String())

	clear()
}

func flushLine(l *line) {
	msg := lineStr("%s", l)
	fmt.Fprintf(defWriter, "%s\n", msg)
}

func lineStr(lFmt string, l *line) string {
	var msg string

	if len(l.msg) == 0 {
		msg = fmt.Sprintf(lFmt, l.prefix)
	} else {
		suffix := "- %s"
		msg = fmt.Sprintf(lFmt+suffix, l.prefix, l.msg)
	}

	return msg
}

func appendLine(l *line) {
	genPrefix(l)

	if autoflush {
		flushLine(l)
		return
	}

	defLogger.linesMutex.Lock()
	defer defLogger.linesMutex.Unlock()

	defLogger.lines = append(defLogger.lines, l)
}

func appendEmpty() {
	l := &line{src: getSource()}
	appendLine(l)
}

func appendMsg(msg string) {
	l := &line{
		msg: msg,
		src: getSource(),
	}
	appendLine(l)
}

func genPrefix(l *line) {
	p := fmt.Sprintf("%s%s ", defPrefix, l.src)
	l.prefix = p
}

func getSource() *source {
	var pc uintptr
	var pcs [1]uintptr

	// skip [runtime.Callers, this function, this function's caller]
	// skip
	// 1. runtime.Callers
	// 2. getSource
	// 3. appendLine
	// 4. Msg, Objs, etc.
	runtime.Callers(4, pcs[:])
	pc = pcs[0]

	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	_, fpath, _, _ := runtime.Caller(3)

	var file string
	file = strings.TrimPrefix(f.File, filepath.Dir(fpath))
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
