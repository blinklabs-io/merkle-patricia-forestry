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

import "fmt"

type Leaf struct {
	hash   Hash
	suffix []Nibble
	key    []byte
	value  []byte
}

func newLeaf(suffix []Nibble, key []byte, value []byte) *Leaf {
	l := &Leaf{}
	if suffix != nil {
		l.suffix = append(l.suffix, suffix...)
	}
	if key != nil {
		l.key = append(l.key, key...)
	}
	l.Set(value)
	return l
}

func (l *Leaf) isNode() {}

func (l *Leaf) String() string {
	return fmt.Sprintf(
		"%s #%s { %s (%x) -> %s (%x) }",
		nibblesToHexString(l.suffix),
		l.hash.String()[:10],
		l.key,
		l.key,
		l.value,
		l.value,
	)
}

func (l *Leaf) Hash() Hash {
	return l.hash
}

func (l *Leaf) Value() []byte {
	return l.value
}

func (l *Leaf) Set(value []byte) {
	l.value = make([]byte, len(value))
	copy(l.value, value)
	l.updateHash()
}

func (l *Leaf) generateProof(path []Nibble) (*Proof, error) {
	if string(path) != string(l.suffix) {
		return nil, ErrKeyNotExist
	}
	leafPath := keyToPath(l.key)
	var proofVal []byte
	if string(path) == string(l.suffix) {
		proofVal = append(proofVal, l.value...)
	}
	proof := newProof(
		leafPath,
		proofVal,
	)
	return proof, nil
}

func (l *Leaf) updateHash() {
	tmpVal := []byte{}
	head := hashHead(l.suffix)
	tmpVal = append(tmpVal, head...)
	tail := hashTail(l.suffix)
	tmpVal = append(tmpVal, tail...)
	valueHash := HashValue(l.value)
	tmpVal = append(tmpVal, valueHash.Bytes()...)
	l.hash = HashValue(tmpVal)
}

func hashHead(suffix []Nibble) []byte {
	if len(suffix)%2 == 0 {
		// Return 0xff for even length
		return []byte{0xff}
	} else {
		// Return 0x0 and first nibble for odd length
		return []byte{0x0, byte(suffix[0])}
	}
}

func hashTail(suffix []Nibble) []byte {
	if len(suffix)%2 == 0 {
		// Return entire suffix for even length
		return nibblesToBytes(suffix)
	} else {
		// Return suffix minus first nibble for odd length
		return nibblesToBytes(suffix[1:])
	}
}
