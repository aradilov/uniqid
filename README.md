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

| Component | Size    | Description                          |
|-----------|---------|--------------------------------------|
| serverID  | 16 bits | Unique ID of the machine/instance    |
| uniqid    | 48 bits | Atomic incrementing counter           |

All fields are stored in **big‑endian** order.

Hex output is always exactly **16 bytes**.

```
[ 16 bits serverID ][ 48 bits sequence ]
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


## How it works

The core of the ID generator is a 64‑bit value composed of:

```
[ 16 bits serverID ][ 48 bits sequence ]
```

Below is a step‑by‑step explanation of how the `Get()` method generates a unique ID:

### 1. Initialization of the global sequence counter

```go
var uniqueAdID = func() uint64 {
    return uint64(time.Now().UnixNano())
}()
```

This global variable is initialized **once at program startup** with the current Unix time in nanoseconds.  
It becomes the starting point of a monotonically increasing sequence.

### 2. Ensuring `serverID` is initialized

```go
once.Do(initServerID)
```

A `sync.Once` guarantees that `initServerID()` runs **exactly once**, even under heavy concurrency.  
The server ID is derived from the machine's external IPv4 address or explicitly set via `SetServerID`.

### 3. Atomic increment of the sequence

```go
adID := atomic.AddUint64(&uniqueAdID, 1)
```

The 64‑bit global counter is atomically incremented, producing a strictly increasing sequence without locks.  
This ensures uniqueness across all goroutines on the same server.

### 4. Extracting the lower 48 bits

```go
const mask48 uint64 = (uint64(1) << 48) - 1
```

This mask keeps only the **lowest 48 bits** of the incremented sequence:

```
0x0000FFFFFFFFFFFF
```

### 5. Building the final 64‑bit ID

```go
return (uint64(serverID) << 48) | (adID & mask48)
```

- `serverID << 48` places the server ID into the **upper 16 bits**.
- `(adID & mask48)` fills the lower 48 bits with the incrementing sequence.
- `|` combines the two values.

Final structure:

```
[ 16 bits serverID ][ 48 bits incrementing sequence ]
```

### 6. Practical uniqueness guarantees

Because the sequence counter occupies 48 bits:

- Total unique values per server: `2^48 ≈ 281 trillion`.
- At **1,000,000 IDs per second**, wraparound occurs only after:

```
2^48 / 1e6 ≈ 281 million seconds ≈ 8.9 years
```

This means a **single server** can safely generate **1 million unique IDs per second for nearly 9 years** before any theoretical risk of collision.

---

## License

MIT
