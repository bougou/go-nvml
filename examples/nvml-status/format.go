package main

import (
	"fmt"
	"strings"
)

const bytesPerMiB = 1024 * 1024

// formatMiB formats byte counts as integer MiB (e.g. "6204 MiB").
func formatMiB(b uint64) string {
	return fmt.Sprintf("%d MiB", b/bytesPerMiB)
}

func formatMemoryUsage(used, total uint64) string {
	return fmt.Sprintf("%s / %s (%s)", formatMiB(used), formatMiB(total), memoryUsedPercent(used, total))
}

func formatSkipped(call, reason string) string {
	return fmt.Sprintf("(skipped %s: %s)", call, reason)
}

// OutputFormat controls how status is rendered to stdout.
type OutputFormat string

const (
	FormatPlain    OutputFormat = "plain"
	FormatMarkdown OutputFormat = "markdown"
)

func parseOutputFormat(s string) (OutputFormat, error) {
	switch OutputFormat(strings.ToLower(strings.TrimSpace(s))) {
	case FormatPlain:
		return FormatPlain, nil
	case FormatMarkdown:
		return FormatMarkdown, nil
	default:
		return "", fmt.Errorf("unknown format %q (use plain or markdown)", s)
	}
}

// statusWriter renders hierarchical GPU status in plain text or Markdown.
type statusWriter interface {
	H1(title string)
	H2(title string)
	H3(title string)
	H4(title string)
	H5(title string)
	Group(title string)
	Field(key, value string)
	Text(line string)
	Blank()
	Rule()
	Skip(call, reason string)
	EndSection()
}

func newStatusWriter(format OutputFormat) statusWriter {
	switch format {
	case FormatMarkdown:
		return &markdownWriter{}
	default:
		return &plainWriter{}
	}
}

// plainWriter uses indentation and light markers to show hierarchy.
type plainWriter struct {
	indentStack []int
}

func (w *plainWriter) currentIndent() int {
	if len(w.indentStack) == 0 {
		return 0
	}
	return w.indentStack[len(w.indentStack)-1]
}

func (w *plainWriter) prefix() string {
	return strings.Repeat("  ", w.currentIndent())
}

func (w *plainWriter) println(text string) {
	fmt.Println(w.prefix() + text)
}

func (w *plainWriter) H1(title string) {
	w.indentStack = []int{0}
	w.println(title)
	w.println(strings.Repeat("=", len(title)))
}

func (w *plainWriter) H2(title string) {
	w.indentStack = []int{1}
	w.println(title)
	w.println(strings.Repeat("-", len(title)))
}

func (w *plainWriter) H3(title string) {
	w.indentStack = []int{2}
	w.println(title)
}

func (w *plainWriter) H4(title string) {
	w.indentStack = append(w.indentStack, 3)
	w.println(title)
}

func (w *plainWriter) H5(title string) {
	w.indentStack = append(w.indentStack, 4)
	w.println(title)
}

func (w *plainWriter) Group(title string) {
	w.H4(title)
}

func (w *plainWriter) EndSection() {
	if len(w.indentStack) <= 1 {
		return
	}
	w.indentStack = w.indentStack[:len(w.indentStack)-1]
}

func (w *plainWriter) Field(key, value string) {
	fmt.Printf(w.prefix()+"  %s: %s\n", key, value)
}

func (w *plainWriter) Text(line string) {
	w.println("  " + line)
}

func (w *plainWriter) printf(format string, args ...interface{}) {
	fmt.Printf(w.prefix()+format+"\n", args...)
}

func (w *plainWriter) Blank() {
	fmt.Println()
}

func (w *plainWriter) Rule() {
	w.indentStack = nil
	fmt.Println(strings.Repeat("-", 72))
}

func (w *plainWriter) Skip(call, reason string) {
	w.printf("(skipped %s: %s)", call, reason)
}

// markdownWriter renders the same structure as Markdown headings and nested lists.
type markdownWriter struct {
	listIndentStack []int
}

func (w *markdownWriter) currentListIndent() int {
	if len(w.listIndentStack) == 0 {
		return 0
	}
	return w.listIndentStack[len(w.listIndentStack)-1]
}

func (w *markdownWriter) println(text string) {
	fmt.Println(text)
}

func (w *markdownWriter) H1(title string) {
	w.listIndentStack = []int{0}
	w.println("# " + title)
}

func (w *markdownWriter) H2(title string) {
	w.listIndentStack = []int{0}
	w.println("## " + title)
}

func (w *markdownWriter) H3(title string) {
	w.listIndentStack = []int{0}
	w.println("### " + title)
}

func (w *markdownWriter) H4(title string) {
	w.listIndentStack = append(w.listIndentStack, 1)
	w.println("#### " + title)
}

func (w *markdownWriter) H5(title string) {
	w.listIndentStack = append(w.listIndentStack, 2)
	w.println("##### " + title)
}

func (w *markdownWriter) Group(title string) {
	w.listItem(fmt.Sprintf("**%s**", title))
	w.listIndentStack = append(w.listIndentStack, w.currentListIndent()+1)
}

func (w *markdownWriter) EndSection() {
	if len(w.listIndentStack) <= 1 {
		return
	}
	w.listIndentStack = w.listIndentStack[:len(w.listIndentStack)-1]
}

func (w *markdownWriter) Field(key, value string) {
	w.listItem(fmt.Sprintf("**%s**: %s", key, value))
}

func (w *markdownWriter) Text(line string) {
	w.listItem(line)
}

func (w *markdownWriter) listItem(text string) {
	w.println(strings.Repeat("  ", w.currentListIndent()) + "- " + text)
}

func (w *markdownWriter) Blank() {
	fmt.Println()
}

func (w *markdownWriter) Rule() {
	w.println("---")
}

func (w *markdownWriter) Skip(call, reason string) {
	w.listItem(fmt.Sprintf("_skipped %s: %s_", call, reason))
}
