package storage

import (
	"testing"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/prompb"
)

func TestRetentionFilterMatch(t *testing.T) {
	rf, err := NewRetentionFilter(`cpu_usage{env=~"dev|staging",host!="skip"}`, 48*time.Hour)
	if err != nil {
		t.Fatalf("cannot create retention filter: %s", err)
	}
	f := func(labels []prompb.Label, want bool) {
		t.Helper()
		if got := rf.matches(labels); got != want {
			t.Fatalf("unexpected match result for %v: got %v; want %v", labels, got, want)
		}
	}
	f([]prompb.Label{{Name: "__name__", Value: "cpu_usage"}, {Name: "env", Value: "dev"}, {Name: "host", Value: "a"}}, true)
	f([]prompb.Label{{Name: "__name__", Value: "cpu_usage"}, {Name: "env", Value: "prod"}, {Name: "host", Value: "a"}}, false)
	f([]prompb.Label{{Name: "__name__", Value: "cpu_usage"}, {Name: "env", Value: "staging"}, {Name: "host", Value: "skip"}}, false)
}

func TestStorageRetentionFilter(t *testing.T) {
	rfBroad, err := NewRetentionFilter(`{tier=~"short|long"}`, 10*24*time.Hour)
	if err != nil {
		t.Fatalf("cannot create broad retention filter: %s", err)
	}
	rfShort, err := NewRetentionFilter(`{tier="short"}`, 48*time.Hour)
	if err != nil {
		t.Fatalf("cannot create short retention filter: %s", err)
	}
	s := MustOpenStorage(t.TempDir(), OpenOptions{
		Retention:        30 * 24 * time.Hour,
		RetentionFilters: []RetentionFilter{rfBroad, rfShort},
	})
	defer s.MustClose()

	now := time.Now()
	metricNameRaw := func(tier string) []byte {
		var mn MetricName
		mn.MetricGroup = []byte("temperature_celsius")
		mn.Tags = []Tag{{Key: []byte("tier"), Value: []byte(tier)}}
		return mn.marshalRaw(nil)
	}
	rows := []MetricRow{
		{MetricNameRaw: metricNameRaw("short"), Timestamp: now.Add(-5 * 24 * time.Hour).UnixMilli(), Value: 10},
		{MetricNameRaw: metricNameRaw("short"), Timestamp: now.Add(-time.Hour).UnixMilli(), Value: 11},
		{MetricNameRaw: metricNameRaw("long"), Timestamp: now.Add(-5 * 24 * time.Hour).UnixMilli(), Value: 20},
		{MetricNameRaw: metricNameRaw("long"), Timestamp: now.Add(-time.Hour).UnixMilli(), Value: 21},
	}
	s.AddRows(rows, defaultPrecisionBits)
	s.DebugFlush()
	if err := s.ForceMergePartitions(""); err != nil {
		t.Fatalf("cannot force merge partitions: %s", err)
	}

	var m Metrics
	s.UpdateMetrics(&m)
	if got := m.TableMetrics.TotalRowsCount(); got != 3 {
		t.Fatalf("unexpected row count after retention-filter merge: got %d; want 3", got)
	}
}
