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

import "testing"

func TestLeafExpectedHash(t *testing.T) {
	testDefs := []struct {
		prefix       []Nibble
		key          []byte
		value        []byte
		expectedHash string
	}{
		// Prefix with even length
		{
			prefix: bytesToNibbles(
				HashValue([]byte{0xab}).Bytes(),
			),
			key:          []byte{0xab},
			value:        []byte{0x0, 0x1, 0xe, 0xf},
			expectedHash: "201e6c905db9d8ba1d107e3fbd1e9af545d7b0505b297f73b6f92fd5e4d9c235",
		},
		// Prefix with odd length
		{
			prefix: bytesToNibbles(
				HashValue([]byte{0xab}).Bytes(),
			)[1:],
			key:          []byte{0xab},
			value:        []byte{0x0, 0x1, 0xe, 0xf},
			expectedHash: "87899327d3cef386073418f94e188ce6cbd410fa9312d7ca790a1dbc34368c36",
		},
		// Hashed key as prefix
		// This hash is validated against the original JS implementation
		{
			prefix: bytesToNibbles(
				HashValue([]byte{0xab, 0xcd}).Bytes(),
			),
			key:          []byte{0xab, 0xcd},
			value:        []byte{0x12, 0x34},
			expectedHash: "1887f50447e27c729c781598745de46ed35c8f5a68cec25b68e6178a2cfc8e96",
		},
	}
	for _, testDef := range testDefs {
		l := newLeaf(
			testDef.prefix,
			testDef.key,
			testDef.value,
		)
		leafHash := l.Hash().String()
		if leafHash != testDef.expectedHash {
			t.Errorf("did not got expected hash: got %s, expected %s", leafHash, testDef.expectedHash)
		}
	}
}
