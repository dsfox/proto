package mtproto

// The wire towards the phone for invitations people send themselves (#47).
//
// Hand-written like the mls.* methods: the constructor ids are the CRC32 of
// the declarations above each one, and tests/test_mls_constructors.py in the
// outer repository recomputes them from the text on every side that writes
// them down.

const (
	// invite.mint phone:string = invite.Minted;
	CRC32_invite_mint = int32(-734254852) // 0xd43c28fc

	// invite.minted code:string expires:int = invite.Minted;
	CRC32_invite_minted = int32(730805919) // 0x2b8f369f
)

func (m *TLInviteMint) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_invite_mint)
	x.String(m.Phone)
	return nil
}

func (m *TLInviteMint) Decode(dBuf *DecodeBuf) error {
	m.Phone = dBuf.String()
	return dBuf.err
}

func (m *Invite_Minted) Encode(x *EncodeBuf, layer int32) error {
	x.Int(CRC32_invite_minted)
	x.String(m.Code)
	x.Int(m.Expires)
	return nil
}

func (m *Invite_Minted) Decode(dBuf *DecodeBuf) error {
	m.Code = dBuf.String()
	m.Expires = dBuf.Int()
	return dBuf.err
}

// InviteConstructorNames is the table of our own invite constructors, for the
// tests that check each is registered and none collides.
func InviteConstructorNames() map[int32]string {
	return map[int32]string{
		CRC32_invite_mint:   "invite.mint",
		CRC32_invite_minted: "invite.minted",
	}
}

func init() {
	clazzIdRegisters2[CRC32_invite_mint] = func() TLObject { return &TLInviteMint{} }
	clazzIdRegisters2[CRC32_invite_minted] = func() TLObject { return &Invite_Minted{} }

	rpcContextRegisters["TLInviteMint"] = RPCContextTuple{
		"/mtproto.RPCInvite/invite_mint",
		func() interface{} { return new(Invite_Minted) },
	}
}
