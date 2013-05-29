/*

Package panda implements PandaStream API as outlined in http://www.pandastream.com/docs/api

Public functions defined are:

    Init(AccessKey string, SecretKey string, CloudId string, ApiHost string, ApiPort int)

    ApiURL() string
    Get(path string, data map[string]string) (string, error)
    Post(path string, data map[string]string) (string, error)
    Put(path string, data map[string]string) (string, error)
    Delete(path string, data map[string]string) (string, error)

*/

package panda

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ApiHost string = "api.pandastream.com"
const ApiPort int = 443
const ApiVersion = 2

type PandaApi struct {
	AccessKey  string
	SecretKey  string
	CloudId    string
	ApiHost    string
	ApiPort    int
	ApiVersion int
	transport  *http.Transport
	client     *http.Client
}

type PandaApiInterface interface {
	Init(AccessKey string, SecretKey string, CloudId string, ApiHost string, ApiPort int)
	apiProtocol() string
	apiPath() string
	ApiURL() string
	generateTimestamp() string
	signedParams(http_verb string, path string, data map[string]string, timestamp string) (map[string]string, error)
	generateSignature(http_verb string, request_uri string, params map[string]string) (string, error)
	canonicalQS(m map[string]string) string
	httpRequest(http_verb string, path string, data map[string]string) (string, error)
	Get(path string, data map[string]string) (string, error)
	Post(path string, data map[string]string) (string, error)
	Put(path string, data map[string]string) (string, error)
	Delete(path string, data map[string]string) (string, error)
}

func Version() string {
	return "0.1.0"
}

func (api PandaApi) generateTimestamp() string {
	timenow := time.Now().UTC()
	return timenow.Format("2006-01-02T15:04:05.999999+00:00")
}

func URLEscape(s string) string {
	new_s := strings.Replace(url.QueryEscape(s), "%7E", "~", -1)
	new_s = strings.Replace(new_s, " ", "%20", -1)
	new_s = strings.Replace(new_s, "/", "%2F", -1)
	return new_s
}

func (api PandaApi) apiProtocol() string {
	if api.ApiPort == 443 {
		return "https"
	} else {
		return "http"
	}
}

func (api PandaApi) apiPath() string {
	return "/v" + strconv.Itoa(api.ApiVersion)
}

// returns current base API URL
func (api PandaApi) ApiURL() string {
	return api.apiProtocol() + "://" + api.ApiHost + api.apiPath()
}

// sign the request with signature as outlined in the API docs
func (api PandaApi) signedParams(http_verb string, path string, data map[string]string, timestamp string) (map[string]string, error) {

	AuthParams := map[string]string{"cloud_id": api.CloudId, "access_key": api.AccessKey, "timestamp": timestamp}

	for k, v := range data {
		AuthParams[URLEscape(k)] = URLEscape(v)
	}

	AdditionalParams := make(map[string]string)
	for k, v := range AuthParams {
		AdditionalParams[k] = v
	}

	delete(AdditionalParams, "file")

	signature, err := api.generateSignature(http_verb, path, AdditionalParams)
	if err != nil {
		return nil, err
	}

	AuthParams["signature"] = signature
	return AuthParams, nil
}

// build POST request
func (api PandaApi) buildPostRequest(path string, params map[string]string, file string, FileType string) (*http.Request, error) {

	boundary, end := "^{---panda---}v", "\r\n"

	fp, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	fstat, err := fp.Stat()
	if err != nil {
		return nil, err
	}

	FileSize := fstat.Size()

	BodyHeader := bytes.NewBuffer(nil)

	for k, v := range params {
		BodyHeader.WriteString(fmt.Sprintf("--%s%s", boundary, end))
		BodyHeader.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"%s%s", k, end, end))
		BodyHeader.WriteString(fmt.Sprintf("%s%s", v, end))
	}

	BodyHeader.WriteString(fmt.Sprintf("--%s%s", boundary, end))
	BodyHeader.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"; filename=\"%s\"%s", file, file, end))
	BodyHeader.WriteString(fmt.Sprintf("Content-Type: %s%s%s", FileType, end, end))

	BodyFooter := bytes.NewBufferString(end + "--" + boundary + "--" + end)

	r, w := io.Pipe()

	go func() {
		BodySlices := []io.Reader{BodyHeader, fp, BodyFooter}

		for _, k := range BodySlices {
			_, err = io.Copy(w, k)
			if err != nil {
				w.CloseWithError(err)
				return
			}
		}

		fp.Close()
		w.Close()
	}()

	BodyLen := int64(BodyHeader.Len()) + FileSize + int64(BodyFooter.Len())

	HttpHeader := make(http.Header)
	HttpHeader.Add("Content-Type", "multipart/form-data; boundary="+boundary)

	RealUrl, _ := url.Parse(path)

	PostRequest := &http.Request{
		Method:        "POST",
		URL:           RealUrl,
		Host:          api.ApiHost,
		Header:        HttpHeader,
		Body:          r,
		ContentLength: BodyLen,
	}

	return PostRequest, nil
}

// builds generic HTTP request
func (api PandaApi) httpRequest(http_verb string, path string, data map[string]string) (string, error) {

	var HttpReq *http.Request
	var err error

	signedParams, err := api.signedParams(http_verb, path, data, api.generateTimestamp())
	if err != nil {
		return "", err
	}

	CanonicalQS := api.canonicalQS(signedParams)

	var RequestURL string = api.ApiURL() + path + "?" + CanonicalQS

	if file, ok := data["file"]; ok {
		if strings.ToUpper(http_verb) == "POST" {
			post_req, err := api.buildPostRequest(RequestURL, data, file, "application/octet-stream")
			if err != nil {
				return "", err
			}
			HttpReq = post_req
		}
	} else {
		HttpReq, err = http.NewRequest(http_verb, RequestURL, nil)
	}

	resp, err := api.client.Do(HttpReq)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

// builds the required query string - the keys have to be sorted
func (api PandaApi) canonicalQS(m map[string]string) string {
	var keys []string
	var qs []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		qs = append(qs, URLEscape(k)+"="+URLEscape(m[k]))
	}

	return strings.Join(qs, "&")
}

// generates the signature using HMAC/SHA256
func (api PandaApi) generateSignature(http_verb string, request_uri string, params map[string]string) (string, error) {

	qs := api.canonicalQS(params)

	var s []string = []string{strings.ToUpper(http_verb), strings.ToLower(api.ApiHost), request_uri, qs}

	string_to_sign := strings.Join(s, "\n")
	mac := hmac.New(sha256.New, []byte(api.SecretKey))
	mac.Write([]byte(string_to_sign))
	gmac := mac.Sum(nil)

	return strings.Trim(base64.StdEncoding.EncodeToString(gmac), " "), nil
}

func (api PandaApi) Get(path string, data map[string]string) (string, error) {
	return api.httpRequest("GET", path, data)
}

func (api PandaApi) Post(path string, data map[string]string) (string, error) {
	return api.httpRequest("POST", path, data)
}

func (api PandaApi) Put(path string, data map[string]string) (string, error) {
	return api.httpRequest("PUT", path, data)
}

func (api PandaApi) Delete(path string, data map[string]string) (string, error) {
	return api.httpRequest("DELETE", path, data)
}

// initialise the client
func (api *PandaApi) Init(AccessKey string, SecretKey string, CloudId string, ApiHost string, ApiPort int) {
	api.transport = &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: false},
		DisableCompression: false,
	}

	api.client = &http.Client{Transport: api.transport}

	api.AccessKey = AccessKey
	api.SecretKey = SecretKey
	api.CloudId = CloudId
	api.ApiHost = ApiHost
	api.ApiPort = ApiPort
	api.ApiVersion = ApiVersion
}
