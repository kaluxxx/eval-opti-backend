| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `curl -s http://localhost:8080/api/v1/export/csv?days=30 -o nul` | 2.072 ± 0.005 | 2.068 | 2.082 | 51.48 ± 2.11 |
| `curl -s http://localhost:8080/api/v2/export/csv?days=30 -o nul` | 0.040 ± 0.002 | 0.038 | 0.042 | 1.00 |
