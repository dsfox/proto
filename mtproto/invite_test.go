package mtproto

import "testing"

// Hand-written TL is exactly the kind of code that looks right and puts the
// bytes in the wrong order, so both types go out and come back.
func TestInviteMintSurvivesItsOwnEncoding(t *testing.T) {
	original := &TLInviteMint{Phone: "+79991234567"}
	var decoded TLInviteMint
	roundTrip(t, original, &decoded)
	if decoded.Phone != "+79991234567" {
		t.Fatalf("the number came back as %q", decoded.Phone)
	}

	answer := &Invite_Minted{Code: "123456", Expires: 1788000000}
	var back Invite_Minted
	roundTrip(t, answer, &back)
	if back.Code != "123456" || back.Expires != 1788000000 {
		t.Fatalf("the answer came back as %s", back.String())
	}
}

func TestInviteIsInTheTableTheServerReadsFrom(t *testing.T) {
	for id, name := range InviteConstructorNames() {
		if NewTLObjectByClassID(id) == nil {
			t.Errorf("%s (%d) is not registered, so the server would never see it", name, id)
		}
	}
}

func TestInviteMintKnowsWhereToGo(t *testing.T) {
	tuple, ok := GetRPCContextRegisters()["TLInviteMint"]
	if !ok {
		t.Fatal("TLInviteMint has no route, so the proxy would refuse it")
	}
	if tuple.Method != "/mtproto.RPCInvite/invite_mint" || tuple.NewReplyFunc == nil {
		t.Fatalf("TLInviteMint is routed to %q with reply %v", tuple.Method, tuple.NewReplyFunc)
	}
}
