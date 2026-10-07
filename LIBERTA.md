# Liberta

Liberta is a minimal upstream-tracking fork of VictoriaMetrics Community Edition. Its only intended functional divergence is per-series retention filters.

## Added feature

`-retentionFilter='selector:duration'` may be specified multiple times on `victoria-metrics` / `vmstorage` through the shared `vmstorage` flags.

Example:

```text
-retentionPeriod=400d \
-retentionFilter='{resolution="raw"}:8d' \
-retentionFilter='{resolution="1s"}:65d'
```

Semantics:

- unmatched series use `-retentionPeriod`;
- when multiple filters match, the shortest retention wins;
- filter retention must be at least 24 hours and no longer than the global retention;
- expired samples are removed during normal background merges;
- partially expired blocks are trimmed rather than retained until the whole block expires;
- historical partitions are force-merged once when the retention-filter configuration changes;
- IndexDB follows the global retention, matching VictoriaMetrics' documented retention-filter behavior.

The implementation uses the retention hooks already present in the Apache-2.0 Community source tree and is independently implemented from public source and documentation. It does not incorporate VictoriaMetrics Enterprise source code.

## Upstream tracking

`.github/workflows/sync-upstream.yml` periodically rebases Liberta onto `VictoriaMetrics/VictoriaMetrics:master`, runs the retention/storage tests, builds `app/victoria-metrics`, and updates `main` only if all checks pass. Conflicts stop the update rather than silently dropping the Liberta patch.
