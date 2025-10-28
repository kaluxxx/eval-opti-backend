| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `curl -s http://localhost:8080/api/v1/stats?days=100 -o nul` | 31.4 ± 2.6 | 26.9 | 35.2 | 1.35 ± 0.12 |
| `curl -s http://localhost:8080/api/v2/stats?days=100 -o nul` | 23.3 ± 0.7 | 22.2 | 24.5 | 1.00 |
