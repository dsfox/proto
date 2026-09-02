package mtproto

// The one thing the server says to a device on its own: there is something in
// your box.
//
// No payload, on purpose. The box is fetched through mls.getWelcomes and
// mls.getCommits and emptied through their confirmations, so there is exactly
// one path to apply a commit or a welcome and exactly one to say it was
// applied. An update that carried the bytes would be a second path, and a
// second chance to apply the same commit twice.
//
// Why it exists: commits were polled while messages were pushed, so after any
// change of membership the first message overtook the commit that opens it on
// every phone but the committer's, and stood on the screen as a lock for the
// round trip it took to fetch that commit (#156).
//
// Safe towards a client that does not know it: it carries no pts, so it is
// never part of a difference, and both clients drop an update they cannot parse
// and go on - measured on both before this was written.
//
// updateMlsMailbox = Update;
const CRC32_updateMlsMailbox = int32(-1291471772) // 0xb305b464

const Predicate_updateMlsMailbox = "updateMlsMailbox"

type TLUpdateMlsMailbox struct {
	Data2 *Update
}

func MakeTLUpdateMlsMailbox(data2 *Update) *TLUpdateMlsMailbox {
	if data2 == nil {
		data2 = &Update{}
	}
	data2.PredicateName = Predicate_updateMlsMailbox
	data2.Constructor = TLConstructor(CRC32_updateMlsMailbox)
	return &TLUpdateMlsMailbox{Data2: data2}
}

func (m *TLUpdateMlsMailbox) To_Update() *Update {
	m.Data2.PredicateName = Predicate_updateMlsMailbox
	return m.Data2
}

func (m *TLUpdateMlsMailbox) GetPredicateName() string {
	return Predicate_updateMlsMailbox
}

func (m *TLUpdateMlsMailbox) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_updateMlsMailbox)
	return nil
}

func (m *TLUpdateMlsMailbox) CalcByteSize(layer int32) int {
	return 4
}

func (m *TLUpdateMlsMailbox) Decode(dBuf *DecodeBuf) error {
	return dBuf.GetError()
}

func (m *TLUpdateMlsMailbox) String() string {
	return `{"@type":"updateMlsMailbox"}`
}

func (m *TLUpdateMlsMailbox) DebugString() string {
	return m.String()
}

func init() {
	clazzIdRegisters2[CRC32_updateMlsMailbox] = func() TLObject {
		return MakeTLUpdateMlsMailbox(nil)
	}
	// So that Update.Encode can find the constructor by name at any layer.
	clazzNameRegisters2[Predicate_updateMlsMailbox] = map[int]int32{0: CRC32_updateMlsMailbox}
	clazzIdNameRegisters2[CRC32_updateMlsMailbox] = Predicate_updateMlsMailbox
}
