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
	"fmt"
)

// InvalidLengthError is returned when decoder reads invalid length.
type InvalidLengthError struct {
	Length int
	Where  string
}

func (i *InvalidLengthError) Error() string {
	return fmt.Sprintf("invalid %s length: %d", i.Where, i.Length)
}

// UnexpectedClazzIDErr means that unknown or unexpected type id was decoded.
type UnexpectedClazzIDErr struct {
	ClazzID uint32
}

func (e *UnexpectedClazzIDErr) Error() string {
	return fmt.Sprintf("unexpected clazzID %#x", uint32(e.ClazzID))
}

// NewUnexpectedClazzID return new UnexpectedClazzIDErr.
func NewUnexpectedClazzID(id uint32) error {
	return &UnexpectedClazzIDErr{ClazzID: id}
}
