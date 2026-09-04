package mtproto

import (
	"os"
	"strings"
	"testing"
)

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

func TestInviteMintForChatSurvivesItsOwnEncoding(t *testing.T) {
	original := &TLInviteMintForChat{ChatId: 120090, Phone: "+79991234567"}
	var decoded TLInviteMintForChat
	roundTrip(t, original, &decoded)
	if decoded.ChatId != 120090 || decoded.Phone != "+79991234567" {
		t.Fatalf("came back as chat %d, number %q", decoded.ChatId, decoded.Phone)
	}
}

func TestInviteMintForChatKnowsWhereToGo(t *testing.T) {
	tuple, ok := GetRPCContextRegisters()["TLInviteMintForChat"]
	if !ok {
		t.Fatal("TLInviteMintForChat has no route, so the proxy would refuse it")
	}
	if tuple.Method != "/mtproto.RPCInvite/invite_mintForChat" || tuple.NewReplyFunc == nil {
		t.Fatalf("routed to %q with reply %v", tuple.Method, tuple.NewReplyFunc)
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

// A route with no client behind it is not a route: the session looks the
// method's gRPC prefix up in its own config, and a prefix missing there sends
// the request to the stub layer, which refuses it with a licence message
// that explains nothing. That is how invite.mint failed on its first run
// (#47), and mls.* would fail the same way the day somebody trims the file.
func TestOurRoutePrefixesHaveAClientInTheSessionConfig(t *testing.T) {
	config, err := os.ReadFile("../../server/teamgramd/etc2/session.yaml")
	if err != nil {
		t.Skipf("the session config is not beside this module: %v", err)
	}
	for _, tuple := range []RPCContextTuple{
		rpcContextRegisters["TLInviteMint"],
		rpcContextRegisters["TLMlsPublishKeyPackages"],
	} {
		prefix := tuple.Method[:strings.LastIndex(tuple.Method, "/")]
		if !strings.Contains(string(config), `"`+prefix+`"`) {
			t.Errorf("%s is routed to %s, and session.yaml names no client for %s", tuple.Method, tuple.Method, prefix)
		}
	}
}
