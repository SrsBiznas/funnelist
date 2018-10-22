package main

import (
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHeadOrEmptyOnMultipleEntries(t *testing.T) {
	m := make(map[string][]string, 2)

	element1 := make([]string, 2)
	element1[0] = "success"
	element1[1] = "firstFail"

	element2 := make([]string, 1)
	element2[0] = "secondFail"

	m["key"] = element1
	m["notKey"] = element2

	result := headOrEmpty(m, "key")

	if result != "success" {
		t.Errorf("Multiple entries did not retrieve head correctly")
	}
}

func TestHeadOrEmptyOnSingleEntries(t *testing.T) {
	m := make(map[string][]string, 2)

	element1 := make([]string, 1)
	element1[0] = "success"

	element2 := make([]string, 1)
	element2[0] = "secondFail"

	m["key"] = element1
	m["notKey"] = element2

	result := headOrEmpty(m, "key")

	if result != "success" {
		t.Errorf("Multiple entries did not retrieve head correctly")
	}
}

func TestHeadOrEmptyOnEmptyEntries(t *testing.T) {
	m := make(map[string][]string, 2)

	element1 := make([]string, 0)

	element2 := make([]string, 1)
	element2[0] = "secondFail"

	m["key"] = element1
	m["notKey"] = element2

	result := headOrEmpty(m, "key")

	if result != "" {
		t.Errorf("Multiple entries did not retrieve head correctly")
	}
}

func TestHeadOrEmptyOnMissingMapEntries(t *testing.T) {
	m := make(map[string][]string, 1)

	element2 := make([]string, 1)
	element2[0] = "secondFail"

	m["notKey"] = element2

	result := headOrEmpty(m, "key")

	if result != "" {
		t.Errorf("Multiple entries did not retrieve head correctly")
	}
}

func TestConvertProxiedRequest(t *testing.T) {
	proxyHeaders := make(map[string]string)
	proxyHeaders["Content-Type"] = "application/x-www-form-urlencoded"

	proxyBody := "email=noreply%2Btest%40srs.bizn.as&secondary=expected_value"

	proxiedRequest := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Headers:    proxyHeaders,
		Body:       proxyBody,
	}

	httpRequest, err := convertProxiedRequest(proxiedRequest)

	if err != nil {
		t.Errorf("Error converting proxied request to http request (%q)", err)
	}

	if httpRequest.Method != "POST" {
		t.Errorf("HTTP Method is incorrect: %s", httpRequest.Method)
	}

	headers := httpRequest.Header
	contentType := headers.Get("Content-Type")

	expectedContentType := "application/x-www-form-urlencoded"

	if expectedContentType != contentType {
		t.Errorf("Content Type did not match: (%q)", contentType)
	}

	form := httpRequest.PostForm

	if form == nil {
		t.Errorf("Post form is empty when is should not be")
	}

	ev := url.Values{}

	ev.Add("email", "noreply+test@srs.bizn.as")
	ev.Add("secondary", "expected_value")

	if !reflect.DeepEqual(form, ev) {
		t.Errorf("Post form does not match expected: (%p)", form)
	}
}

func TestConvertProxiedRequestWithMixedCaseHeaders(t *testing.T) {
	proxyHeaders := make(map[string]string)
	proxyHeaders["content-type"] = "application/x-www-form-urlencoded"

	proxyBody := "email=noreply%2Btest%40srs.bizn.as&secondary=expected_value"

	proxiedRequest := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Headers:    proxyHeaders,
		Body:       proxyBody,
	}

	httpRequest, err := convertProxiedRequest(proxiedRequest)

	if err != nil {
		t.Errorf("Error converting proxied request to http request (%q)", err)
	}

	if httpRequest.Method != "POST" {
		t.Errorf("HTTP Method is incorrect: %s", httpRequest.Method)
	}

	headers := httpRequest.Header
	contentType := headers.Get("Content-Type")

	expectedContentType := "application/x-www-form-urlencoded"

	if expectedContentType != contentType {
		t.Errorf("Content Type did not match: (%q)", contentType)
	}

	form := httpRequest.PostForm

	if form == nil {
		t.Errorf("Post form is empty when is should not be")
	}

	ev := url.Values{}

	ev.Add("email", "noreply+test@srs.bizn.as")
	ev.Add("secondary", "expected_value")

	if !reflect.DeepEqual(form, ev) {
		t.Errorf("Post form does not match expected: (%p)", form)
	}
}

func TestEnsureOutputMapLinesUp(t *testing.T) {

	postForm := url.Values{}

	// Add Env Vars
	os.Setenv("ADDITIONAL_FIELDS", "secondary,missing")

	postForm.Add("email", "noreply+test@srs.bizn.as")
	postForm.Add("secondary", "expected_value")
	postForm.Add("ignored", "failure")

	outputMap := createOutputMap(postForm)

	expectedMap := make(map[string]string, 2)
	expectedMap["email"] = "noreply+test@srs.bizn.as"
	expectedMap["secondary"] = "expected_value"
	expectedMap["missing"] = ""

	if !reflect.DeepEqual(outputMap, expectedMap) {
		t.Errorf("Post form does not match expected: (%s)", outputMap)
	}
}

func TestSuccessRedirect(t *testing.T) {
	expectedLocation := "https://success.test.srs.bizn.as"

	os.Setenv("SUCCESS_URL", expectedLocation)

	result := successRedirect()

	if result.StatusCode != 302 {
		t.Errorf("Incorrect Status Code: %d", result.StatusCode)
	}

	location := result.Headers["Location"]

	if location != expectedLocation {
		t.Errorf("Incorrect redirect location header: %s", location)
	}
}

func TestFailureRedirect(t *testing.T) {
	expectedLocation := "https://failure.test.srs.bizn.as"

	os.Setenv("FAILURE_URL", expectedLocation)

	result := failureRedirect()

	if result.StatusCode != 302 {
		t.Errorf("Incorrect Status Code: %d", result.StatusCode)
	}

	location := result.Headers["Location"]

	if location != expectedLocation {
		t.Errorf("Incorrect redirect location header: %s", location)
	}
}
