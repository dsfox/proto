# The schema proto/mtproto is generated from

teamgram's `mtprotoc` is private and its input, `scheme.tl`, was never
published. `proto/mtproto`'s twenty-three generated files are made from this
directory instead, by `proto/ice9/mtprotoc` (#276), and
`tests/test_proto_generator_reproduces_the_tree.py` requires every one of
them to be the tree's, byte for byte.

## The files

- `layers.tl` - Telegram's API, laid out as teamgram's input is: a section per
  layer, newest first. The top section is the whole schema at that layer; a
  section `// LAYER 227 from 228` lists only what differs at 227 - a line that
  replaces the one above, or `- name` for an entity the layer above has and
  this one has not; the old layers 134 to 71 are a handful of methods each.
  Within a section, lines are in the order teamgram's input has them, which is
  the order of everything generated. Names are TL's (`help.getConfig`); a line
  without `#id` has the id TL defines, the crc32 of the line.
- `mtproto.tl` - the MTProto service types and methods, layer 0: what is served
  when a client's layer is not listed. From Telegram's `mtproto.tl`, plus the
  five base types teamgram also serves at layer 0.
- `teamgram.tl` - teamgram's own types and methods, layer 0: the predefined
  users, the message boxes, the raw business data, and the old `initConnection`
  they still serve.
- `services.txt` - the gRPC service each method is answered from, in the order
  of `schema.tl.sync_service.proto`. Telegram's schema has no such grouping; it
  is teamgram's, read from their output once.
- `extra_fields.txt` - eleven fields teamgram's generator writes after a
  message's own, where a name has several types over the layers: one of them
  again under the plain name, for the server code that reads it.
- `overrides.txt` - the lines whose id is not the crc32 of their text, with the
  reason: two predicates teamgram shortened, three ids Telegram's own
  `mtproto.tl` assigned, and the `initConnection` teamgram serves under the old
  id. The test refuses any other such line, and an entry no longer needed.

## Where it came from

Read back from the generated tree at layer 228 (`class_name_registers.go` for
the predicates, their order and their ids by layer; the id-to-name map for the
order within a layer; `rpc_client_registers.go` for the result types) with the
definitions from Telegram's per-layer schema files in the Android client
(`TMessagesProj_AppTests/tlscheme/N.json`), then finished by hand from the
codec where the two disagreed. Android's files carry a few flag bits the
official schema never had and lack two that `createStickerSet` dropped without a
new id; the sync protos said which.

Three things teamgram's generator does that no input says are kept as data or
as code with a comment: the two hand-styled codec blocks of `FutureSalts` and
`FutureSalt` (`emit/codec_handwritten.go`), the fallbacks in the `Message` and
`MessageMedia` codecs for a client whose layer lacks a media or an action
(`emit/codec_special.go`), and ice9's own `updateMlsMailbox` case
(`emit/ice9.go`, what `codegen.patch` used to hold).

## Adding a layer

    go run ./ice9/mtprotoc import -schema ice9/schema -tlscheme <path to N.json>
    ice9/mtprotoc/build.sh
    go test ./ice9/mtprotoc/... ./mtproto/

The import makes the new layer the top section in the order the schema file
has, moves what the previous layer had and the new one changes into a
`from` section, and lists every new method it put into `services.txt` under
the service of the method before it - read that list and move what belongs
elsewhere. A line whose id is not the crc32 of its text is printed too: a
typo in the schema file, or a line to add to `overrides.txt` with its reason.
