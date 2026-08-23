package mtproto

// What a client one layer ahead of the schema used to be answered with.
//
// Every Encode in the generated schema starts by asking GetClazzID for the
// constructor to write at the client's layer, and switches on the answer. The
// registry is generated up to a fixed layer - 228 here - and a client that
// announces a newer one got 0 back, which matches no case in any switch: the
// encoder wrote nothing, and the server answered with an rpc_result carrying an
// empty body. Nothing in the logs said so. The handler ran, returned its
// object, and the bytes were dropped on the way out; the client saw a reply it
// could not parse and, in a debug build, died on it.
//
// Telegram Android 12.10 announces layer 229, so this was not a corner: it was
// every reply to every request from the carried client.
//
// A constructor listed at some layer holds until a later layer changes it, so
// for a client ahead of the schema the newest entry we have is the honest
// answer - it is what "this server speaks layer 228" means on the wire. A layer
// *older* than anything listed for a predicate is a different thing entirely -
// the predicate did not exist then - and is still refused.

type clazzTop struct {
	layer int
	id    int32
}

// The newest layer each predicate is listed at, so the answer above costs a
// lookup rather than a scan of the predicate's layers on every object encoded.
var clazzNameNewest = func() map[string]clazzTop {
	newest := make(map[string]clazzTop, len(clazzNameRegisters2))
	for name, byLayer := range clazzNameRegisters2 {
		top := clazzTop{}
		for layer, id := range byLayer {
			if layer > top.layer {
				top = clazzTop{layer: layer, id: id}
			}
		}
		if top.layer > 0 {
			newest[name] = top
		}
	}
	return newest
}()
