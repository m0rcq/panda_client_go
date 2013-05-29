package panda

import (
	"testing"
)

const (
	testCloudId   = "<INSERT YOUR CLOUD ID HERE>"
	testAccessKey = "<INSERT YOUR ACCESS KEY HERE>"
	testSecretKey = "<INSERT YOUR SECRET KEY HERE>"
	testApiHost   = "api.pandastream.com"
	testApiPort   = 443
)

func Test_SignatureGeneration(t *testing.T) {
	// example values from http://www.pandastream.com/docs/api#api_authentication
	exampleTestCloudId := "123456789"
	exampleTestAccessKey := "abcdefgh"
	exampleTestSecretKey := "ijklmnop"
	exampleTestApiHost := "api.pandastream.com"
	exampleTestApiPort := 443
	exampleTestTimestamp := "2011-03-01T15:39:10.260762Z"
	exampleTestSignedSignature := "kVnZs%2FNX13ldKPdhFYoVnoclr8075DwiZF0TGgIbMsc%3D"

	client := &PandaApi{}
	client.Init(exampleTestAccessKey, exampleTestSecretKey, exampleTestCloudId, exampleTestApiHost, exampleTestApiPort)

	signedParams, err := client.signedParams("GET", "/videos.json", nil, exampleTestTimestamp)

	if err != nil {
		t.Error(err.Error())
	}

	if URLEscape(signedParams["signature"]) != exampleTestSignedSignature {
		t.Log(signedParams["signature"])
		t.Error("Signature does not match")
	}
}
