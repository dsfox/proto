package emit

// ice9's own lines in the generated code. updateMlsMailbox is hand-written
// (mtproto/mls_update.go) and never in the schema; the Update codec must
// still know it (#93). GetClazzID's fallback for a client ahead of the
// schema is in ClassNameRegisters.

// ice9EncodeCase is the case appended to a type's Encode, by TL type.
var ice9EncodeCase = map[string]string{
	"Update": `
	case Predicate_updateMlsMailbox: // hand-written, see mls_update.go
		t := MakeTLUpdateMlsMailbox(m)
		err := t.Encode(x, layer)
		if err != nil {
			return err
		}
`,
}

// ice9DecodeCase is the case a type's Decode gets after the case of an id.
var ice9DecodeCase = map[string]map[uint32]string{
	"Update": {
		0x8951abef: `	case 0xb305b464: // hand-written, see mls_update.go
		m2 := MakeTLUpdateMlsMailbox(m)
		m2.Decode(dBuf)
`,
	},
}
