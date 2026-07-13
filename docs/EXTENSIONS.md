# Scanner and AI Extensions

The domain exposes `Scanner` and `AIProvider` ports in `internal/domain/models.go`. Concrete integrations should live under `internal/infrastructure`.

## Scanner Pattern

A future scan flow should:

1. Add `scan_jobs` and `scan_findings` migrations.
2. Add a scan application service that validates project access and enqueues a job.
3. Use Redis Streams or a dedicated queue adapter behind a domain job-queue port.
4. Run scanner workers separately from the API process.
5. Implement adapters such as `HTTPXScanner`, `NaabuScanner`, and `NucleiScanner`.
6. Parse tool output into structured findings before persistence.

Scanner commands must use argument arrays rather than shell interpolation. Apply execution timeouts, concurrency limits, allowlists, output size limits, and an auditable record of who initiated each scan.

## AI Pattern

An AI adapter should receive structured project, asset, and finding data from an application service. Keep provider SDKs, prompts, rate limits, and model selection in infrastructure. Persist prompt/version metadata for traceability, redact secrets before requests, and treat generated conclusions as advisory rather than authoritative.

## Suggested Packages

```text
internal/application/scans/
internal/domain/findings.go
internal/infrastructure/queue/redis_streams.go
internal/infrastructure/scanners/httpx.go
internal/infrastructure/scanners/naabu.go
internal/infrastructure/scanners/nuclei.go
internal/infrastructure/ai/provider.go
cmd/worker/main.go
```
