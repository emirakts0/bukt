<div align="center">
  <img src="./readme/banner.png" alt="Bukt Logo" width="1000">
  <h1 align="center">Bukt</h1>
  <p align="center">
    Bukt is an in-memory key-value store written in Go, built for speed and simplicity, with data organized into lightweight, secure buckets.
  </p>
</div>

---

## Project Status

> Bukt is a **conceptual and learning-oriented project** built to explore the internals of high-performance in-memory key-value stores.  
> It is **not production-ready** and may contain incomplete features or experimental implementations.


## Design Philosophy

-   **Read-Optimized Performance:** Bukt is architected for read-heavy workloads. The data path for retrieving data from memory is highly concurrent and designed to be virtually lock-free, enabling extremely high throughput for read operations.

## Features

- **Bucket Mechanism:** Organize your data into isolated namespaces called buckets. Each bucket can be protected with a unique authentication token, ensuring secure data isolation.
- **Time-to-Live (TTL):** Set an automatic expiration time for your keys. Bukt efficiently manages and removes expired data in the background.
- **Single-Read Keys:** Create keys that are automatically deleted after being read once, ideal for temporary or single-use data patterns.
- **Multiple Transport Layers:**
  - **HTTP/REST:** A simple and convenient API for standard web-based interactions.
  - **TCP (Binary Protocol):** Ultra-fast TCP server with binary protocol for high-performance, low-latency communication (using gnet).
  - **gRPC:** (Planned) For structured service-to-service communication.
- **Container Ready:** Comes with a setup for quick, isolated deployments.

---

## API at a Glance

### TCP Binary Protocol

For performance-critical applications, the TCP protocol offers lower latency and higher throughput. It uses a custom binary frame format for communication.

#### Frame Structure

`[Length(4)][Command(1)][RequestID(8)][Payload(variable)]`

-   **Length**: Total frame size (header + payload) as a 4-byte unsigned integer.
-   **Command**: A single byte representing the operation (e.g., `SET`, `GET`, `DELETE`).
-   **RequestID**: An 8-byte unique identifier for the request, used to correlate responses.
-   **Payload**: Variable-length data specific to the command.

#### Commands

-   **`SET (0x01)`**: Stores a key-value pair.
-   **`GET (0x02)`**: Retrieves a key.
-   **`DELETE (0x03)`**: Deletes a key.

### HTTP/REST API

The HTTP API provides a simple, stateless interface for managing buckets and key-value pairs. All endpoints are prefixed with `/api/v1`.

#### Buckets

-   **`POST /buckets`**: Creates a new bucket.
-   **`GET /buckets`**: Lists all available buckets.
-   **`GET /buckets/{name}`**: Retrieves details for a specific bucket.
-   **`DELETE /buckets/{name}`**: Deletes a bucket. Requires the bucket's auth token in the body.

#### Key-Value Operations

All key-value operations require an `X-Auth-Token` header containing the authentication token for the bucket.

-   **`POST /kv`**: Stores a new key-value pair.
-   **`GET /kv/{key}`**: Retrieves the value for a given key.
-   **`DELETE /kv/{key}`**: Deletes a key-value pair.

---

## Benchmarks

Benchmarks were run on a single CPU thread with a 95% read ratio.

| Metric | Sequential Mode | Parallel Mode |
| :--- | :--- | :--- |
| **Throughput** | `9,584 ops/sec` | `31,972 ops/sec` |
| **Data Transfer** | `9.36 MB/s` | `31.22 MB/s` |
| **Average Latency** | `103µs` | `1.5ms` |
| ↳ Reads | `90,988 (94.9%)` | `304,034 (95.1%)`|
| ↳ Writes | `4,853 (5.1%)` | `15,725 (4.9%)` |

---

<p align="center">
  <em>Inspired by the concept of "buckets" for data organization.</em>
</p>
