package mail

type Message struct {
	From     Address
	To       Address
	ReplyTo  ReplyTo
	Subject  Subject
	HTMLBody HTMLBody
	TextBody TextBody
}

func NewMessage(
	from Address,
	to Address,
	replyTo ReplyTo,
	subject Subject,
	htmlBody HTMLBody,
	textBody TextBody,
) Message {
	return Message{
		From:     from,
		To:       to,
		ReplyTo:  replyTo,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}
}
