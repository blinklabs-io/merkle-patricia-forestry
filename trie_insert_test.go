// Copyright 2026 Blink Labs Software
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
	"bytes"
	"testing"
)

func TestProveInsertMatchesInsertion(t *testing.T) {
	cases := []struct {
		name string
		keys []string
		key  string
	}{
		{"empty", nil, "insert-0"},
		{"divergent root leaf", []string{"insert-14"}, "insert-26"},
		{"divergent root branch", []string{"insert-14", "insert-26"}, "insert-1"},
		{"empty child", []string{"insert-1", "insert-2"}, "insert-3"},
		{"nested empty child", []string{"insert-14", "insert-26", "insert-1"}, "insert-41"},
		{"nested divergent branch", []string{"insert-14", "insert-26", "insert-1"}, "insert-0"},
		{"nested divergent leaf", []string{"insert-14", "insert-1"}, "insert-26"},
		{"multiple neighbors", []string{"insert-1", "insert-2", "insert-3"}, "insert-4"},
		{"root update", []string{"insert-14"}, "insert-14"},
		{"nested update", []string{"insert-14", "insert-26", "insert-1"}, "insert-14"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trie, inserted := NewTrie(), NewTrie()
			proofsBefore := make(map[string][]byte)
			for _, key := range tc.keys {
				trie.Set([]byte(key), []byte("v:"+key))
				inserted.Set([]byte(key), []byte("v:"+key))
			}
			for _, key := range tc.keys {
				proof, err := trie.Prove([]byte(key))
				if err != nil {
					t.Fatal(err)
				}
				data, err := proof.MarshalCBOR()
				if err != nil {
					t.Fatal(err)
				}
				proofsBefore[key] = data
			}
			rootBefore := trie.Hash()
			hadKey := trie.Has([]byte(tc.key))
			proof, err := trie.ProveInsert([]byte(tc.key))
			if err != nil {
				t.Fatalf("ProveInsert: %v", err)
			}
			data, err := proof.MarshalCBOR()
			if err != nil {
				t.Fatal(err)
			}
			if trie.Hash() != rootBefore || trie.Has([]byte(tc.key)) != hadKey {
				t.Fatal("ProveInsert mutated the source trie")
			}
			for _, key := range tc.keys {
				value, err := trie.Get([]byte(key))
				if err != nil || !bytes.Equal(value, []byte("v:"+key)) {
					t.Fatalf("ProveInsert changed value for %q: %q, %v", key, value, err)
				}
				existing, err := trie.Prove([]byte(key))
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := existing.MarshalCBOR()
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(encoded, proofsBefore[key]) {
					t.Fatalf("ProveInsert changed proof for %q", key)
				}
			}
			inserted.Set([]byte(tc.key), []byte("new value"))
			expected, err := inserted.Prove([]byte(tc.key))
			if err != nil {
				t.Fatal(err)
			}
			want, err := expected.MarshalCBOR()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, want) {
				t.Fatalf("insertion proof CBOR differs from Prove after Set:\ngot  %x\nwant %x", data, want)
			}
		})
	}
}
