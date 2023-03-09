package telegram

type FakeRecipient struct {
	ID string
}

func (f FakeRecipient) Recipient() string {
	return f.ID
}
