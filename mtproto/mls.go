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
	CRC32_mls_publishKeyPackages = int32(-913436181) // 0xc98e11eb

	// mls.publishResult added:int available:int should_refill:Bool devices:int = mls.PublishResult;
	CRC32_mls_publishResult = int32(-472421573) // 0xe3d76b3b

	// mls.claimKeyPackages user_id:long = mls.KeyPackages;
	CRC32_mls_claimKeyPackages = int32(88879177) // 0x054c3049

	// mls.keyPackages packages:Vector<bytes> = mls.KeyPackages;
	CRC32_mls_keyPackages = int32(-548140819) // 0xdf5408ed

	// mls.sendWelcome user_id:long peer_id:long welcome:bytes = mls.Ok;
	CRC32_mls_sendWelcome = int32(2042714623) // 0x79c159ff

	// mls.ok ok:Bool = mls.Ok;
	CRC32_mls_ok = int32(-1518331278) // 0xa5801a72

	// mls.claimConversation peer_id:long group_id:bytes = mls.Conversation;
	CRC32_mls_claimConversation = int32(-1101602434) // 0xbe56e17e
	// mls.conversation peer_id:long group_id:bytes = mls.Conversation;
	CRC32_mls_conversation = int32(622211617) // 0x25163221
	// mls.devicesOf users:Vector<long> = mls.DeviceCounts;
	CRC32_mls_devicesOf = int32(-657797125) // 0xd8cacffb
	// mls.deviceCounts counts:Vector<int> names:Vector<bytes> = mls.DeviceCounts;
	CRC32_mls_deviceCounts = int32(1890672928) // 0x70b16120
	// mls.getWelcomes = mls.Welcomes;
	CRC32_mls_getWelcomes = int32(-512239425) // 0xe177d8bf

	// mls.welcomes welcomes:Vector<mls.Welcome> = mls.Welcomes;
	CRC32_mls_welcomes = int32(-1921518262) // 0x8d77f54a

	// mls.welcome id:long from_id:long welcome:bytes = mls.Welcome;
	CRC32_mls_welcome = int32(215890102) // 0x0cde38b6

	// mls.confirmWelcomes ids:Vector<long> = mls.Ok;
	CRC32_mls_confirmWelcomes = int32(-1226029994) // 0xb6ec4456

	// mls.setRecoverySecret secret:string = mls.Ok;
	CRC32_mls_setRecoverySecret = int32(-369099376) // 0xe9fffd90

	// mls.sendCommit group_id:bytes epoch:long members:Vector<long> commit:bytes = mls.CommitResult;
	CRC32_mls_sendCommit = int32(-945155929) // 0xc7aa10a7

	// mls.commitResult accepted:Bool epoch:long = mls.CommitResult;
	CRC32_mls_commitResult = int32(191372459) // 0x0b681cab

	// mls.getCommits = mls.Commits;
	CRC32_mls_getCommits = int32(1356576713) // 0x50dbb7c9

	// mls.commits commits:Vector<mls.Commit> = mls.Commits;
	CRC32_mls_commits = int32(-902742102) // 0xca313faa

	// mls.commit id:long from_id:long group_id:bytes epoch:long commit:bytes = mls.Commit;
	CRC32_mls_commit = int32(-130530128) // 0xf83844b0

	// mls.confirmCommits ids:Vector<long> = mls.Ok;
	CRC32_mls_confirmCommits = int32(96655983) // 0x05c2da6f
)

// A welcome is what lets a device into a conversation somebody started with it.
// It travels through its own methods rather than as a message, so that no
// client has to hide anything from a chat list.

func (m *TLMlsSendWelcome) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_sendWelcome)
	x.Long(m.UserId)
	// Which chat it is for, so the device joining does not have to guess from
	// who sent it - guessing filed a group as the conversation with whoever
	// invited them (#115).
	x.Long(m.PeerId)
	x.StringBytes(m.Welcome)
	return nil
}

func (m *TLMlsSendWelcome) Decode(dBuf *DecodeBuf) error {
	m.UserId = dBuf.Long()
	m.PeerId = dBuf.Long()
	m.Welcome = dBuf.StringBytes()
	return dBuf.err
}

// What a phone hands over in place of its recovery phrase: a derivation of the
// words, which is enough to recognise somebody typing them and nothing else.
func (m *TLMlsSetRecoverySecret) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_setRecoverySecret)
	x.String(m.Secret)
	return nil
}

func (m *TLMlsSetRecoverySecret) Decode(dBuf *DecodeBuf) error {
	m.Secret = dBuf.String()
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
	x.Long(m.PeerId)
	x.StringBytes(m.Welcome)
	return nil
}

func (m *Mls_Welcome) Decode(dBuf *DecodeBuf) error {
	m.Id = dBuf.Long()
	m.FromId = dBuf.Long()
	m.PeerId = dBuf.Long()
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

func (m *TLMlsClaimConversation) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_claimConversation)
	x.Long(m.PeerId)
	x.StringBytes(m.GroupId)
	return nil
}

func (m *TLMlsClaimConversation) Decode(dBuf *DecodeBuf) error {
	m.PeerId = dBuf.Long()
	m.GroupId = dBuf.StringBytes()
	return dBuf.err
}

func (m *Mls_Conversation) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_conversation)
	x.Long(m.PeerId)
	x.StringBytes(m.GroupId)
	return nil
}

func (m *Mls_Conversation) Decode(dBuf *DecodeBuf) error {
	m.PeerId = dBuf.Long()
	m.GroupId = dBuf.StringBytes()
	return dBuf.err
}

func (m *TLMlsDevicesOf) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_devicesOf)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Users)))
	for _, id := range m.Users {
		x.Long(id)
	}
	return nil
}

func (m *TLMlsDevicesOf) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.devicesOf: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.devicesOf: %d people is not a number of people", count)
	}
	m.Users = make([]int64, 0, count)
	for i := int32(0); i < count; i++ {
		m.Users = append(m.Users, dBuf.Long())
	}
	return dBuf.err
}

func (m *Mls_DeviceCounts) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_deviceCounts)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Counts)))
	for _, n := range m.Counts {
		x.Int(n)
	}
	// And the names of those devices, all of them in one list: the counts say
	// where to cut it. One entry per counted device, empty when the device
	// published before key packages said whose they are (#136).
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Names)))
	for _, name := range m.Names {
		x.StringBytes(name)
	}
	return nil
}

func (m *Mls_DeviceCounts) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.deviceCounts: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.deviceCounts: %d counts is not a number of counts", count)
	}
	m.Counts = make([]int32, 0, count)
	total := int32(0)
	for i := int32(0); i < count; i++ {
		n := dBuf.Int()
		m.Counts = append(m.Counts, n)
		if n > 0 {
			total += n
		}
	}

	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.deviceCounts: expected a vector of names, got %d", v)
	}
	named := dBuf.Int()
	// The counts are what cuts this list, so a list of another length cannot be
	// cut at all. Refused here rather than handed on: a caller that split it
	// anyway would attribute one person's devices to the next.
	if named != total {
		return fmt.Errorf("mls.deviceCounts: %d name(s) for %d device(s)", named, total)
	}
	m.Names = make([][]byte, 0, named)
	for i := int32(0); i < named; i++ {
		m.Names = append(m.Names, dBuf.StringBytes())
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
	x.StringBytes(m.Name)
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
	m.Name = dBuf.StringBytes()
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
	// How many devices of this account have published anything. It is what tells
	// a phone that another phone of the same person has signed in since the
	// conversation started - the comparison of members is about people and
	// cannot see it (#41).
	x.Int(m.Devices)
	return nil
}

func (m *Mls_PublishResult) Decode(dBuf *DecodeBuf) error {
	m.Added = dBuf.Int()
	m.Available = dBuf.Int()
	m.ShouldRefill = dBuf.Int() == int32(-1720552011)
	m.Devices = dBuf.Int()
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
// A commit moves a conversation to its next epoch. It travels through its own
// methods for the same reason a welcome does: handshake traffic never touches
// the message pipeline, so no client has to hide it from a chat list.

func (m *TLMlsSendCommit) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_sendCommit)
	x.StringBytes(m.GroupId)
	x.Long(m.Epoch)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Members)))
	for _, id := range m.Members {
		x.Long(id)
	}
	x.StringBytes(m.Commit)
	return nil
}

func (m *TLMlsSendCommit) Decode(dBuf *DecodeBuf) error {
	m.GroupId = dBuf.StringBytes()
	m.Epoch = dBuf.Long()
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.sendCommit: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.sendCommit: %d members is not a number of members", count)
	}
	m.Members = make([]int64, 0, count)
	for i := int32(0); i < count; i++ {
		m.Members = append(m.Members, dBuf.Long())
	}
	m.Commit = dBuf.StringBytes()
	return dBuf.err
}

func (m *Mls_CommitResult) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_commitResult)
	if m.Accepted {
		x.Int(int32(-1720552011)) // boolTrue
	} else {
		x.Int(int32(-1132882121)) // boolFalse
	}
	x.Long(m.Epoch)
	return nil
}

func (m *Mls_CommitResult) Decode(dBuf *DecodeBuf) error {
	m.Accepted = dBuf.Int() == int32(-1720552011)
	m.Epoch = dBuf.Long()
	return dBuf.err
}

func (m *TLMlsGetCommits) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_getCommits)
	return nil
}

func (m *TLMlsGetCommits) Decode(dBuf *DecodeBuf) error {
	return dBuf.err
}

func (m *Mls_Commit) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_commit)
	x.Long(m.Id)
	x.Long(m.FromId)
	x.StringBytes(m.GroupId)
	x.Long(m.Epoch)
	x.StringBytes(m.Commit)
	return nil
}

func (m *Mls_Commit) Decode(dBuf *DecodeBuf) error {
	m.Id = dBuf.Long()
	m.FromId = dBuf.Long()
	m.GroupId = dBuf.StringBytes()
	m.Epoch = dBuf.Long()
	m.Commit = dBuf.StringBytes()
	return dBuf.err
}

func (m *Mls_Commits) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_commits)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Commits)))
	for _, c := range m.Commits {
		if err := c.Encode(x, layer); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mls_Commits) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.commits: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.commits: %d commits is not a number of commits", count)
	}
	m.Commits = make([]*Mls_Commit, 0, count)
	for i := int32(0); i < count; i++ {
		if id := dBuf.Int(); id != CRC32_mls_commit {
			return fmt.Errorf("mls.commits: expected a commit, got %d", id)
		}
		c := &Mls_Commit{}
		if err := c.Decode(dBuf); err != nil {
			return err
		}
		m.Commits = append(m.Commits, c)
	}
	return dBuf.err
}

func (m *TLMlsConfirmCommits) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_mls_confirmCommits)
	x.Int(int32(0x1cb5c415))
	x.Int(int32(len(m.Ids)))
	for _, id := range m.Ids {
		x.Long(id)
	}
	return nil
}

func (m *TLMlsConfirmCommits) Decode(dBuf *DecodeBuf) error {
	if v := dBuf.Int(); v != int32(0x1cb5c415) {
		return fmt.Errorf("mls.confirmCommits: expected a vector, got %d", v)
	}
	count := dBuf.Int()
	if count < 0 || count > 1000 {
		return fmt.Errorf("mls.confirmCommits: %d ids is not a number of ids", count)
	}
	m.Ids = make([]int64, 0, count)
	for i := int32(0); i < count; i++ {
		m.Ids = append(m.Ids, dBuf.Long())
	}
	return dBuf.err
}

func init() {
	clazzIdRegisters2[CRC32_mls_publishKeyPackages] = func() TLObject { return &TLMlsPublishKeyPackages{} }
	clazzIdRegisters2[CRC32_mls_publishResult] = func() TLObject { return &Mls_PublishResult{} }
	clazzIdRegisters2[CRC32_mls_claimKeyPackages] = func() TLObject { return &TLMlsClaimKeyPackages{} }
	clazzIdRegisters2[CRC32_mls_keyPackages] = func() TLObject { return &Mls_KeyPackages{} }
	clazzIdRegisters2[CRC32_mls_sendWelcome] = func() TLObject { return &TLMlsSendWelcome{} }
	clazzIdRegisters2[CRC32_mls_setRecoverySecret] = func() TLObject { return &TLMlsSetRecoverySecret{} }
	clazzIdRegisters2[CRC32_mls_ok] = func() TLObject { return &Mls_Ok{} }
	clazzIdRegisters2[CRC32_mls_claimConversation] = func() TLObject { return &TLMlsClaimConversation{} }
	clazzIdRegisters2[CRC32_mls_conversation] = func() TLObject { return &Mls_Conversation{} }
	clazzIdRegisters2[CRC32_mls_devicesOf] = func() TLObject { return &TLMlsDevicesOf{} }
	clazzIdRegisters2[CRC32_mls_deviceCounts] = func() TLObject { return &Mls_DeviceCounts{} }
	clazzIdRegisters2[CRC32_mls_getWelcomes] = func() TLObject { return &TLMlsGetWelcomes{} }
	clazzIdRegisters2[CRC32_mls_welcomes] = func() TLObject { return &Mls_Welcomes{} }
	clazzIdRegisters2[CRC32_mls_welcome] = func() TLObject { return &Mls_Welcome{} }
	clazzIdRegisters2[CRC32_mls_confirmWelcomes] = func() TLObject { return &TLMlsConfirmWelcomes{} }
	clazzIdRegisters2[CRC32_mls_sendCommit] = func() TLObject { return &TLMlsSendCommit{} }
	clazzIdRegisters2[CRC32_mls_commitResult] = func() TLObject { return &Mls_CommitResult{} }
	clazzIdRegisters2[CRC32_mls_getCommits] = func() TLObject { return &TLMlsGetCommits{} }
	clazzIdRegisters2[CRC32_mls_commits] = func() TLObject { return &Mls_Commits{} }
	clazzIdRegisters2[CRC32_mls_commit] = func() TLObject { return &Mls_Commit{} }
	clazzIdRegisters2[CRC32_mls_confirmCommits] = func() TLObject { return &TLMlsConfirmCommits{} }

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
	rpcContextRegisters["TLMlsClaimConversation"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_claimConversation",
		func() interface{} { return new(Mls_Conversation) },
	}
	rpcContextRegisters["TLMlsDevicesOf"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_devicesOf",
		func() interface{} { return new(Mls_DeviceCounts) },
	}
	rpcContextRegisters["TLMlsGetWelcomes"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_getWelcomes",
		func() interface{} { return new(Mls_Welcomes) },
	}
	rpcContextRegisters["TLMlsSendCommit"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_sendCommit",
		func() interface{} { return new(Mls_CommitResult) },
	}
	rpcContextRegisters["TLMlsGetCommits"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_getCommits",
		func() interface{} { return new(Mls_Commits) },
	}
	rpcContextRegisters["TLMlsConfirmCommits"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_confirmCommits",
		func() interface{} { return new(Mls_Ok) },
	}
	rpcContextRegisters["TLMlsConfirmWelcomes"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_confirmWelcomes",
		func() interface{} { return new(Mls_Ok) },
	}
	rpcContextRegisters["TLMlsSetRecoverySecret"] = RPCContextTuple{
		"/mtproto.RPCMls/mls_setRecoverySecret",
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
		CRC32_mls_setRecoverySecret:  "mls.setRecoverySecret",
		CRC32_mls_sendCommit:         "mls.sendCommit",
		CRC32_mls_commitResult:       "mls.commitResult",
		CRC32_mls_getCommits:         "mls.getCommits",
		CRC32_mls_commits:            "mls.commits",
		CRC32_mls_commit:             "mls.commit",
		CRC32_mls_confirmCommits:     "mls.confirmCommits",
		CRC32_mls_claimConversation:  "mls.claimConversation",
		CRC32_mls_conversation:       "mls.conversation",
		CRC32_mls_devicesOf:          "mls.devicesOf",
		CRC32_mls_deviceCounts:       "mls.deviceCounts",
	}
}
