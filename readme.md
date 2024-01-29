# Splitskies

## Running

Requires [`reflex`](https://github.com/cespare/reflex). Copy the `env/example.yaml` file and name it `env/dev.yaml`. Add in your Twilio Auth tokens and Verify Service SID if you want to try out the login flow.

```
% reflex -v -c reflex.conf
```

This should start a new service on :7707.

If you don't want to add Twilio tokens, then run the service, visit it, and change your cookie's SessionID value to `adminsession`; this session is hard-coded in `app.go`. Using this session you can skip the verification process handled by Twilio and get straight to creating a Username and using the app.