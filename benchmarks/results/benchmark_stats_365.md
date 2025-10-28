| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `curl -s http://localhost:8080/api/v1/stats?days=365 -o nul` | 49.1 ± 3.2 | 43.2 | 52.3 | 2.19 ± 0.17 |
| `curl -s http://localhost:8080/api/v2/stats?days=365 -o nul` | 22.4 ± 0.9 | 21.2 | 23.9 | 1.00 |
