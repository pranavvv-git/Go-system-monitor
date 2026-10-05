package prometheus

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/aadityya4real/Go-system-monitor/internal/collector"
)

func Render(w io.Writer, metrics []collector.Metric) error {
	seen := map[string]bool{}

	for _, metric := range metrics {
		if !seen[metric.Name] {
			if _, err := fmt.Fprintf(w, "# HELP %s %s\n", metric.Name, sanitizeHelp(metric.Help)); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "# TYPE %s %s\n", metric.Name, metric.Type); err != nil {
				return err
			}
			seen[metric.Name] = true
		}

		if _, err := fmt.Fprintf(w, "%s%s %s\n", metric.Name, renderLabels(metric.Labels), strconv.FormatFloat(metric.Value, 'f', -1, 64)); err != nil {
			return err
		}
	}

	return nil
}

func renderLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, escapeLabel(labels[key])))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

func sanitizeHelp(help string) string {
	return strings.ReplaceAll(help, "\n", " ")
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}
