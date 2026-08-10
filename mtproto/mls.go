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
	CRC32_mls_publishKeyPackages = int32(940659472) // 0x38115310

	// mls.publishResult added:int available:int should_refill:Bool = mls.PublishResult;
	CRC32_mls_publishResult = int32(-1429473241) // 0xaacbf827

	// mls.claimKeyPackages user_id:long = mls.KeyPackages;
	CRC32_mls_claimKeyPackages = int32(88879177) // 0x054c3049

	// mls.keyPackages packages:Vector<bytes> = mls.KeyPackages;
	CRC32_mls_keyPackages = int32(-548140819) // 0xdf5408ed

	// mls.sendWelcome user_id:long welcome:bytes = mls.Ok;
	CRC32_mls_sendWelcome = int32(-773834602) // 0xd1e03896

	// mls.ok ok:Bool = mls.Ok;
	CRC32_mls_ok = int32(-1518331278) // 0xa5801a72

	// mls.getWelcomes = mls.Welcomes;
	CRC32_mls_getWelcomes = int32(-512239425) // 0xe177d8bf

	// mls.welcomes welcomes:Vector<mls.Welcome> = mls.Welcomes;
	CRC32_mls_welcomes = int32(-1921518262) // 0x8d77f54a

	// mls.welcome id:long from_id:long welcome:bytes = mls.Welcome;
	CRC32_mls_welcome = int32(-180214709) // 0xf542244b

	// mls.confirmWelcomes ids:Vector<long> = mls.Ok;
	CRC32_mls_confirmWelcomes = int32(-1226029994) // 0xb6ec4456
)

// A welcome is what lets a device into a conversation somebody started with it.
// It travels through its own methods rather than as a message, so that no
// client has to hide anything from a chat list.

func (m *TLMlsSendWelcome) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_sendWelcome)
	x.Long(m.UserId)
	x.StringBytes(m.Welcome)
	return nil
}

func (m *TLMlsSendWelcome) Decode(dBuf *DecodeBuf) error {
	m.UserId = dBuf.Long()
	m.Welcome = dBuf.StringBytes()
	return dBuf.err
}

func (m *Mls_Ok) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_ok)
	if m.Ok {
		x.Int(int32(-1720552011)) // boolTrue
	} else {
		x.Int(int32(-1132882121)) // boolFalse
	}
	return nil
}

func (m *Mls_Ok) Decode(dBuf *DecodeBuf) error {
	m.Ok = dBuf.Int() == int32(-1720552011)
	return dBuf.err
}

func (m *TLMlsGetWelcomes) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_getWelcomes)
	return nil
}

func (m *TLMlsGetWelcomes) Decode(dBuf *DecodeBuf) error {
	return dBuf.err
}

func (m *Mls_Welcome) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_welcome)
	x.Long(m.Id)
	x.Long(m.FromId)
	x.StringBytes(m.Welcome)
	return nil
}

func (m *Mls_Welcome) Decode(dBuf *DecodeBuf) error {
	m.Id = dBuf.Long()
	m.FromId = dBuf.Long()
	m.Welcome = dBuf.StringBytes()
	return dBuf.err
}

func (m *Mls_Welcomes) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_welcomes)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Welcomes)))
	for _, w := range m.Welcomes {
		if err := w.Encode(x, layer); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mls_Welcomes) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.welcomes: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.welcomes: %d welcomes is not a number of welcomes", count)
	}
	m.Welcomes = make([]*Mls_Welcome, 0, count)
	for i := int32(0); i < count; i++ {
		if id := dBuf.Int(); id != CRC32_mls_welcome {
			return fmt.Errorf("mls.welcomes: expected a welcome, got %d", id)
		}
		w := &Mls_Welcome{}
		if err := w.Decode(dBuf); err != nil {
			return err
		}
		m.Welcomes = append(m.Welcomes, w)
	}
	return dBuf.err
}

func (m *TLMlsConfirmWelcomes) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_confirmWelcomes)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Ids)))
	for _, id := range m.Ids {
		x.Long(id)
	}
	return nil
}

func (m *TLMlsConfirmWelcomes) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.confirmWelcomes: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.confirmWelcomes: %d ids is not a number of ids", count)
	}
	m.Ids = make([]int64, 0, count)
	for i := int32(0); i < count; i++ {
		m.Ids = append(m.Ids, dBuf.Long())
	}
	return dBuf.err
}

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
	clazzIdRegisters2[CRC32_mls_sendWelcome] = func() TLObject { return &TLMlsSendWelcome{} }
	clazzIdRegisters2[CRC32_mls_ok] = func() TLObject { return &Mls_Ok{} }
	clazzIdRegisters2[CRC32_mls_getWelcomes] = func() TLObject { return &TLMlsGetWelcomes{} }
	clazzIdRegisters2[CRC32_mls_welcomes] = func() TLObject { return &Mls_Welcomes{} }
	clazzIdRegisters2[CRC32_mls_welcome] = func() TLObject { return &Mls_Welcome{} }
	clazzIdRegisters2[CRC32_mls_confirmWelcomes] = func() TLObject { return &TLMlsConfirmWelcomes{} }

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
	rpcContextRegisters["TLMlsSendWelcome"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_sendWelcome",
		func() interface{} { return new(Mls_Ok) },
	}
	rpcContextRegisters["TLMlsGetWelcomes"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_getWelcomes",
		func() interface{} { return new(Mls_Welcomes) },
	}
	rpcContextRegisters["TLMlsConfirmWelcomes"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_confirmWelcomes",
		func() interface{} { return new(Mls_Ok) },
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
		CRC32_mls_sendWelcome:        "mls.sendWelcome",
		CRC32_mls_ok:                 "mls.ok",
		CRC32_mls_getWelcomes:        "mls.getWelcomes",
		CRC32_mls_welcomes:           "mls.welcomes",
		CRC32_mls_welcome:            "mls.welcome",
		CRC32_mls_confirmWelcomes:    "mls.confirmWelcomes",
	}
}
