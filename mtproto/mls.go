package mtproto

// Our own methods, for end-to-end encryption on MLS (2bytes #36).
//
// This file is the reason this module is forked. Everything else could live in
// the server: the iOS and Android API layers are ordinary code and our methods
// sit in files of their own there. But incoming objects are built here, from a
// map this package keeps private, and there is no way in from outside.
//
// The constructor ids are computed the way TL computes them - CRC32 of the
// declaration - so they are stable, reproducible from the text above each type,
// and were checked against every id this module already knows.
//
// Kept in one file on purpose. When this module is updated from upstream, this
// is the only thing to carry across.

import "fmt"

const (
	// mls.publishKeyPackages key_packages:Vector<bytes> last_resort:bytes = mls.PublishResult;
	CRC32_mls_publishKeyPackages = int32(0x38115310)

	// mls.publishResult added:int available:int should_refill:Bool = mls.PublishResult;
	CRC32_mls_publishResult = int32(-1429076441) // 0xaacbf827

	// mls.claimKeyPackages user_id:long = mls.KeyPackages;
	CRC32_mls_claimKeyPackages = int32(0x054c3049)

	// mls.keyPackages packages:Vector<bytes> = mls.KeyPackages;
	CRC32_mls_keyPackages = int32(-548557587) // 0xdf5408ed
)

func (m *TLMlsPublishKeyPackages) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_publishKeyPackages)
	x.Int(int32(0x1cb5c415)) // vector
	x.Int(int32(len(m.KeyPackages)))
	for _, p := range m.KeyPackages {
		x.StringBytes(p)
	}
	x.StringBytes(m.LastResort)
	return nil
}

func (m *TLMlsPublishKeyPackages) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.publishKeyPackages: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		// A length is attacker-controlled until it has been looked at.
		return fmt.Errorf("mls.publishKeyPackages: %d key packages is not a number of key packages", count)
	}
	m.KeyPackages = make([][]byte, 0, count)
	for i := int32(0); i < count; i++ {
		m.KeyPackages = append(m.KeyPackages, dBuf.StringBytes())
	}
	m.LastResort = dBuf.StringBytes()
	return dBuf.err
}

func (m *Mls_PublishResult) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_publishResult)
	x.Int(m.Added)
	x.Int(m.Available)
	if m.ShouldRefill {
		x.Int(int32(-1720552011)) // boolTrue, 0x997275b5
	} else {
		x.Int(int32(-1132882121)) // boolFalse, 0xbc799737
	}
	return nil
}

func (m *Mls_PublishResult) Decode(dBuf *DecodeBuf) error {
	m.Added = dBuf.Int()
	m.Available = dBuf.Int()
	m.ShouldRefill = dBuf.Int() == int32(-1720552011)
	return dBuf.err
}

func (m *TLMlsClaimKeyPackages) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_claimKeyPackages)
	x.Long(m.UserId)
	return nil
}

func (m *TLMlsClaimKeyPackages) Decode(dBuf *DecodeBuf) error {
	m.UserId = dBuf.Long()
	return dBuf.err
}

func (m *Mls_KeyPackages) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_keyPackages)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Packages)))
	for _, p := range m.Packages {
		x.StringBytes(p)
	}
	return nil
}

func (m *Mls_KeyPackages) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.keyPackages: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.keyPackages: %d packages is not a number of packages", count)
	}
	m.Packages = make([][]byte, 0, count)
	for i := int32(0); i < count; i++ {
		m.Packages = append(m.Packages, dBuf.StringBytes())
	}
	return dBuf.err
}

// Registered here rather than in the generated table, so that regenerating that
// table from upstream cannot silently drop them.
func init() {
	clazzIdRegisters2[CRC32_mls_publishKeyPackages] = func() TLObject { return &TLMlsPublishKeyPackages{} }
	clazzIdRegisters2[CRC32_mls_publishResult] = func() TLObject { return &Mls_PublishResult{} }
	clazzIdRegisters2[CRC32_mls_claimKeyPackages] = func() TLObject { return &TLMlsClaimKeyPackages{} }
	clazzIdRegisters2[CRC32_mls_keyPackages] = func() TLObject { return &Mls_KeyPackages{} }

	// And where each request goes. Without this the proxy has nowhere to send
	// them and answers "not found method", which names the symptom and not the
	// missing line.
	rpcContextRegisters["TLMlsPublishKeyPackages"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_publishKeyPackages",
		func() interface{} { return new(Mls_PublishResult) },
	}
	rpcContextRegisters["TLMlsClaimKeyPackages"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_claimKeyPackages",
		func() interface{} { return new(Mls_KeyPackages) },
	}
}

// MlsConstructorNames is what the log and the tests use to name our methods,
// since they are not in the generated tables that name everything else.
func MlsConstructorNames() map[int32]string {
	return map[int32]string{
		CRC32_mls_publishKeyPackages: "mls.publishKeyPackages",
		CRC32_mls_publishResult:      "mls.publishResult",
		CRC32_mls_claimKeyPackages:   "mls.claimKeyPackages",
		CRC32_mls_keyPackages:        "mls.keyPackages",
	}
}
