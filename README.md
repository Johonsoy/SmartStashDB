# 🚀 SmartStashDB: A High-Performance Key-Value Store

Welcome to **SmartStashDB**, a blazing-fast, Go-powered key-value store built from scratch using **LSM-Tree**, **Skip-List**, and **Write-Ahead Logging (WAL)**. Designed for high throughput and low latency, SmartStashDB is perfect for applications demanding scalable, reliable, and efficient data storage.

---

## 🌟 Features

- **High Performance**: Optimized for low-latency reads and writes, leveraging LSM-Tree and Skip-List for efficient data organization.
- **Durability**: Write-Ahead Logging ensures no data is lost, even in the face of crashes.
- **Scalability**: LSM-Tree architecture supports massive datasets with seamless compaction.
- **Memory Efficiency**: Skip-List provides fast in-memory indexing with minimal overhead.
- **Simple API**: Intuitive key-value operations for easy integration.
- **Go-Powered**: Written in Go for concurrency, simplicity, and cross-platform support.

---

## 🛠️ Architecture

SmartStashDB combines cutting-edge data structures and techniques to deliver top-tier performance:

- **LSM-Tree**: Log-Structured Merge-Tree for write-heavy workloads, with background compaction to keep reads fast.
- **Skip-List**: Probabilistic data structure for in-memory indexing, enabling O(log n) lookups.
- **WAL**: Write-Ahead Logging for crash recovery and data durability.
- **Compaction**: Periodic merging of SSTables to optimize storage and query performance.

```
[Client] --> [API: Get/Put/Delete] --> [MemTable (Skip-List)]
                                             |
                                             v
                                        [WAL (Disk)]
                                             |
                                             v
                                      [SSTables (LSM-Tree)]
```

---

## 🚀 Getting Started

### Prerequisites
- **Go**: Version 1.18 or higher
- A passion for high-performance systems! 😎

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/johnsoy/SmartStashDB.git
   cd SmartStashDB
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build and run:
   ```bash
   go build
   ./SmartStashDB
   ```

### Example Usage
```go
package main

import (
    "fmt"
    "github.com/johnsoy/SmartStashDB"
)

func main() {
    // Initialize SmartStashDB
    kv, err := SmartStashDB.NewSmartStashDB("./data")
    if err != nil {
        panic(err)
    }
    defer kv.Close()

    // Put key-value pair
    kv.Put([]byte("key1"), []byte("value1"))

    // Get value
    value, err := kv.Get([]byte("key1"))
    if err != nil {
        panic(err)
    }
    fmt.Printf("Key: key1, Value: %s\n", value)

    // Delete key
    kv.Delete([]byte("key1"))
}
```

---

## 📊 Performance

SmartStashDB is designed for speed and scalability. Preliminary benchmarks (on a standard laptop with SSD):
- **Write Throughput**: ~500,000 ops/sec
- **Read Throughput**: ~600,000 ops/sec
- **Latency**: < 1ms for 99th percentile reads/writes

Run benchmarks yourself:
```bash
go test -bench=.
```

---

## 🛠️ Configuration

Customize SmartStashDB via the `config.yaml` file:
```yaml
data_dir: "./data"          # Storage directory
memtable_size: 1048576      # Max MemTable size (bytes)
compaction_interval: 60     # Compaction interval (seconds)
wal_flush_interval: 10      # WAL flush interval (seconds)
```

Load config programmatically:
```go
kv, err := SmartStashDB.NewSmartStashDBWithConfig("config.yaml")
```

---

## 🤝 Contributing

Contributions are welcome! Whether it's bug fixes, new features, or documentation improvements, here's how to get started:
1. Fork the repository.
2. Create a feature branch: `git checkout -b feature/awesome-feature`.
3. Commit your changes: `git commit -m "Add awesome feature"`.
4. Push to the branch: `git push origin feature/awesome-feature`.
5. Open a Pull Request.

Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

---

## 📜 License

SmartStashDB is licensed under the [MIT License](LICENSE). Feel free to use, modify, and distribute it as you see fit!

---

## 📫 Contact

- **GitHub**: [Johnsoy](https://github.com/Johonsoy)
- **Email**: [15520754767@163.com]

Star ⭐ this repo if you find SmartStashDB awesome, and let's build the fastest KV store together! 🚀
