// Copyright (c) 2026 The Teamgram Authors (https://teamgram.net).
//  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bin

import (
	"io"
)

// encodeBytes is same as encodeString, but for bytes.
func (e *Encoder) encodeBytes(v []byte) {
	l := len(v)
	if l <= maxSmallStringLength {
		_ = e.w.WriteByte(byte(l))
		_, _ = e.w.Write(v)
		currentLen := l + 1
		_, _ = e.w.Write(make([]byte, nearestPaddedValueLength(currentLen)-currentLen))
	} else {
		_ = e.w.WriteByte(firstLongStringByte)
		_ = e.w.WriteByte(byte(l))
		_ = e.w.WriteByte(byte(l >> 8))
		_ = e.w.WriteByte(byte(l >> 16))
		_, _ = e.w.Write(v)
		currentLen := l + 4
		_, _ = e.w.Write(make([]byte, nearestPaddedValueLength(currentLen)-currentLen))
	}
}

// decodeBytes is same as decodeString, but for bytes.
//
// NB: v is slice of b.
func decodeBytes(b []byte) (n int, v []byte, err error) {
	if len(b) == 0 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	if b[0] == firstLongStringByte {
		if len(b) < 4 {
			return 0, nil, io.ErrUnexpectedEOF
		}
		strLen := uint32(b[1]) | uint32(b[2])<<8 | uint32(b[3])<<16
		if len(b) < (int(strLen) + 4) {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return nearestPaddedValueLength(int(strLen) + 4), b[4 : strLen+4], nil
	}
	strLen := int(b[0])
	if len(b) < (strLen + 1) {
		return 0, nil, io.ErrUnexpectedEOF
	}
	if strLen > maxSmallStringLength {
		return 0, nil, &InvalidLengthError{
			Length: strLen,
			Where:  "bytes",
		}
	}
	return nearestPaddedValueLength(strLen + 1), b[1 : strLen+1], nil
}
