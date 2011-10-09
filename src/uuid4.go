/* Code shamelessly copied/pasted from http://code.google.com/p/tideland-cgl/source/browse/cgl.go */
package main

import (
	"io"
	"crypto/rand"
	"log"
	"encoding/hex"
)

// UUID represent a universal identifier with 16 bytes.
type UUID []byte

// NewUUID generates a new UUID based on version 4.
func NewUUID() UUID {
        uuid := make([]byte, 16)

        _, err := io.ReadFull(rand.Reader, uuid)

        if err != nil {
                log.Fatal(err)
        }

        // Set version (4) and variant (2).

        var version byte = 4 << 4
        var variant byte = 2 << 4

        uuid[6] = version | (uuid[6] & 15)
        uuid[8] = variant | (uuid[8] & 15)

        return uuid
}

// Raw returns a copy of the UUID bytes.
func (uuid UUID) Raw() []byte {
        raw := make([]byte, 16)

        copy(raw, uuid[0:16])

        return raw
}

// String returns a hexadecimal string representation with
// standardized separators.
func (uuid UUID) String() string {
        base := hex.EncodeToString(uuid.Raw())

        return base[0:8] + "-" + base[8:12] + "-" + base[12:16] + "-" + base[16:20] + "-" + base[20:32]
}
