package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"net/url"

	"io"
	"io/ioutil"

	"github.com/pivotal-cf/credhub-cli/config"
	"github.com/pivotal-cf/credhub-cli/models"
)

func NewPutValueRequest(config config.Config, secretIdentifier string, secretContent string, overwrite bool) *http.Request {
	secret := models.SecretBody{
		ContentType: "value",
		Name:        secretIdentifier,
		Value:       secretContent,
		Overwrite:   overwrite,
	}

	return newSecretRequest("PUT", config, secretIdentifier, secret)
}

func NewPutPasswordRequest(config config.Config, secretIdentifier string, secretContent string, overwrite bool) *http.Request {
	secret := models.SecretBody{
		ContentType: "password",
		Name:        secretIdentifier,
		Value:       secretContent,
		Overwrite:   overwrite,
	}

	return newSecretRequest("PUT", config, secretIdentifier, secret)
}

func NewPutCertificateRequest(config config.Config, secretIdentifier string, root string, cert string, priv string, overwrite bool) *http.Request {
	certificate := models.Certificate{
		Ca:          root,
		Certificate: cert,
		PrivateKey:  priv,
	}
	secretBody := models.SecretBody{
		ContentType: "certificate",
		Name:        secretIdentifier,
		Value:       &certificate,
		Overwrite:   overwrite,
	}
	return newSecretRequest("PUT", config, secretIdentifier, secretBody)
}

func NewPutRsaSshRequest(config config.Config, secretIdentifier, keyType, publicKey, privateKey string, overwrite bool) *http.Request {
	key := models.RsaSsh{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}

	secret := models.SecretBody{
		ContentType: keyType,
		Name:        secretIdentifier,
		Value:       &key,
		Overwrite:   overwrite,
	}
	return newSecretRequest("PUT", config, secretIdentifier, secret)
}

func NewPutCaRequest(config config.Config, caIdentifier, caType, cert, priv string) *http.Request {
	ca := models.CaParameters{
		Certificate: cert,
		PrivateKey:  priv,
	}
	caBody := models.CaBody{
		Name:        caIdentifier,
		ContentType: caType,
		Value:       &ca,
	}

	return newCaRequest("PUT", config, caIdentifier, caBody)
}

func NewPostCaRequest(config config.Config, caIdentifier, caType string, params models.SecretParameters) *http.Request {
	caGenerateRequestBody := models.GenerateCaRequest{
		ContentType: caType,
		Name:        caIdentifier,
		Parameters:  &params,
	}

	return newCaRequest("POST", config, caIdentifier, caGenerateRequestBody)
}

func NewGetCaRequest(config config.Config, caIdentifier string) *http.Request {
	return newCaRequest("GET", config, caIdentifier, nil)
}

func NewGenerateSecretRequest(config config.Config, secretIdentifier string, parameters models.SecretParameters, contentType string, overwrite bool) *http.Request {
	generateRequest := models.GenerateSecretRequest{
		ContentType: contentType,
		Overwrite:   overwrite,
		Parameters:  &parameters,
	}

	return newSecretRequest("POST", config, secretIdentifier, generateRequest)
}

func NewRegenerateSecretRequest(config config.Config, secretIdentifier string) *http.Request {
	regenerateRequest := models.RegenerateSecretRequest{
		Regenerate: true,
	}

	return newSecretRequest("POST", config, secretIdentifier, regenerateRequest)
}

func NewGetSecretRequest(config config.Config, secretIdentifier string) *http.Request {
	return newSecretRequest("GET", config, secretIdentifier, nil)
}

func NewDeleteSecretRequest(config config.Config, secretIdentifier string) *http.Request {
	return newSecretRequest("DELETE", config, secretIdentifier, nil)
}

func NewInfoRequest(config config.Config) *http.Request {
	url := config.ApiURL + "/info"

	request, _ := http.NewRequest("GET", url, nil)

	return request
}

func NewAuthTokenRequest(cfg config.Config, user string, pass string) *http.Request {
	authUrl := cfg.AuthURL + "/oauth/token/"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Add("response_type", "token")
	data.Add("username", user)
	data.Add("password", pass)
	request, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	request.SetBasicAuth(config.AuthClient, config.AuthPassword)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return request
}

func NewRefreshTokenRequest(cfg config.Config) *http.Request {
	authUrl := cfg.AuthURL + "/oauth/token/"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", cfg.RefreshToken)
	request, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	request.SetBasicAuth(config.AuthClient, config.AuthPassword)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return request
}

func NewTokenRevocationRequest(cfg config.Config) (*http.Request, error) {
	requestUrl := cfg.AuthURL + "/oauth/token/revoke/" + cfg.RefreshToken
	request, err := http.NewRequest("DELETE", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", "Bearer "+cfg.AccessToken)
	return request, nil
}

func NewBodyClone(req *http.Request) io.ReadCloser {
	var result io.ReadCloser = nil
	if req.Body != nil {
		var bodyBytes []byte
		buf := new(bytes.Buffer)
		rc, ok := req.Body.(io.ReadCloser)
		if !ok {
			rc = ioutil.NopCloser(req.Body)
		}
		buf.ReadFrom(rc)
		bodyBytes = buf.Bytes()
		req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
		result = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}
	return result
}

func NewFindAllCredentialPathsRequest(config config.Config) *http.Request {
	url := config.ApiURL + "/api/v1/data?paths=true"

	return newRequest("GET", config, url, nil)
}

func NewFindCredentialsBySubstringRequest(config config.Config, partialSecretIdentifier string) *http.Request {
	urlString := config.ApiURL + "/api/v1/data?name-like=" + url.QueryEscape(partialSecretIdentifier)

	return newRequest("GET", config, urlString, nil)
}

func NewFindCredentialsByPathRequest(config config.Config, path string) *http.Request {
	urlString := config.ApiURL + "/api/v1/data?path=" + url.QueryEscape(path)

	return newRequest("GET", config, urlString, nil)
}

func newSecretRequest(requestType string, config config.Config, secretIdentifier string, bodyModel interface{}) *http.Request {
	var urlString string
	if requestType == "GET" {
		urlString = config.ApiURL + "/api/v1/data?name=" + url.QueryEscape(secretIdentifier) + "&current=true"
	} else if requestType == "DELETE" {
		urlString = config.ApiURL + "/api/v1/data?name=" + url.QueryEscape(secretIdentifier)
	} else if requestType == "PUT" {
		urlString = config.ApiURL + "/api/v1/data"
	} else {
		urlString = config.ApiURL + "/api/v1/data/" + secretIdentifier
	}

	return newRequest(requestType, config, urlString, bodyModel)
}

func newCaRequest(requestType string, config config.Config, caIdentifier string, bodyModel interface{}) *http.Request {
	var urlString string
	if requestType == "GET" {
		urlString = config.ApiURL + "/api/v1/ca?name=" + url.QueryEscape(caIdentifier) + "&current=true"
	} else {
		urlString = config.ApiURL + "/api/v1/ca"
	}

	return newRequest(requestType, config, urlString, bodyModel)
}

func newRequest(requestType string, config config.Config, url string, bodyModel interface{}) *http.Request {
	var request *http.Request
	if bodyModel == nil {
		request, _ = http.NewRequest(requestType, url, nil)
	} else {
		body, _ := json.Marshal(bodyModel)
		request, _ = http.NewRequest(requestType, url, bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Authorization", "Bearer "+config.AccessToken)
	return request
}
