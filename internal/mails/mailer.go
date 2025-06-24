package mails

type Client interface {
	Send(template, username, email string, data any, isSandbox bool) (int, error)
}
