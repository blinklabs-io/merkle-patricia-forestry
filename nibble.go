// Copyright 2025 Blink Labs Software
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mpf

import (
	"encoding/hex"
	"strings"
)

type Nibble byte

func (n Nibble) String() string {
	return hex.EncodeToString([]byte{byte(n)})[1:]
}

// bytesToNibbles splits a series of bytes into a series of nibbles
func bytesToNibbles(data []byte) []Nibble {
	ret := []Nibble{}
	for _, dataByte := range data {
		tmpNibbles := byteToNibbles(dataByte)
		ret = append(ret, tmpNibbles...)
	}
	return ret
}

// byteToNibbles splits a byte into two bytes representing the upper and lower 4 bits of the original byte.
// The value 0xab would be returned as [0x0a, 0x0b]
func byteToNibbles(data byte) []Nibble {
	// Split byte into two bytes representing the upper and lower 4 bits
	return []Nibble{
		Nibble(data >> 4),
		Nibble(data & 0xf),
	}
}

// nibblesToBytes converts a series of Nibbles into a byte slice representing the original bytes.
// This function assumes the input length is even
func nibblesToBytes(data []Nibble) []byte {
	ret := make([]byte, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		tmpByte := byte(data[i]<<4) + byte(data[i+1])
		ret = append(ret, tmpByte)
	}
	return ret
}

// nibblesToHexString converts a series of Nibbles into a hex string representing those nibbles.
func nibblesToHexString(data []Nibble) string {
	var sb strings.Builder
	for _, nibble := range data {
		sb.WriteString(nibble.String())
	}
	return sb.String()
}

// keyToPath converts an arbitrary key to the sequence of Nibbles representing the path to the value
func keyToPath(key []byte) []Nibble {
	keyHash := HashValue(key)
	keyHashNibbles := bytesToNibbles(keyHash.Bytes())
	return keyHashNibbles
}
