package telegram

import "strconv"

type FakeRecipient struct {
	ID string
}

func (f FakeRecipient) Recipient() string {
	return f.ID
}

func RecipientFromInt64(chatId int64) FakeRecipient {
	return FakeRecipient{ID: strconv.FormatInt(chatId, 10)}
}
