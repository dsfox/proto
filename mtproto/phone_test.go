package mtproto

import (
	"bytes"
	"testing"
)

// The ICE-candidate channel: phone.sendSignalingData is forwarded to the other
// phone as this update, and a direct path exists only if the bytes arrive.
// The generated codec was once misread as carrying the call id alone, so the
// payload is asserted through the container a phone actually receives.
func TestTheSignalingUpdateCarriesItsData(t *testing.T) {
	candidate := []byte("candidate:1 1 udp 2130706431 192.0.2.1 54321 typ host")

	update := MakeTLUpdatePhoneCallSignalingData(nil)
	update.SetPhoneCallId(0x1122334455667788)
	update.SetData_FLAGBYTES(candidate)

	original := MakeTLUpdates(&Updates{
		Updates: []*Update{update.To_Update()},
		Users:   []*User{},
		Chats:   []*Chat{},
		Date:    1,
		Seq:     0,
	}).To_Updates()

	buf := NewEncodeBuf(512)
	if err := original.Encode(buf, 228); err != nil {
		t.Fatalf("cannot encode the container: %v", err)
	}

	decoded := &Updates{}
	if err := decoded.Decode(NewDecodeBuf(buf.GetBuf())); err != nil {
		t.Fatalf("cannot decode the container: %v", err)
	}
	if len(decoded.Updates) != 1 {
		t.Fatalf("came back with %d updates, sent 1", len(decoded.Updates))
	}
	got := decoded.Updates[0]
	if uint32(got.Constructor) != 0x2661bf09 {
		t.Fatalf("came back under constructor %x", uint32(got.Constructor))
	}
	if got.PhoneCallId != 0x1122334455667788 {
		t.Errorf("call id came back as %x", got.PhoneCallId)
	}
	if !bytes.Equal(got.Data_FLAGBYTES, candidate) {
		t.Errorf("the candidate came back as %q", got.Data_FLAGBYTES)
	}
}
