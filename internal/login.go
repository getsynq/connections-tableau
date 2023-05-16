package internal

import (
	"bytes"
	"encoding/json"
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

func LoginUserPassword(baseURL, apiVersion, site, username, password string) (token, siteId string, err error) {

	var payload = []byte(fmt.Sprintf(`
<tsRequest>
  <credentials name="%s" password="%s" >
    <site contentUrl="%s" />
  </credentials>
</tsRequest>
`, username, password, site))

	return login(baseURL, apiVersion, payload)
}

func LoginPersonalAccessToken(baseURL, apiVersion, site, tokenName, tokenValue string) (token, siteId string, err error) {

	var payload = []byte(fmt.Sprintf(`
<tsRequest>
  <credentials personalAccessTokenName="%s"
    personalAccessTokenSecret="%s" >
	  <site contentUrl="%s" />
  </credentials>
</tsRequest>
`, tokenName, tokenValue, site))

	return login(baseURL, apiVersion, payload)
}

type ServerInfoResponse struct {
	ServerInfo struct {
		ProductVersion struct {
			Value string `json:"value"`
			Build string `json:"build"`
		} `json:"productVersion"`
		RestApiVersion string `json:"restApiVersion"`
	} `json:"serverInfo"`
}

func GetVersion(connectionUri string) (string, error) {
	client := &http.Client{}
	versionReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/2.4/serverInfo", connectionUri), nil)
	if err != nil {
		return "", err
	}
	versionReq.Header.Set("accept", "application/json")
	versionResp, err := client.Do(versionReq)
	if err != nil {
		return "", err
	}
	defer versionResp.Body.Close()
	versionRespJson, err := io.ReadAll(versionResp.Body)
	if err != nil {
		return "", err
	}

	serverInfoResponse := &ServerInfoResponse{}
	err = json.Unmarshal(versionRespJson, serverInfoResponse)
	if err != nil {
		return "", err
	}
	fmt.Printf("Tableau server version: %s\n", serverInfoResponse.ServerInfo.ProductVersion.Value)
	fmt.Printf("Tableau API version: %s\n", serverInfoResponse.ServerInfo.RestApiVersion)

	return serverInfoResponse.ServerInfo.RestApiVersion, nil
}

func login(baseURL, apiVersion string, payload []byte) (token, siteId string, err error) {
	loginURL := fmt.Sprintf("%s/api/%s/auth/signin", baseURL, apiVersion)
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
