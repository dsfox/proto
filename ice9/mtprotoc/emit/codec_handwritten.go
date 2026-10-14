package emit

// futureSaltsCodec is the codec of FutureSalts and FutureSalt as teamgram's
// generator writes them, by an older template than every other type's:
// the type's Encode returns the constructor's directly, the comments carry
// the TL line, and a bare vector is written without the sz/offset bookkeeping.
// They are the first two blocks of codec_schema.tl.pb.go, kept verbatim.
const futureSaltsCodec = `///////////////////////////////////////////////////////////////////////////////
// FutureSalts <--
//  + TL_FutureSalts
//

func (m *FutureSalts) Encode(x *EncodeBuf, layer int32) error {
	predicateName := m.PredicateName
	if predicateName == "" {
		if n, ok := clazzIdNameRegisters2[int32(m.Constructor)]; ok {
			predicateName = n
		}
	}
	switch predicateName {
	case Predicate_future_salts:
		t := m.To_FutureSalts()

		return t.Encode(x, layer)
	default:
		// err := fmt.Errorf("invalid predicate error: %s", m.PredicateName)
		// log.Errorf("invalid predicate(%s, %d): %s", m.PredicateName, layer)

		return nil
	}
}

func (m *FutureSalts) CalcByteSize(layer int32) int {
	return 0
}

func (m *FutureSalts) Decode(dBuf *DecodeBuf) error {
	m.Constructor = TLConstructor(dBuf.Int())
	switch uint32(m.Constructor) {
	case 0xae500895:
		m2 := MakeTLFutureSalts(m)
		m2.Decode(dBuf)

	default:
		return fmt.Errorf("invalid constructorId: 0x%x", uint32(m.Constructor))
	}
	return dBuf.GetError()
}

// To_FutureSalts
// future_salts#ae500895 req_msg_id:long now:int salts:vector<future_salt> = FutureSalts;
func (m *FutureSalts) To_FutureSalts() *TLFutureSalts {
	m.PredicateName = Predicate_future_salts
	return &TLFutureSalts{
		Data2: m,
	}
}

// MakeTLFutureSalts
// future_salts#ae500895 req_msg_id:long now:int salts:vector<future_salt> = FutureSalts;
func MakeTLFutureSalts(data2 *FutureSalts) *TLFutureSalts {
	if data2 == nil {
		return &TLFutureSalts{Data2: &FutureSalts{
			PredicateName: Predicate_future_salts,
		}}
	} else {
		data2.PredicateName = Predicate_future_salts
		return &TLFutureSalts{Data2: data2}
	}
}

func (m *TLFutureSalts) To_FutureSalts() *FutureSalts {
	m.Data2.PredicateName = Predicate_future_salts
	return m.Data2
}

func (m *TLFutureSalts) SetReqMsgId(v int64) { m.Data2.ReqMsgId = v }
func (m *TLFutureSalts) GetReqMsgId() int64  { return m.Data2.ReqMsgId }

func (m *TLFutureSalts) SetNow(v int32) { m.Data2.Now = v }
func (m *TLFutureSalts) GetNow() int32  { return m.Data2.Now }

func (m *TLFutureSalts) SetSalts(v []*TLFutureSalt) { m.Data2.Salts = v }
func (m *TLFutureSalts) GetSalts() []*TLFutureSalt  { return m.Data2.Salts }

func (m *TLFutureSalts) GetPredicateName() string {
	return Predicate_future_salts
}

func (m *TLFutureSalts) Encode(x *EncodeBuf, layer int32) error {
	clazzId := GetClazzID(Predicate_future_salts, int(layer))
	switch uint32(clazzId) {
	case 0xae500895:
		// future_salts#ae500895 req_msg_id:long now:int salts:vector<future_salt> = FutureSalts;
		x.UInt(0xae500895)
		// no flags

		x.Long(m.GetReqMsgId())
		x.Int(m.GetNow())
		// x.Int(int32(CRC32_vector))
		salts := m.GetSalts()
		x.Int(int32(len(salts)))
		for _, v := range salts {
			if err := v.Encode(x, layer); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("not found clazzId by (%s, %d)", Predicate_future_salts, layer)
	}
}

func (m *TLFutureSalts) CalcByteSize(layer int32) int {
	return 0
}

func (m *TLFutureSalts) Decode(dBuf *DecodeBuf) error {
	switch uint32(m.Data2.Constructor) {
	case 0xae500895:
		// future_salts#ae500895 req_msg_id:long now:int salts:vector<future_salt> = FutureSalts;
		// not has flags

		m.SetReqMsgId(dBuf.Long())
		m.SetNow(dBuf.Int())
		// c2 := dBuf.Int()
		// if c2 != int32(CRC32_vector) {
		//     dBuf.err = fmt.Errorf("invalid CRC32_vector, c%d: %d", 2, c2)
		//     return dBuf.err
		// }
		l2 := dBuf.Int()
		if l2 < 0 || l2 > maxVectorLen {
			return fmt.Errorf("invalid vector length: %d", l2)
		}
		v2 := make([]*TLFutureSalt, l2)
		for i := int32(0); i < l2; i++ {
			v2[i] = &TLFutureSalt{}
			v2[i].Decode(dBuf)
		}
		m.SetSalts(v2)

		return dBuf.GetError()
	default:
		return fmt.Errorf("invalid constructor: %x", uint32(m.Data2.Constructor))
	}
}

///////////////////////////////////////////////////////////////////////////////
// FutureSalt <--
//  + TL_FutureSalt
//

func (m *FutureSalt) Encode(x *EncodeBuf, layer int32) error {
	predicateName := m.PredicateName
	if predicateName == "" {
		if n, ok := clazzIdNameRegisters2[int32(m.Constructor)]; ok {
			predicateName = n
		}
	}
	switch predicateName {
	case Predicate_future_salt:
		t := m.To_FutureSalt()
		return t.Encode(x, layer)

	default:
		// err := fmt.Errorf("invalid predicate error: %s", m.PredicateName)
		// log.Errorf(err.Error())
		return nil
	}
}

func (m *FutureSalt) CalcByteSize(layer int32) int {
	return 0
}

func (m *FutureSalt) Decode(dBuf *DecodeBuf) error {
	m.Constructor = TLConstructor(dBuf.Int())
	switch uint32(m.Constructor) {
	case 0x949d9dc:
		m2 := MakeTLFutureSalt(m)
		m2.Decode(dBuf)

	default:
		return fmt.Errorf("invalid constructorId: 0x%x", uint32(m.Constructor))
	}
	return dBuf.GetError()
}

// To_FutureSalt
// future_salt#0949d9dc valid_since:int valid_until:int salt:long = FutureSalt;
func (m *FutureSalt) To_FutureSalt() *TLFutureSalt {
	m.PredicateName = Predicate_future_salt
	return &TLFutureSalt{
		Data2: m,
	}
}

// MakeTLFutureSalt
// future_salt#0949d9dc valid_since:int valid_until:int salt:long = FutureSalt;
func MakeTLFutureSalt(data2 *FutureSalt) *TLFutureSalt {
	if data2 == nil {
		return &TLFutureSalt{Data2: &FutureSalt{
			PredicateName: Predicate_future_salt,
		}}
	} else {
		data2.PredicateName = Predicate_future_salt
		return &TLFutureSalt{Data2: data2}
	}
}

func (m *TLFutureSalt) To_FutureSalt() *FutureSalt {
	m.Data2.PredicateName = Predicate_future_salt
	return m.Data2
}

func (m *TLFutureSalt) SetValidSince(v int32) { m.Data2.ValidSince = v }
func (m *TLFutureSalt) GetValidSince() int32  { return m.Data2.ValidSince }

func (m *TLFutureSalt) SetValidUntil(v int32) { m.Data2.ValidUntil = v }
func (m *TLFutureSalt) GetValidUntil() int32  { return m.Data2.ValidUntil }

func (m *TLFutureSalt) SetSalt(v int64) { m.Data2.Salt = v }
func (m *TLFutureSalt) GetSalt() int64  { return m.Data2.Salt }

func (m *TLFutureSalt) GetPredicateName() string {
	return Predicate_future_salt
}

func (m *TLFutureSalt) Encode(x *EncodeBuf, layer int32) error {
	clazzId := GetClazzID(Predicate_future_salt, int(layer))
	switch uint32(clazzId) {
	case 0x949d9dc:
		// future_salt#0949d9dc valid_since:int valid_until:int salt:long = FutureSalt;
		// FIXME(@benqi)
		// x.UInt(0x949d9dc)
		// no flags

		x.Int(m.GetValidSince())
		x.Int(m.GetValidUntil())
		x.Long(m.GetSalt())
		return nil
	default:
		return fmt.Errorf("not found clazzId by (%s, %d)", Predicate_future_salt, layer)
	}
}

func (m *TLFutureSalt) CalcByteSize(layer int32) int {
	return 0
}

func (m *TLFutureSalt) Decode(dBuf *DecodeBuf) error {
	switch uint32(m.Data2.Constructor) {
	case 0x949d9dc:
		// future_salt#0949d9dc valid_since:int valid_until:int salt:long = FutureSalt;
		// not has flags

		m.SetValidSince(dBuf.Int())
		m.SetValidUntil(dBuf.Int())
		m.SetSalt(dBuf.Long())
		return dBuf.GetError()
	default:
		return fmt.Errorf("invalid constructor: %x", uint32(m.Data2.Constructor))
	}
}
`
