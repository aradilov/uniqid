# uniqid

A fast and allocation-free unique 64‑bit ID generator with a deterministic structure and a compact 16‑character hexadecimal representation.

This ID format is designed for high‑performance systems with millions of requests per second, allowing efficient sharding, logging, and analytics without relying on UUID libraries.

---

## Features

- Generates unique 64‑bit IDs.
- Zero‑allocation hex encoding via `Append`.
- Extracts `serverID` from a hex ID using `GetServerID`.
- Atomic counter with no locks.
- Lexicographically sortable by time.
- Compact: 8 bytes → 16 hex characters.
- Extremely fast (nanosecond‑level operations).

---

## ID Structure

The 64‑bit value is encoded as:

| Component | Size     | Description                          |
|----------|----------|--------------------------------------|
| serverID | 16 bits  | Unique ID of the machine/instance    |
| timestamp| 32 bits  | Unix timestamp (seconds)             |
| counter  | 16 bits  | Atomic incrementing counter          |

All fields are stored in **big‑endian** order.

Hex output is always exactly **16 bytes**.

```
[ 2 bytes serverID ][ 4 bytes timestamp ][ 2 bytes counter ]
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/aradilov/uniqid"
)

func main() {
    // generate a uint64 ID
    id := uniqid.Get()
    fmt.Println("uint64 ID:", id)

    // encode into a 16-byte hex buffer
    buf := make([]byte, 16)
    uniqid.Append(id, buf)
    fmt.Println("hex:", string(buf))

    // extract serverID directly from the hex ID
    srv := uniqid.GetServerID(buf)
    fmt.Println("serverID:", srv)
}
```

---

## Extracting ServerID from Hex
The function `GetServerID` parses the first 4 hex characters and returns the original serverID.

```go
func GetServerID(hex []byte) uint16 {
    if len(hex) < 4 {
        return 0
    }

    b0h, b0l := fromHex(hex[0]), fromHex(hex[1])
    b1h, b1l := fromHex(hex[2]), fromHex(hex[3])
    if b0h == 0xff || b0l == 0xff || b1h == 0xff || b1l == 0xff {
        return 0 // invalid hex input
    }

    b0 := (b0h << 4) | b0l
    b1 := (b1h << 4) | b1l

    return uint16(b0)<<8 | uint16(b1)
}
```

---

## Benchmarks

Measured on Intel i5‑1038NG7, Go 1.23:

```
BenchmarkGet-8         ~3.5 ns/op
BenchmarkAppend-8      ~12–15 ns/op
BenchmarkGetServerID   ~8 ns/op
Allocations:           0 B/op
```

---

## Why Not UUID?

| Feature             | UUID         | uniqid |
|--------------------|--------------|--------|
| Size               | 36 chars     | 16 chars |
| Speed              | slower       | extremely fast |
| Contains timestamp | sometimes    | always |
| Contains serverID  | no           | yes |
| Sortable           | no           | yes |
| Allocations        | many         | zero |

`uniqid` is ideal for ultra‑high load environments such as real‑time bidding, logging systems, distributed workers, and microservices.

---

## Installation

```
go get github.com/aradilov/uniqid
```

---

## License

MIT
