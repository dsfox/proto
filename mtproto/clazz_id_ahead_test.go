package mtproto

import "testing"

// The layer Telegram Android 12.10 announces - TLRPC.java, LAYER. The schema
// here is generated up to 228, so this is the number that used to empty every
// reply the server sent.
const layerOfTheCarriedClient = 229

func TestAClientAheadOfTheSchemaIsServedWhatWeHave(t *testing.T) {
	for _, layer := range []int{layerOfTheCarriedClient, 260} {
		if id := uint32(GetClazzID(Predicate_config, layer)); id != 0xcc1a241e {
			t.Errorf("config at layer %d: 0x%x, want the layer 228 constructor 0xcc1a241e", layer, id)
		}
		if id := uint32(GetClazzID(Predicate_user, layer)); id != 0xb1b8cc83 {
			t.Errorf("user at layer %d: 0x%x, want the layer 228 constructor 0xb1b8cc83", layer, id)
		}
	}
}

func TestALayerTheSchemaKnowsIsUnchanged(t *testing.T) {
	for layer := 200; layer <= 228; layer++ {
		if id := uint32(GetClazzID(Predicate_config, layer)); id != 0xcc1a241e {
			t.Errorf("config at layer %d: 0x%x, want 0xcc1a241e", layer, id)
		}
	}
}

// The gate. Answering 0 makes the encoder write an empty body and say nothing
// about it, so no predicate may answer 0 at a layer above the newest generated
// one - which is what every future client will ask for.
func TestNoPredicateEncodesToNothingAboveTheNewestLayer(t *testing.T) {
	for name := range clazzNameRegisters2 {
		for _, layer := range []int{layerOfTheCarriedClient, 260} {
			if GetClazzID(name, layer) == 0 {
				t.Fatalf("%s at layer %d has no constructor, so it would be encoded as no bytes at all", name, layer)
			}
		}
	}
}

// The other direction is not a client ahead of us but a predicate that did not
// exist yet, and inventing a constructor for it would be a guess.
func TestALayerBeforeAPredicateExistedIsStillRefused(t *testing.T) {
	checked := 0
	for name, byLayer := range clazzNameRegisters2 {
		if _, always := byLayer[0]; always {
			continue
		}
		oldest := 1 << 30
		for layer := range byLayer {
			if layer < oldest {
				oldest = layer
			}
		}
		if oldest < 2 {
			continue
		}
		if id := GetClazzID(name, oldest-1); id != 0 {
			t.Fatalf("%s answered 0x%x at layer %d, one below the %d it first appears at", name, uint32(id), oldest-1, oldest)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no predicate with a first layer was found, so this checked nothing")
	}
}
