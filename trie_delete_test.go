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

import "testing"

func TestTrieDeleteCollapsesNestedBranch(t *testing.T) {
	type entry struct {
		key   string
		value string
	}
	testCases := []struct {
		name      string
		entries   []entry
		deleteKey string
	}{
		{
			name: "shared byte prefix",
			entries: []entry{
				{key: "81", value: "1"},
				{key: "189", value: "2"},
				{key: "206", value: "2"},
			},
			deleteKey: "206",
		},
		{
			name: "remaining child is branch",
			entries: []entry{
				{key: "81", value: "1"},
				{key: "189", value: "2"},
				{key: "1", value: "3"},
			},
			deleteKey: "1",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			trie := NewTrie()
			for _, entry := range testCase.entries {
				trie.Set([]byte(entry.key), []byte(entry.value))
			}

			if err := trie.Delete([]byte(testCase.deleteKey)); err != nil {
				t.Fatalf("Delete(%s): %v", testCase.deleteKey, err)
			}
			if trie.Has([]byte(testCase.deleteKey)) {
				t.Fatal("deleted key remains in trie")
			}

			want := NewTrie()
			for _, entry := range testCase.entries {
				if entry.key == testCase.deleteKey {
					continue
				}
				want.Set([]byte(entry.key), []byte(entry.value))
				value, err := trie.Get([]byte(entry.key))
				if err != nil {
					t.Fatalf("Get(%s): %v", entry.key, err)
				}
				if string(value) != entry.value {
					t.Fatalf(
						"Get(%s) = %q, want %q",
						entry.key,
						value,
						entry.value,
					)
				}
				if _, err := trie.Prove([]byte(entry.key)); err != nil {
					t.Fatalf("Prove(%s): %v", entry.key, err)
				}
			}
			if trie.Hash() != want.Hash() {
				t.Fatalf("root hash = %s, want %s", trie.Hash(), want.Hash())
			}
		})
	}
}
