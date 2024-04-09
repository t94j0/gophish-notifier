package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	errNoSignature      = errors.New("header X-Gophish-Signature not provided")
	errInvalidSignature = errors.New("invalid signature")
)

func validateSignature(body []byte, r *http.Request) error {
	secret := viper.GetString("secret")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	var actual string
	if _, err := fmt.Sscanf(r.Header.Get("X-Gophish-Signature"), "sha256=%s", &actual); err != nil {
		return errNoSignature
	}
	if !hmac.Equal([]byte(expected), []byte(actual)) {
		return errInvalidSignature
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debug(string(body))

	if err := validateSignature(body, r); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := NewWebhookResponse(body)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if response.Success {
		return
	}

	sender, err := senderDispatch(response.Message, response, []byte(response.Details))
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profiles := viper.GetStringSlice("profiles")
	for _, profile := range profiles {
		if profile == "email" {
			if err := sender.SendEmail(); err != nil {
				log.Error(err)
				return
			}
		}
		if profile == "slack" {
			if err := sender.SendSlack(); err != nil {
				log.Error(err)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	addr := net.JoinHostPort(viper.GetString("listen_host"), viper.GetString("listen_port"))
	log.Infof("Server listening on %s%s", addr, viper.GetString("webhook_path"))
	http.ListenAndServe(addr, http.HandlerFunc(handler))
}
