# funnelist
AWS Lambda method for handling sales funnel / email subscription
ingestion

Building
--------

To compile for use in AWS lambda, use the following command line:

    env GOOS=linux GOARCH=amd64 go build -v -o $GOPATH/bin/funnelist

AWS Configuration
=================

Resource Integration Request
----------------------------

Due to the fact that lambda functions explicitly expect JSON by default, the API Gateway must ensure that `Proxy Integration` is
enabled.

In Resources -> <Your Endpoint> -> Integration Request, check the
**Use Lambda Proxy integration** option.

Environment Variables
---------------------

The `S3_BUCKET` environment variable indicates the bucket to
store received emails in.

The `ADDITIONAL_FIELDS` environment variable allows for additional
form fields to be included (instead of just `email`).

The `SUCCESS_URL` environment variable sets the location for the
HTTP redirect when the operation completed successfully.

The `FAILURE_URL` environment variable sets the location for the
HTTP redirect when the operation fails.

Adding to Your Website
----------------------

The lambda is designed to operate with basic HTML forms. The only
required field is "email". Any additional fields may be added
and will be saved if they were included in the `ADDITIONAL_FIELDS`
environment variable.
