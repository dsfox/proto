package mtproto

import (
	"bytes"
	"hash/crc32"
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
	if len(ours) != 17 {
		t.Fatalf("expected seventeen of our own, found %d", len(ours))
	}

	seen := map[int32]bool{}
	for id, name := range ours {
		if seen[id] {
			t.Errorf("%s reuses an id we already took", name)
		}
		seen[id] = true
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
// Written after getting it wrong: two of the four were converted to signed by
// hand and came out different from their own comments. The server stayed
// consistent with itself, so everything worked and nothing matched the text -
// which is worse than a failure, because it survives.
func TestOurIdsAreTheCrc32OfTheirDeclarations(t *testing.T) {
	declarations := map[int32]string{
		CRC32_mls_publishKeyPackages: "mls.publishKeyPackages key_packages:Vector<bytes> last_resort:bytes = mls.PublishResult;",
		CRC32_mls_publishResult:      "mls.publishResult added:int available:int should_refill:Bool = mls.PublishResult;",
		CRC32_mls_claimKeyPackages:   "mls.claimKeyPackages user_id:long = mls.KeyPackages;",
		CRC32_mls_keyPackages:        "mls.keyPackages packages:Vector<bytes> = mls.KeyPackages;",
		CRC32_mls_sendWelcome:        "mls.sendWelcome user_id:long welcome:bytes = mls.Ok;",
		CRC32_mls_ok:                 "mls.ok ok:Bool = mls.Ok;",
		CRC32_mls_getWelcomes:        "mls.getWelcomes = mls.Welcomes;",
		CRC32_mls_welcomes:           "mls.welcomes welcomes:Vector<mls.Welcome> = mls.Welcomes;",
		CRC32_mls_welcome:            "mls.welcome id:long from_id:long welcome:bytes = mls.Welcome;",
		CRC32_mls_confirmWelcomes:    "mls.confirmWelcomes ids:Vector<long> = mls.Ok;",
		CRC32_mls_sendCommit:         "mls.sendCommit group_id:bytes epoch:long members:Vector<long> commit:bytes = mls.CommitResult;",
		CRC32_mls_commitResult:       "mls.commitResult accepted:Bool epoch:long = mls.CommitResult;",
		CRC32_mls_getCommits:         "mls.getCommits = mls.Commits;",
		CRC32_mls_commits:            "mls.commits commits:Vector<mls.Commit> = mls.Commits;",
		CRC32_mls_commit:             "mls.commit id:long from_id:long group_id:bytes epoch:long commit:bytes = mls.Commit;",
		CRC32_mls_confirmCommits:     "mls.confirmCommits ids:Vector<long> = mls.Ok;",
	}

	for id, declaration := range declarations {
		want := int32(crc32.ChecksumIEEE([]byte(declaration)))
		if id != want {
			t.Errorf("%q\n  is 0x%08x by CRC32 and 0x%08x in the code",
				declaration, uint32(want), uint32(id))
		}
	}
}
