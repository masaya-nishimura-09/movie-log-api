package mail

type ReplyTo string

func NewReplyTo(value string) (ReplyTo, error) {
	if value == "" {
		return "", nil
	}
	address, err := NewAddress(value)
	if err != nil {
		return "", err
	}
	return ReplyTo(address), nil
}
