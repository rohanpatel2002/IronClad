# Governance & Rate Limiting Architecture

The `gate-go` service enforces real-time API rate limits and generates audit compliance reports for security governance.

## 1. Rate Limiting Strategies
- **Token Bucket Rate Limiter**: Fixed request tokens per client IP with burst capacity.
- **Leaky Bucket Limiter**: Smooths out traffic spikes by leaking requests at a constant interval.
- **Dynamic Resource-Aware Limiter**: Monitors system memory consumption (Alloc MB) and throttles incoming traffic under high resource pressure.

## 2. Compliance Reporting
The governance module produces regulatory compliance summaries:
- **SOC 2 Type II Compliance Reports** (`GenerateSOC2Summary` & `GenerateSOC2CSV`)
- **ISO 27001 Annex A.12 Change Controls** (`GenerateISO27001Summary`)
