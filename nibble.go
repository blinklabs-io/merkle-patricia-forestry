// Copyright 2024 Blink Labs Software
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

type Nibble byte

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
