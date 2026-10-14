// Copyright 2024 Teamgram Authors
//  All rights reserved.
//
// Author: Benqi (wubenqi@gmail.com)
//

package mtproto

func (m *ReplyMarkup) ToForwardMessage() (replyMarkup *ReplyMarkup) {
	if m.GetPredicateName() != Predicate_replyInlineMarkup {
		return
	}

	// ice9: the inline rows of layer 229 on, and the rows an older client sent (#276)
	for _, row := range m.GetRows_VECTORKEYBOARDINLINEBUTTONROW() {
		for _, button := range row.GetButtons() {
			if button.GetType().GetPredicateName() != Predicate_inlineButtonTypeUrl {
				return
			}
		}
	}
	for _, row := range m.GetRows_VECTORKEYBOARDBUTTONROW() {
		for _, button := range row.GetButtons() {
			if button.GetPredicateName() != Predicate_keyboardButtonUrl {
				return
			}
		}
	}

	replyMarkup = m
	return
}
