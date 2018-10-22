package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/satori/go.uuid"
)

func convertProxiedRequest(request events.APIGatewayProxyRequest) (http.Request, error) {
	reqHeaders := make(http.Header)

	for k := range request.Headers {
		reqHeaders.Add(k, request.Headers[k])
	}

	r := http.Request{
		Header: reqHeaders,
		Method: request.HTTPMethod,
	}
	r.Body = ioutil.NopCloser(bytes.NewReader([]byte(request.Body)))

	err := r.ParseForm()

	if err != nil {
		log.Printf("Error parsing multipart form: %s", err.Error())
		return http.Request{}, err
	}

	return r, nil
}

func createOutputMap(formParams url.Values) map[string]string {
	email := headOrEmpty(formParams, "email")

	fieldsVar := os.Getenv("ADDITIONAL_FIELDS")
	additionalFields := strings.Split(fieldsVar, ",")

	output := make(map[string]string)

	key := ""

	for i := range additionalFields {
		key = strings.TrimSpace(additionalFields[i])
		output[key] = headOrEmpty(formParams, key)
	}

	output["email"] = email

	return output
}

func headOrEmpty(dict url.Values, key string) string {
	array := dict[key]
	if array == nil || len(array) == 0 {
		return ""
	}

	return array[0]
}

func redirect(location string) events.APIGatewayProxyResponse {
	headers := make(map[string]string)
	headers["Location"] = location

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 302,
		Headers:    headers,
	}
}

func failureRedirect() events.APIGatewayProxyResponse {
	location := os.Getenv("FAILURE_URL")
	return redirect(location)
}

func saveToBucket(jsonContent []byte) error {
	instanceUUID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	s3Bucket := os.Getenv("S3_BUCKET")

	svc := s3.New(session.New())

	outputKey := fmt.Sprintf(
		"%s.json",
		instanceUUID.String(),
	)

	saveInputs := s3.PutObjectInput{
		Body:   bytes.NewReader(jsonContent),
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(outputKey)}

	_, err = svc.PutObject(&saveInputs)

	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			switch err.Code() {
			default:
				fmt.Println(err.Error())
			}
		} else {
			return err
		}
	}

	return nil
}

func successRedirect() events.APIGatewayProxyResponse {
	location := os.Getenv("SUCCESS_URL")
	return redirect(location)
}

func handleRequest(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	r, err := convertProxiedRequest(request)

	if err != nil {
		log.Printf(
			"Error converting proxied request: (%q)", err.Error(),
		)
		return failureRedirect(), nil
	}

	outputMap := createOutputMap(r.PostForm)

	jsonResult, err := json.Marshal(outputMap)

	if err != nil {
		log.Printf(
			"Error serializing JSON: (%q)", err.Error(),
		)
		return failureRedirect(), nil
	}

	err = saveToBucket(jsonResult)

	if err != nil {
		log.Printf(
			"Error saving to S3: (%q)", err.Error(),
		)
		return failureRedirect(), nil
	}

	return successRedirect(), nil
}

func main() {
	lambda.Start(handleRequest)
}
