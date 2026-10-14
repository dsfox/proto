package emit

import (
	"bytes"
	"fmt"

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
	for _, group := range [][]*schema.Entity{
		m.Service.Entities,
		Constructors(m.API),
		Constructors(m.Teamgram),
		Methods(m.API),
		Methods(m.Teamgram),
	} {
		for _, e := range group {
			ids := m.IDs(e)
			for _, id := range ids {
				name := m.Names[e]
				if len(ids) > 1 {
					name = fmt.Sprintf("%s_%x", name, id)
				}
				comment := fmt.Sprintf("%x", id)
				if v := m.Version(e, id); v != nil && !v.Explicit {
					comment = "0x" + comment
				}
				fmt.Fprintf(&b, "  CRC32_%s = %d; // %s\n", name, int32(id), comment)
			}
		}
	}
	b.WriteString("}\n")
	return b.Bytes()
}
