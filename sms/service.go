package sms

type SMSVerifier interface {
	Send(phoneNumber string) error
	Check(phoneNumber, code string) (bool, error)
}
