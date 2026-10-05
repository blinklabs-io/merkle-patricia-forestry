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
	"testing"

	"github.com/blinklabs-io/gouroboros/cbor"
)

// TestProofForkPrefixOddCbor builds tries that produce Fork proof steps whose
// neighbour prefix has an odd (and non-trivial even) number of nibbles, and
// asserts the encoded proof CBOR matches the canonical encoding produced by the
// reference @aiken-lang/merkle-patricia-forestry TypeScript library (the same
// encoding the on-chain Aiken validator consumes). Fork prefixes occupy one byte per nibble.
func TestProofForkPrefixOddCbor(t *testing.T) {
	cases := []struct {
		name     string
		keys     []string
		key      string
		wantRoot string
		wantCbor string
	}{
		{
			// Fork neighbour prefix = single nibble [0x0a] (odd length 1).
			name:     "odd-1-nibble",
			keys:     []string{"k-47-0", "k-47-1", "k-47-2"},
			key:      "k-47-0",
			wantRoot: "3f71ecd4f4fe3ef5af39fb8a81ecffb84e358183179d3d2f8b80fd36a8fd363f",
			wantCbor: "9fd87a9f00d8799f08410a582008553949a44e0ff9587441047d948a73d122deb3069effdfc727cbc082280be3ffffff",
		},
		{
			// Fork neighbour prefix = [0x02,0x0b] (even length 2). The old
			// packing produced [0x2b]; the correct encoding is [0x02,0x0b].
			name:     "even-2-nibbles",
			keys:     []string{"x-1610-0", "x-1610-1", "x-1610-2"},
			key:      "x-1610-0",
			wantRoot: "e6c7d5ecfa57a3e8711149e2dccedf3cab5495dd23bebe10c2a64b2e03e95d1c",
			wantCbor: "9fd87a9f00d8799f0d42020b58202f63cbe03d0dc4760156a50dece7250a500bca06bc2ac5f6cdcaf902dc3bedb4ffffff",
		},
		{
			// Fork neighbour prefix = [0x07,0x0a,0x0d] (odd length 3).
			name:     "odd-3-nibbles",
			keys:     []string{"y-6196-0", "y-6196-1", "y-6196-2"},
			key:      "y-6196-1",
			wantRoot: "460282cc474ca62a3e1d418d6757b4f0db71230834e5ff5813b5b8fc3d247f61",
			wantCbor: "9fd87a9f00d8799f0343070a0d5820bec6e8fb6d1fd5a414ab227db2e22cbdd68ac765775edf9005bd90e70776aa75ffffff",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			trie := NewTrie()
			for _, k := range c.keys {
				trie.Set([]byte(k), []byte("v:"+k))
			}
			if root := hex.EncodeToString(trie.Hash().Bytes()); root != c.wantRoot {
				t.Fatalf("trie root mismatch\n got=%s\nwant=%s", root, c.wantRoot)
			}
			proof, err := trie.Prove([]byte(c.key))
			if err != nil {
				t.Fatalf("Prove(%q): %v", c.key, err)
			}
			b, err := cbor.Encode(proof)
			if err != nil {
				t.Fatalf("encode proof: %v", err)
			}
			if got := hex.EncodeToString(b); got != c.wantCbor {
				t.Fatalf("proof CBOR mismatch\n got=%s\nwant=%s", got, c.wantCbor)
			}
		})
	}
}
