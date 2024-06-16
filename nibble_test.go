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
	"reflect"
	"testing"
)

func TestNibbleByteToNibble(t *testing.T) {
	var testByte byte = 0xab
	expectedNibbles := []Nibble{0xa, 0xb}
	nibbles := byteToNibbles(testByte)
	if !reflect.DeepEqual(nibbles, expectedNibbles) {
		t.Errorf("did not get expected nibbles: got %#v, expected %#v", nibbles, expectedNibbles)
	}
}

func TestNibbleBytesToNibbles(t *testing.T) {
	testBytes := []byte{0xab, 0xcd, 0xef, 0x00, 0x01}
	expectedNibbles := []Nibble{0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x0, 0x0, 0x0, 0x1}
	nibbles := bytesToNibbles(testBytes)
	if !reflect.DeepEqual(nibbles, expectedNibbles) {
		t.Errorf("did not get expected nibbles: got %#v, expected %#v", nibbles, expectedNibbles)
	}
}
