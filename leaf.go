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

type Leaf struct {
	hash   Hash
	suffix []Nibble
	value  []byte
}

func newLeaf(suffix []Nibble, value []byte) *Leaf {
	l := &Leaf{}
	if suffix != nil {
		l.suffix = append(l.suffix, suffix...)
	}
	l.Set(value)
	return l
}

func (l Leaf) isChildNode() {}

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

func (l *Leaf) updateHash() {
	valueHash := HashValue(l.value)
	head := hashHead(l.suffix)
	tail := hashTail(l.suffix)
	tmpVal := append(head, tail...)
	tmpVal = append(tmpVal, valueHash.Bytes()...)
	l.hash = HashValue(tmpVal)
}

func hashHead(suffix []Nibble) []byte {
	if len(suffix)%2 == 0 {
		// Return 0xff for even length
		return []byte{0xff}
	} else {
		// Return 0x00 and first nibble for odd length
		return []byte{0x00, byte(suffix[0])}
	}
}

func hashTail(suffix []Nibble) []byte {
	var ret []byte
	if len(suffix)%2 == 0 {
		// Return entire suffix for even length
		for _, nibble := range suffix {
			ret = append(ret, byte(nibble))
		}
	} else {
		// Return suffix minus first nibble for odd length
		for _, nibble := range suffix[1:] {
			ret = append(ret, byte(nibble))
		}
	}
	return ret
}
