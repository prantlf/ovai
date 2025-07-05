package auth

import (
	"encoding/json"
	"os"
	"time"

	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/web"
	"github.com/tidwall/jsonc"
)

type request struct {
	GrantType string `json:"grant_type"`
	Assertion string `json:"assertion"`
}

type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

var Account = readAccount()

var accessToken string
var accessExpires time.Time

func readAccount() *account {
	accntFile := os.Getenv("OVAI_ACCOUNT")
	if len(accntFile) == 0 {
		accntFile = "google-account.json"
	}
	log.Dbg("read %s", accntFile)
	accntJson, err := os.ReadFile(accntFile)
	if err != nil {
		log.Ftl("reading %s failed: %v", accntFile, err)
	}
	var accnt account
	jsonc.ToJSONInPlace(accntJson)
	if err := json.Unmarshal(accntJson, &accnt); err != nil {
		log.Ftl("decoding %s failed: %v", accntFile, err)
	}
	if len(accnt.Scope) == 0 {
		accnt.Scope = "https://www.googleapis.com/auth/cloud-platform"
	}
	if len(accnt.AuthUri) == 0 {
		accnt.AuthUri = "https://www.googleapis.com/oauth2/v4/token"
	}
	return &accnt
}

func RefreshAccessToken() (string, error) {
	signedToken, err := createToken()
	if err != nil {
		return "", err
	}
	reqJson := &request{
		GrantType: "urn:ietf:params:oauth:grant-type:jwt-bearer",
		Assertion: signedToken,
	}

	req, err := web.CreatePostRequest(Account.AuthUri, reqJson)
	if err != nil {
		return "", err
	}
	var resJson response
	if _, err := web.DispatchRequest(req, &resJson); err != nil {
		return "", err
	}

	expires := time.Duration((resJson.ExpiresIn - 20) * int(time.Second))
	accessToken = resJson.AccessToken
	accessExpires = time.Now().Add(expires)
	if log.IsDbg {
		log.Dbg("< got access with %d characters until %s",
			len(accessToken), accessExpires.Format(time.DateTime))
	}
	return accessToken, nil
}

func UseAccessToken() (string, error) {
	if len(accessToken) > 0 && time.Now().Compare(accessExpires) < 0 {
		return accessToken, nil
	}
	return RefreshAccessToken()
}
