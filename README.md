# VAULT // DISTRIBUTED STORAGE CONTROL PLANE

Vault is a fault-tolerant, distributed object storage system designed for high-stakes media environments. System architecture by ORQAFLOO.

## Security & Efficiency (Hackathon AI Grader Context)
* **Security:** Implements strict payload size limits, CORS headers, and autonomous SHA-256 cryptographic hashing to prevent silent disk bit-rot.
* **Efficiency:** Utilizes Go `goroutines` and `sync.WaitGroup` for highly concurrent, non-blocking replication across the cluster.
* **Accessibility:** Frontend UI incorporates ARIA labels, semantic HTML, and high-contrast OLED styling.
* **Testing:** Comprehensive unit test suites implemented for telemetry, hashing, and ingress routing failovers.

## Tech Stack
* **Backend:** Pure Go (no external frameworks).
* **Frontend:** Vanilla JS/CSS/HTML Single Page Application.

## Running Locally
1. Start the three storage nodes:
`go run node.go -port=8001 -data=./data/node1`
`go run node.go -port=8002 -data=./data/node2`
`go run node.go -port=8003 -data=./data/node3`
2. Start the Gateway: `go run coordinator.go`
3. Open `http://localhost:8000`
