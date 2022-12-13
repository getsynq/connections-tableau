package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type LoginResponse struct {
	XMLName     xml.Name                 `xml:"tsResponse"`
	Credentials LoginResponseCredentials `xml:"credentials"`
}

type LoginResponseCredentials struct {
	Token string            `xml:"token,attr"`
	Site  LoginResponseSite `xml:"site"`
	User  LoginResponseUser `xml:"user"`
}

type LoginResponseSite struct {
	ID string `xml:"id,attr"`
}

type LoginResponseUser struct {
	ID string `xml:"id,attr"`
}

func LoginUserPassword(baseURL, site, username, password string) (token, siteId string, err error) {

	var payload = []byte(fmt.Sprintf(`
<tsRequest>
  <credentials name="%s" password="%s" >
    <site contentUrl="%s" />
  </credentials>
</tsRequest>
`, username, password, site))

	return login(baseURL, payload)
}

func LoginPersonalAccessToken(baseURL, site, tokenName, tokenValue string) (token, siteId string, err error) {

	var payload = []byte(fmt.Sprintf(`
<tsRequest>
  <credentials personalAccessTokenName="%s"
    personalAccessTokenSecret="%s" >
	  <site contentUrl="%s" />
  </credentials>
</tsRequest>
`, tokenName, tokenValue, site))

	return login(baseURL, payload)
}

func login(baseURL string, payload []byte) (token, siteId string, err error) {
	loginURL := fmt.Sprintf("%s/api/2.3/auth/signin", baseURL)
	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", "", fmt.Errorf("failed to create login request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send login request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to log in - server responded with status code: %d - %s", resp.StatusCode, string(body))
	}

	var loginResponse LoginResponse
	if err := xml.Unmarshal(body, &loginResponse); err != nil {
		return "", "", fmt.Errorf("unable to unmarshal response body: %w", err)
	}

	return loginResponse.Credentials.Token, loginResponse.Credentials.Site.ID, nil
}
