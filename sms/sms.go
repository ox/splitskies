package sms

import (
	"fmt"
	"splitskies/config"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
)

type TwilioSMSService struct {
	verifyServiceSid string
	client           *twilio.RestClient
}

func Connect(conf config.TwilioConfig) SMSVerifier {
	t := &TwilioSMSService{
		verifyServiceSid: conf.VerifyServicesSID,
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: conf.AccountSID,
			Password: conf.AuthToken,
		}),
	}

	return t
}

func (t *TwilioSMSService) Send(to string) error {
	params := &openapi.CreateVerificationParams{}
	params.SetTo(to)
	params.SetChannel("sms")

	_, err := t.client.VerifyV2.CreateVerification(t.verifyServiceSid, params)
	if err != nil {
		return fmt.Errorf("could not create verification: %w", err)
	}
	return nil
}

func (t *TwilioSMSService) Check(phoneNumber, code string) (bool, error) {
	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(phoneNumber)
	params.SetCode(code)

	resp, err := t.client.VerifyV2.CreateVerificationCheck(t.verifyServiceSid, params)
	if err != nil {
		return false, fmt.Errorf("could not create verification check: %w", err)
	}

	return *resp.Status == "approved", nil
}
