package emit

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// protoHeader opens every generated .proto file.
const protoHeader = `/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright (c) 2026 The Teamgram Authors (https://teamgram.net).
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

syntax = "proto3";

package mtproto;
`

// CRC32Proto is schema.tl.crc32.proto before clang-format: the TLConstructor
// enum. The service types come first, then the constructors of the API and
// of teamgram, then the methods of both; a predicate with several ids has an
// enum value per id, newest first. The comment carries the id in hex, with
// "0x" when the line had no id and the generator computed it.
func CRC32Proto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(`
// import "google/protobuf/wrappers.proto";

option go_package = "github.com/teamgram/proto/mtproto";

enum TLConstructor {
  // option allow_alias = true;
  CRC32_UNKNOWN = 0;
  CRC32_vector = 481674261;
  CRC32_message2 = 1538843921;
  CRC32_msg_container = 1945237724;
  CRC32_msg_copy = 530561358;
  CRC32_gzip_packed = 812830625;
  CRC32_rpc_result = -212046591;
`)
	for _, c := range m.crc32Constants() {
		fmt.Fprintf(&b, "  CRC32_%s = %d; // %s\n", c.name, int32(c.id), c.comment)
	}
	b.WriteString("}\n")
	return b.Bytes()
}

// crc32Constant is one value of the TLConstructor enum.
type crc32Constant struct {
	name    string
	id      uint32
	comment string
}

// crc32Constants is the enum's values after the six parsed by hand: the
// service types, the constructors of the API and of teamgram, the methods of
// both; a predicate with several ids has a value per id, newest first.
func (m *Model) crc32Constants() []crc32Constant {
	var out []crc32Constant
	for _, group := range [][]*schema.Entity{
		m.Service,
		Constructors(m.API),
		Constructors(m.Teamgram),
		Methods(m.API),
		Methods(m.Teamgram),
	} {
		for _, e := range group {
			ids := e.IDs()
			for _, id := range ids {
				name := m.Names[e]
				if len(ids) > 1 {
					name = fmt.Sprintf("%s_%x", name, id)
				}
				out = append(out, crc32Constant{name: name, id: id, comment: hexComment(e, id)})
			}
		}
	}
	return out
}

// CRC32Helper is schema.tl.crc32_helper.pb.go before gofmt: the TLConstructor
// enum of schema.tl.crc32.proto as Go constants, in the same order with the
// same comments.
func CRC32Helper(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader[:strings.Index(protoHeader, "\nsyntax")])
	b.WriteString(`
package mtproto

const (
	CRC32_UNKNOWN TLConstructor = 0
	CRC32_vector TLConstructor = 481674261
	CRC32_message2 TLConstructor = 1538843921
	CRC32_msg_container TLConstructor = 1945237724
	CRC32_msg_copy TLConstructor = 530561358
	CRC32_gzip_packed TLConstructor = 812830625
	CRC32_rpc_result TLConstructor = -212046591
`)
	for _, c := range m.crc32Constants() {
		fmt.Fprintf(&b, "\tCRC32_%s TLConstructor = %d // %s\n", c.name, int32(c.id), c.comment)
	}
	b.WriteString(")\n")
	return b.Bytes()
}
