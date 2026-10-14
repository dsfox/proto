package emit

// teamgram's generator writes two types' Encode with fallbacks of its own,
// for a client whose layer has no id for a media or an action: Message
// sends such a media as messageMediaUnsupported and such a service message
// as a message, and MessageMedia tries messageMediaUnsupported when nothing
// was written. They are kept as teamgram wrote them.

// codecCaseOverride is a case body of a type's Encode, by type and predicate.
var codecCaseOverride = map[string]map[string]string{
	"Message": {
		"message": `		if m.GetMedia() != nil && GetClazzID(m.GetMedia().GetPredicateName(), int(layer)) == 0 {
			m.Media.PredicateName = Predicate_messageMediaUnsupported
		}

		t := m.To_Message()
		err := t.Encode(x, layer)
		if err != nil {
			return err
		}
`,
		"messageService": `		if m.GetAction() != nil && GetClazzID(m.GetAction().GetPredicateName(), int(layer)) == 0 {
			m.PredicateName = Predicate_message
			m.Media = MakeTLMessageMediaUnsupported(nil).To_MessageMedia()

			t := m.To_Message()
			err := t.Encode(x, layer)
			if err != nil {
				return err
			}
		} else {
			t := m.To_MessageService()
			err := t.Encode(x, layer)
			if err != nil {
				return err
			}
		}
`,
	},
}

// codecEncodePrologue goes between the predicate lookup and the switch.
var codecEncodePrologue = map[string]string{
	"MessageMedia": `	line100 := false
LINE100:

`,
}

// codecEncodeEpilogue goes after the switch, before "return nil".
var codecEncodeEpilogue = map[string]string{
	"MessageMedia": `	if line100 {
		return nil
	}

	if x.GetOffset() == 0 {
		predicateName = Predicate_messageMediaUnsupported
		line100 = true
		goto LINE100
	}

`,
}
