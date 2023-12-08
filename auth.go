package cityu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Auth struct {
	username string
	password string
}

func NewAuth(username, password string) Auth {
	return Auth{
		username: username,
		password: password,
	}
}

func NewAuthFromEnv() Auth {
	return NewAuth(
		os.Getenv("AIMS_USERNAME"),
		os.Getenv("AIMS_PASSWORD"),
	)
}

const CITYU_LOGIN_URL = "https://cp37.cs.cityu.edu.hk/cp"

func (a Auth) Login() error {
	loginPageHtml, err := a.fetchLoginPage()
	if err != nil {
		return err
	}

	stateToken, err := a.getStateToken(loginPageHtml)
	if err != nil {
		return err
	}

	resBody, err := a.authenticate(*stateToken)
	if err != nil {
		return err
	}

	log.WithField("resBody", resBody).Info("login ok")

	return nil
}

var ErrLoginPageNot200 = errors.New("login page response not 200")

func (a Auth) fetchLoginPage() ([]byte, error) {
	// TODO: Timeout
	res, err := http.Get(CITYU_LOGIN_URL)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, ErrLoginPageNot200
	}

	loginPageBs, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return loginPageBs, nil
}

var ErrLoginPageHtmlStateTokenNotFound = errors.New("stateToken not found in the login page HTML")

// getStateToken uses a brute-force method to get the `stateToken` from the HTML
func (a Auth) getStateToken(html []byte) (*string, error) {
	re := regexp.MustCompile(`oktaData ?= ?\{.*"stateToken": ?(".+?").*\}`)
	m := re.FindSubmatch(html)
	if m == nil {
		return nil, ErrLoginPageHtmlStateTokenNotFound
	}

	stateTokenRaw := string(m[1])
	stateToken, err := strconv.Unquote(stateTokenRaw)
	if err != nil {
		return nil, err
	}

	return &stateToken, nil
}

const CITYU_AUTH_URL = "https://auth.cityu.edu.hk/api/v1/authn"

type ErrAuthenticationFailed struct {
	status int
	body   string
}

func (e ErrAuthenticationFailed) Error() string {
	return fmt.Sprintf("authentication failed: %d %s", e.status, e.body)
}

func (a Auth) authenticate(stateToken string) (*string, error) {
	bodyObj := map[string]interface{}{
		"password": a.password,
		"username": a.username,
		"options": map[string]interface{}{
			// TODO: Provide options for the 2 fields
			"warnBeforePasswordExpired": true,
			"multiOptionalFactorEnroll": true,
		},
	}
	bodyBs, err := json.Marshal(bodyObj)
	if err != nil {
		panic(err)
	}

	res, err := http.Post(CITYU_AUTH_URL, "application/json", bytes.NewReader(bodyBs))
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		bodyBs, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		body := string(bodyBs)

		return nil, ErrAuthenticationFailed{
			status: res.StatusCode,
			body:   body,
		}
	}

	bodyBs, err = io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	body := string(bodyBs)
	return &body, nil
}
