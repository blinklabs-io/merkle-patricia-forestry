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

import (
	"testing"
)

// NOTE: this hash and entry list comes from the aiken-lang/merkel-patricia-forestry tests
var fruitsExpectedHash string = "4acd78f345a686361df77541b2e0b533f53362e36620a1fdd3a13e0b61a3b078"
var fruitsTestEntries = []struct {
	key   string
	value string
}{
	{key: `apple[uid: 58]`, value: `🍎`},
	{key: `apricot[uid: 0]`, value: `🤷`},
	{key: `banana[uid: 218]`, value: `🍌`},
	{key: `blueberry[uid: 0]`, value: `🫐`},
	{key: `cherry[uid: 0]`, value: `🍒`},
	{key: `coconut[uid: 0]`, value: `🥥`},
	{key: `cranberry[uid: 0]`, value: `🤷`},
	{key: `fig[uid: 68267]`, value: `🤷`},
	{key: `grapefruit[uid: 0]`, value: `🤷`},
	{key: `grapes[uid: 0]`, value: `🍇`},
	{key: `guava[uid: 344]`, value: `🤷`},
	{key: `kiwi[uid: 0]`, value: `🥝`},
	{key: `kumquat[uid: 0]`, value: `🤷`},
	{key: `lemon[uid: 0]`, value: `🍋`},
	{key: `lime[uid: 0]`, value: `🤷`},
	{key: `mango[uid: 0]`, value: `🥭`},
	{key: `orange[uid: 0]`, value: `🍊`},
	{key: `papaya[uid: 0]`, value: `🤷`},
	{key: `passionfruit[uid: 0]`, value: `🤷`},
	{key: `peach[uid: 0]`, value: `🍑`},
	{key: `pear[uid: 0]`, value: `🍐`},
	{key: `pineapple[uid: 12577]`, value: `🍍`},
	{key: `plum[uid: 15492]`, value: `🤷`},
	{key: `pomegranate[uid: 0]`, value: `🤷`},
	{key: `raspberry[uid: 0]`, value: `🤷`},
	{key: `strawberry[uid: 2532]`, value: `🍓`},
	{key: `tangerine[uid: 11]`, value: `🍊`},
	{key: `tomato[uid: 83468]`, value: `🍅`},
	{key: `watermelon[uid: 0]`, value: `🍉`},
	{key: `yuzu[uid: 0]`, value: `🤷`},
}

func TestTrieEmpty(t *testing.T) {
	trie := NewTrie()
	if trie.Hash() != NullHash {
		t.Errorf(
			"empty trie does not have expected hash: got %s, expected null hash",
			trie.Hash().String(),
		)
	}
	if trie.size != 0 {
		t.Errorf(
			"empty trie does not have expected size: got %d, expected 0",
			trie.size,
		)
	}
}

func TestTrieHashChanges(t *testing.T) {
	trie := NewTrie()
	hash0 := trie.Hash().String()
	trie.Set([]byte("abcd"), []byte("1"))
	hash1 := trie.Hash().String()
	if hash0 == hash1 {
		t.Errorf(
			"hash did not change after insert: old %s, new %s",
			hash0,
			hash1,
		)
		return
	}
	trie.Set([]byte("bcde"), []byte("2"))
	hash2 := trie.Hash().String()
	if hash1 == hash2 {
		t.Errorf(
			"hash did not change after insert: old %s, new %s",
			hash1,
			hash2,
		)
	}
}

func TestTrieDelete(t *testing.T) {
	trie := NewTrie()
	trie.Set([]byte("abcd"), []byte("1"))
	hash1 := trie.Hash().String()
	trie.Set([]byte("bcde"), []byte("2"))
	hash2 := trie.Hash().String()
	if err := trie.Delete([]byte("bcde")); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	hash3 := trie.Hash().String()
	if hash2 == hash3 {
		t.Errorf("hash did not change after delete")
		return
	}
	if hash1 != hash3 {
		t.Errorf(
			"hash is different before and after set/delete of key: got %s, expected %s",
			hash3,
			hash1,
		)
	}
}

func TestTrieGet(t *testing.T) {
	testKey := []byte{0xaa, 0xff}
	testVal := []byte("1")
	trie := NewTrie()
	trie.Set(testKey, testVal)
	tmpVal, err := trie.Get(testKey)
	if err != nil {
		t.Errorf("unexpected error getting key: %s", err)
		return
	}
	if string(tmpVal) != string(testVal) {
		t.Errorf(
			"did not get expected value for key: got %x, expected %x",
			tmpVal,
			testVal,
		)
	}
}

func TestTrieHas(t *testing.T) {
	testKey := []byte{0xaa, 0xff}
	trie := NewTrie()
	if trie.Has(testKey) {
		t.Errorf("has key when shouldn't")
		return
	}
	trie.Set(testKey, []byte("1"))
	if !trie.Has(testKey) {
		t.Errorf("does not have key when should")
	}
}

func TestTrieSetTwice(t *testing.T) {
	expectedRootHash := "eb258590dda64098b24091629f9dbcaf7e6e55011f9a411deb9e5a9793f0d83f"
	trie := NewTrie()
	trie.Set([]byte{0xab, 0xcd}, []byte{0x01, 0x23})
	hash1 := trie.Hash()
	if hash1.String() != expectedRootHash {
		t.Errorf(
			"did not get expected root hash: got %s, expected %s",
			trie.Hash().String(),
			expectedRootHash,
		)
	}
	trie.Set([]byte{0xab, 0xcd}, []byte{0x01, 0x23})
	hash2 := trie.Hash()
	if hash1 != hash2 {
		t.Errorf(
			"root hash changed when setting same value: got %s, expected %s",
			hash2,
			hash1,
		)
	}
}

func TestTrieExpectedHash(t *testing.T) {
	type testDefValue struct {
		key              []byte
		value            []byte
		expectedRootHash string
	}
	testDefs := []struct {
		values           []testDefValue
		expectedRootHash string
	}{
		{
			values: []testDefValue{
				{
					key:              []byte{0xab, 0xcd},
					value:            []byte{0x01, 0x23},
					expectedRootHash: "eb258590dda64098b24091629f9dbcaf7e6e55011f9a411deb9e5a9793f0d83f",
				},
				{
					key:              []byte{0xaa, 0xff},
					value:            []byte{0x45, 0x67},
					expectedRootHash: "6eddba467ac9132f619b06f6bc8577ae4a3a7d64632fe4d7d24b0ad9e58769b4",
				},
			},
		},
		// From https://github.com/blinklabs-io/merkle-patricia-forestry/issues/93
		// TODO: uncomment when we can pass this
		/*
			{
				expectedRootHash: "01b252f957e3138467c540ba230723c16b32d2bfe7f33dd54e8a7ab5d7ca02e9",
				values: []testDefValue{
					{
						key:   []byte(`81`),
						value: []byte(`1`),
					},
					{
						key:   []byte(`189`),
						value: []byte(`2`),
					},
				},
			},
		*/
		// From https://github.com/blinklabs-io/merkle-patricia-forestry/issues/96
		{
			expectedRootHash: "48b5cbcfa858fd9e8b49b6041d1833f842f30c1fec654d42c640056db5ce5afb",
			values: []testDefValue{
				{
					key:   []byte(`81`),
					value: []byte(`11`),
				},
				{
					key:   []byte(`189`),
					value: []byte(`11`),
				},
				{
					key:   []byte(`82`),
					value: []byte(`11`),
				},
			},
		},
	}
	for _, testDef := range testDefs {
		trie := NewTrie()
		for _, value := range testDef.values {
			trie.Set(value.key, value.value)
			if value.expectedRootHash == "" {
				continue
			}
			if trie.Hash().String() != value.expectedRootHash {
				t.Errorf(
					"did not get expected root hash: got %s, expected %s",
					trie.Hash().String(),
					value.expectedRootHash,
				)
			}

		}
		if testDef.expectedRootHash == "" {
			continue
		}
		if trie.Hash().String() != testDef.expectedRootHash {
			t.Errorf(
				"did not get expected root hash: got %s, expected %s",
				trie.Hash().String(),
				testDef.expectedRootHash,
			)
		}
	}
}

func TestTrieFruitsExpectedHash(t *testing.T) {
	trie := NewTrie()
	for _, entry := range fruitsTestEntries {
		trie.Set([]byte(entry.key), []byte(entry.value))
	}
	if trie.Hash().String() != fruitsExpectedHash {
		t.Errorf(
			"did not get expected root hash: got %s, expected %s",
			trie.Hash().String(),
			fruitsExpectedHash,
		)
	}
}

func TestTrieFruitsGet(t *testing.T) {
	trie := NewTrie()
	for _, entry := range fruitsTestEntries {
		trie.Set([]byte(entry.key), []byte(entry.value))
	}
	for _, entry := range fruitsTestEntries {
		tmpVal, err := trie.Get([]byte(entry.key))
		if err != nil {
			t.Fatalf("unexpected error getting key: %s", err)
		}
		if string(tmpVal) != entry.value {
			t.Fatalf(
				"did not get expected value: got %x, expected %x",
				tmpVal,
				entry.value,
			)
		}
	}
}

func TestTrieFruitsSetDeleteConsistentHash(t *testing.T) {
	trie := NewTrie()
	hashes := []Hash{}
	for _, entry := range fruitsTestEntries {
		hashes = append(hashes, trie.Hash())
		trie.Set([]byte(entry.key), []byte(entry.value))
	}
	for i := len(fruitsTestEntries) - 1; i >= 0; i-- {
		entry := fruitsTestEntries[i]
		if err := trie.Delete([]byte(entry.key)); err != nil {
			t.Fatalf("unexpected error deleting key: %s", err)
		}
		if trie.Hash() != hashes[i] {
			t.Fatalf(
				"did not get expected hash: got %s, expected %s",
				trie.Hash().String(),
				hashes[i].String(),
			)
		}
	}
}
