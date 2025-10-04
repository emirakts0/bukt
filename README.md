<div align="center">
  <img src="./readme/banner.png" alt="Bukt Logo" width="1000">
  <h1 align="center">Bukt</h1>
  <p align="center">
    A high-performance in-memory key-value store.
  </p>
</div>

---

Bukt is a modern key-value store written in Go, designed for speed and simplicity. It provides extremely fast read operations by serving data directly from memory. The current API is primarily designed for handling string data.

## Project Status

> **Warning**
> Bukt is currently under active development and should not be used in production environments. The API is subject to change, and the feature set is not yet complete.


## Design Philosophy

-   **Read-Optimized Performance:** Bukt is architected for read-heavy workloads. The data path for retrieving data from memory is highly concurrent and designed to be virtually lock-free, enabling extremely high throughput for read operations.

## Benchmarks

Detailed benchmarks for various workloads (read-heavy, write-heavy, mixed) are underway and will be published soon.

## Features

- **Bucket Mechanism:** Organize your data into isolated namespaces called buckets.
- **Sharding:** Keys are automatically distributed across multiple shards to reduce lock contention and improve concurrency on multi-core systems.
- **Data Compression:** Optional on-the-fly data compression to reduce the storage footprint for large values.
- **Time-to-Live (TTL):** Set an automatic expiration time for your keys. Bukt efficiently manages and removes expired data in the background.
- **Single-Read Keys:** Create keys that are automatically deleted after being read once, ideal for temporary or single-use data patterns.
- **Multiple Transport Layers:**
  - **HTTP/REST:** A simple and convenient API for standard web-based interactions.
  - **TCP & gRPC:** (Planned) For high-performance, low-latency communication between services.
- **Container Ready:** Comes with a setup for quick, isolated deployments.

### Core Architecture

<div align="center">
  <img src="./readme/schema.png" alt="Bukt Schema" width="600">
</div>

## Getting Started

*(Instructions for building and running the project will be added as the project matures.)*

---

<p align="center">
  <em>Inspired by the concept of "buckets" for data organization.</em>
</p>
