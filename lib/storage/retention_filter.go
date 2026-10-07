package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/prompb"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/promrelabel"
)

// RetentionFilter applies a retention period to time series matching a
// Prometheus-compatible series selector.
type RetentionFilter struct {
	matchExpr      string
	retentionMsecs int64
	ie             promrelabel.IfExpression
}

// NewRetentionFilter creates a retention filter for matchExpr.
func NewRetentionFilter(matchExpr string, retention time.Duration) (RetentionFilter, error) {
	if retention <= 0 {
		return RetentionFilter{}, fmt.Errorf("retention must be positive; got %s", retention)
	}
	var ie promrelabel.IfExpression
	if err := ie.Parse(matchExpr); err != nil {
		return RetentionFilter{}, fmt.Errorf("cannot parse retention filter %q: %w", matchExpr, err)
	}
	return RetentionFilter{
		matchExpr:      matchExpr,
		retentionMsecs: retention.Milliseconds(),
		ie:             ie,
	}, nil
}

// Retention returns the retention period configured for rf.
func (rf *RetentionFilter) Retention() time.Duration {
	return time.Duration(rf.retentionMsecs) * time.Millisecond
}

func (rf *RetentionFilter) matches(labels []prompb.Label) bool {
	return rf.ie.Match(labels)
}

func retentionFiltersConfig(rfs []RetentionFilter) string {
	if len(rfs) == 0 {
		return ""
	}
	parts := make([]string, len(rfs))
	for i := range rfs {
		rf := &rfs[i]
		parts[i] = fmt.Sprintf("%s:%d", rf.matchExpr, rf.retentionMsecs)
	}
	return strings.Join(parts, "\n")
}

func metricNameToPromLabels(dst []prompb.Label, mn *MetricName) []prompb.Label {
	dst = append(dst, prompb.Label{Name: "__name__", Value: string(mn.MetricGroup)})
	for i := range mn.Tags {
		tag := &mn.Tags[i]
		dst = append(dst, prompb.Label{Name: string(tag.Key), Value: string(tag.Value)})
	}
	return dst
}
