package mtproto

import (
	"bytes"
	"testing"
)

// Hand-written TL is exactly the kind of code that looks right and puts the
// bytes in the wrong order, so every type goes out and comes back.
func TestOurMethodsSurviveTheirOwnEncoding(t *testing.T) {
	t.Run("publishKeyPackages", func(t *testing.T) {
		original := &TLMlsPublishKeyPackages{
			KeyPackages: [][]byte{{1, 2, 3}, {4, 5}, {}},
			LastResort:  []byte{9, 9, 9, 9},
		}

		var decoded TLMlsPublishKeyPackages
		roundTrip(t, original, &decoded)

		if len(decoded.KeyPackages) != 3 {
			t.Fatalf("came back with %d packages, sent 3", len(decoded.KeyPackages))
		}
		if !bytes.Equal(decoded.KeyPackages[0], []byte{1, 2, 3}) {
			t.Errorf("the first package came back as %v", decoded.KeyPackages[0])
		}
		if !bytes.Equal(decoded.LastResort, []byte{9, 9, 9, 9}) {
			t.Errorf("the last-resort package came back as %v", decoded.LastResort)
		}
	})

	t.Run("publishResult", func(t *testing.T) {
		original := &Mls_PublishResult{Added: 7, Available: 42, ShouldRefill: true}

		var decoded Mls_PublishResult
		roundTrip(t, original, &decoded)

		if decoded.Added != 7 || decoded.Available != 42 || !decoded.ShouldRefill {
			t.Fatalf("came back as %s", decoded.String())
		}
	})

	t.Run("publishResult without refill", func(t *testing.T) {
		// The false case is worth its own test: a boolean encoded as a
		// constructor is exactly where a wrong constant reads as true forever.
		original := &Mls_PublishResult{Added: 0, Available: 100, ShouldRefill: false}

		var decoded Mls_PublishResult
		roundTrip(t, original, &decoded)

		if decoded.ShouldRefill {
			t.Fatal("a device that needs nothing was told to publish more")
		}
	})

	t.Run("claimKeyPackages", func(t *testing.T) {
		original := &TLMlsClaimKeyPackages{UserId: -6067108358985120146}

		var decoded TLMlsClaimKeyPackages
		roundTrip(t, original, &decoded)

		if decoded.UserId != original.UserId {
			t.Fatalf("the user id came back as %d", decoded.UserId)
		}
	})

	t.Run("sendWelcome", func(t *testing.T) {
		original := &TLMlsSendWelcome{UserId: 136907759, Welcome: []byte{1, 2, 3, 4, 5}}

		var decoded TLMlsSendWelcome
		roundTrip(t, original, &decoded)

		if decoded.UserId != original.UserId || !bytes.Equal(decoded.Welcome, original.Welcome) {
			t.Fatalf("came back as %s", decoded.String())
		}
	})

	t.Run("welcomes", func(t *testing.T) {
		original := &Mls_Welcomes{Welcomes: []*Mls_Welcome{
			{Id: 1, FromId: 7, Welcome: []byte{0xAA}},
			{Id: 2, FromId: 8, Welcome: []byte{0xBB, 0xCC}},
		}}

		var decoded Mls_Welcomes
		roundTrip(t, original, &decoded)

		if len(decoded.Welcomes) != 2 {
			t.Fatalf("came back with %d welcomes", len(decoded.Welcomes))
		}
		if decoded.Welcomes[1].FromId != 8 || !bytes.Equal(decoded.Welcomes[1].Welcome, []byte{0xBB, 0xCC}) {
			t.Fatalf("the second one came back as %s", decoded.Welcomes[1].String())
		}
	})

	t.Run("confirmWelcomes", func(t *testing.T) {
		original := &TLMlsConfirmWelcomes{Ids: []int64{5, 6, 7}}

		var decoded TLMlsConfirmWelcomes
		roundTrip(t, original, &decoded)

		if len(decoded.Ids) != 3 || decoded.Ids[2] != 7 {
			t.Fatalf("came back as %v", decoded.Ids)
		}
	})

	t.Run("keyPackages", func(t *testing.T) {
		original := &Mls_KeyPackages{Packages: [][]byte{{0xAA}, {0xBB, 0xCC}}}

		var decoded Mls_KeyPackages
		roundTrip(t, original, &decoded)

		if len(decoded.Packages) != 2 || !bytes.Equal(decoded.Packages[1], []byte{0xBB, 0xCC}) {
			t.Fatalf("came back as %s", decoded.String())
		}
	})
}

// The server builds incoming objects from this table, so a type that is not in
// it arrives as nothing at all - silently.
func TestACommitSurvivesItsOwnEncoding(t *testing.T) {
	t.Run("sendCommit", func(t *testing.T) {
		original := &TLMlsSendCommit{
			GroupId: []byte{7, 7, 7},
			Epoch:   42,
			Members: []int64{11, 22, 33},
			Commit:  []byte{1, 2, 3, 4},
		}

		var decoded TLMlsSendCommit
		roundTrip(t, original, &decoded)

		if !bytes.Equal(decoded.GroupId, []byte{7, 7, 7}) {
			t.Errorf("the group came back as %v", decoded.GroupId)
		}
		if decoded.Epoch != 42 {
			t.Errorf("the epoch came back as %d", decoded.Epoch)
		}
		if len(decoded.Members) != 3 || decoded.Members[2] != 33 {
			t.Errorf("the members came back as %v", decoded.Members)
		}
		if !bytes.Equal(decoded.Commit, []byte{1, 2, 3, 4}) {
			t.Errorf("the commit came back as %v", decoded.Commit)
		}
	})

	// The epoch has to survive a refusal as well as an acceptance: it is what
	// tells the loser of a race how far behind it is, and a zero there would
	// send it back with the same commit for ever.
	t.Run("commitResult, refused", func(t *testing.T) {
		original := &Mls_CommitResult{Accepted: false, Epoch: 9}

		var decoded Mls_CommitResult
		roundTrip(t, original, &decoded)

		if decoded.Accepted {
			t.Error("a refusal came back as an acceptance")
		}
		if decoded.Epoch != 9 {
			t.Errorf("the epoch came back as %d", decoded.Epoch)
		}
	})

	t.Run("commits", func(t *testing.T) {
		original := &Mls_Commits{Commits: []*Mls_Commit{
			{Id: 1, FromId: 2, GroupId: []byte{3}, Epoch: 4, Commit: []byte{5}},
			{Id: 6, FromId: 7, GroupId: []byte{8}, Epoch: 9, Commit: []byte{10}},
		}}

		var decoded Mls_Commits
		roundTrip(t, original, &decoded)

		if len(decoded.Commits) != 2 {
			t.Fatalf("came back with %d commits, sent 2", len(decoded.Commits))
		}
		if decoded.Commits[1].Epoch != 9 || decoded.Commits[1].Id != 6 {
			t.Errorf("the second came back as %s", decoded.Commits[1].String())
		}
	})
}

func TestOurMethodsAreInTheTableTheServerReadsFrom(t *testing.T) {
	for id, name := range MlsConstructorNames() {
		object := NewTLObjectByClassID(id)
		if object == nil {
			t.Errorf("%s (%d) is not registered, so the server would never see it", name, id)
		}
	}
}

// A constructor id that collides with an existing one turns two different
// messages into the same message, and the failure is far from the cause.
func TestOurConstructorIdsAreOurs(t *testing.T) {
	ours := MlsConstructorNames()
	if len(ours) == 0 {
		t.Fatal("the table of our own constructors is empty, so this checks nothing")
	}

	seen := map[int32]bool{}
	for id, name := range ours {
		if seen[id] {
			t.Errorf("%s reuses an id we already took", name)
		}
		seen[id] = true
	}
}

// The update is not a method: it goes out inside an Updates container, which
// dispatches by predicate on the way out and by constructor on the way back.
// Both dispatches are switches in generated code with a hand-written case
// spliced in, and a case that is missing from either fails only at runtime,
// as a push the client never receives.
func TestTheMailboxUpdateSurvivesTheContainer(t *testing.T) {
	original := MakeTLUpdates(&Updates{
		Updates: []*Update{MakeTLUpdateMlsMailbox(nil).To_Update()},
		Users:   []*User{},
		Chats:   []*Chat{},
		Date:    1,
		Seq:     0,
	}).To_Updates()

	// At the layer a phone actually speaks: the container is generated code
	// and knows no layer 0, while the update inside it must be found at any.
	buf := NewEncodeBuf(512)
	if err := original.Encode(buf, 228); err != nil {
		t.Fatalf("cannot encode the container: %v", err)
	}

	dBuf := NewDecodeBuf(buf.GetBuf())
	decoded := &Updates{}
	if err := decoded.Decode(dBuf); err != nil {
		t.Fatalf("cannot decode the container: %v", err)
	}
	if len(decoded.Updates) != 1 {
		t.Fatalf("came back with %d updates, sent 1", len(decoded.Updates))
	}
	if decoded.Updates[0].PredicateName != Predicate_updateMlsMailbox {
		t.Fatalf("came back as %q", decoded.Updates[0].PredicateName)
	}
	if int32(decoded.Updates[0].Constructor) != CRC32_updateMlsMailbox {
		t.Fatalf("came back under constructor %x", uint32(decoded.Updates[0].Constructor))
	}
}

func roundTrip(t *testing.T, original, decoded TLObject) {
	t.Helper()

	buf := NewEncodeBuf(512)
	if err := original.Encode(buf, 0); err != nil {
		t.Fatalf("cannot encode: %v", err)
	}

	raw := buf.GetBuf()
	dBuf := NewDecodeBuf(raw)
	if id := dBuf.Int(); id == 0 {
		t.Fatal("nothing was written")
	}
	if err := decoded.Decode(dBuf); err != nil {
		t.Fatalf("cannot decode: %v", err)
	}
}

// A request with nowhere to go is answered "not found method", which names the
// symptom and not the missing line - so the route is asserted rather than
// assumed.
func TestOurMethodsKnowWhereToGo(t *testing.T) {
	for _, name := range []string{"TLMlsPublishKeyPackages", "TLMlsClaimKeyPackages"} {
		tuple, ok := GetRPCContextRegisters()[name]
		if !ok {
			t.Errorf("%s has no route, so the proxy would refuse it", name)
			continue
		}
		if tuple.Method == "" || tuple.NewReplyFunc == nil {
			t.Errorf("%s is routed to %q with reply %v", name, tuple.Method, tuple.NewReplyFunc)
		}
	}
}

// Each id must be the CRC32 of the declaration written above it. That is how TL
// makes them, and it is what lets anybody recompute them from the text rather
// than trust a number.
//
// The declarations these ids are the CRC32 of are held in one place:
// tests/test_mls_constructors.py in the outer repository. There used to be a
// second copy here, and it did what second copies do - it went stale and stayed
// green-looking, because nothing ran it. By the time anybody looked it was four
// constructors behind and had the wrong text for a fifth, so the one check that
// was supposed to catch a number drifting had itself drifted.
//
// The outer gate reads the numbers straight out of this file and recomputes
// them from the text, and it does the same for both clients - which is the
// whole point, since a number that agrees with itself here and with nothing
// else is exactly the failure that matters.
