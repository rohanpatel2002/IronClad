# SOAR Automation & Distributed Synchronization Model

This document specifies the Security Orchestration, Automation, and Response (SOAR) architecture in IRONCLAD.

## 1. SOAR Engine
- **Autonomous Service Quarantine**: When high-risk automated attacks or anomalous deployment patterns are detected, `QuarantineManager` communicates with OPA to immediately block service access.
- **Anomaly Detection Engine**: Employs sliding-window running mean and standard deviation (Welford's algorithm) to compute real-time Z-scores on decision streams.

## 2. Distributed Lock Synchronization
- `DistLock` implements a distributed mutex over Redis with auto-expiring TTLs and safe UUID ownership verification script execution.
