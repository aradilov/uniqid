package uniqid

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var (
	serverID uint16
	once     sync.Once
)

// SetServerID sets the serverID to the provided value if it has not already been set; panics if serverID is already set.
func SetServerID(id uint16) {
	if serverID > 0 {
		log.Panicf("serverID already set")
	}
	serverID = id
}

// Get generates a globally unique 64-bit identifier combining a server-specific ID and an atomic counter.
func Get() uint64 {
	once.Do(initServerID)
	adID := atomic.AddUint64(&uniqueAdID, 1)
	const mask48 uint64 = (uint64(1) << 48) - 1
	return (uint64(serverID) << 48) | (adID & mask48)
}

// Append appends unique id hex to dst.
func Append(dst []byte) []byte {
	n := Get()

	for i := uint(1); i <= 8; i++ {
		shift := 64 - (i << 3)
		c := byte(n >> shift)
		dst = append(dst, hexByte(c>>4), hexByte(c&0xf))
	}
	return dst
}

func GetServerID(hex []byte) uint16 {
	if len(hex) < 16 {
		return 0
	}

	b0h, b0l := fromHex(hex[0]), fromHex(hex[1])
	b1h, b1l := fromHex(hex[2]), fromHex(hex[3])
	if b0h == 0xff || b0l == 0xff || b1h == 0xff || b1l == 0xff {
		// некоректний hex
		return 0
	}

	b0 := (b0h << 4) | b0l
	b1 := (b1h << 4) | b1l

	// serverID стоїть у старших 16 бітах: [b0 b1]....
	return uint16(b0)<<8 | uint16(b1)
}

func fromHex(b byte) byte {
	switch {
	case '0' <= b && b <= '9':
		return b - '0'
	case 'a' <= b && b <= 'f':
		return b - 'a' + 10
	case 'A' <= b && b <= 'F':
		return b - 'A' + 10
	default:
		return 0xff
	}
}

func hexByte(c byte) byte {
	if c < 10 {
		return '0' + c
	}
	return c - 10 + 'A'
}

// initServerID initializes the serverID using the external IP, setting it based on the last two bytes of the IP address.
func initServerID() {
	if serverID > 0 {
		return
	}
	ip4 := ExternalIP().To4()
	if ip4 == nil {
		log.Panicf("cannot get external ip")
	}

	serverID = uint16(ip4[2])<<8 | uint16(ip4[3])
}

var uniqueAdID = func() uint64 {
	return uint64(time.Now().UnixNano())
}()
