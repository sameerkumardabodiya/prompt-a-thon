# VAULT // DISTRIBUTED STORAGE CONTROL PLANE

Vault is a fault-tolerant, distributed object storage system designed for high-stakes media environments. System architecture by ORQAFLOO.

## The Problem
Storing heavy, high-value creative files (like multi-terabyte, color-graded cinematic travel vlogs) on single points of failure risks catastrophic data loss. Hardware fails, and hard drives suffer from silent bit-rot. 

## The Solution
Vault acts as a distributed cluster and smart gateway. 
* **High Availability:** Objects are replicated across multiple independent storage nodes.
* **Fault Tolerance:** If a node suffers catastrophic failure, the coordinator dynamically routes requests to surviving nodes with zero downtime.
* **Cryptographic Integrity:** Nodes autonomously generate SHA-256 signatures upon ingestion. If a disk silently corrupts a file, the node detects the mismatch, blocks the transfer, and the coordinator bypasses it to retrieve a pristine copy.

## Tech Stack
* **Backend:** Go (Golang) for high-concurrency routing, SHA-256 hashing, and cluster telemetry.
* **Frontend:** Vanilla JS/CSS/HTML Single Page Application directly mapping to the Go API.

## Running Locally
1. Start the three storage nodes in separate terminals:
`go run node.go -port=8001 -data=./data/node1`
`go run node.go -port=8002 -data=./data/node2`
`go run node.go -port=8003 -data=./data/node3`
2. Start the Gateway/UI in a fourth terminal:
`go run coordinator.go`
3. Open `http://localhost:8000` in your browser.
