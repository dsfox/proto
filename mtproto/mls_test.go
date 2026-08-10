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
	if len(ours) != 4 {
		t.Fatalf("expected four of our own, found %d", len(ours))
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
