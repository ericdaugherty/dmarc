module github.com/ericdaugherty/dmarc/inbound

go 1.14

require (
	github.com/DusanKasan/parsemail v1.2.0
	github.com/aws/aws-lambda-go v1.32.0
	github.com/aws/aws-sdk-go-v2 v1.16.4
	github.com/aws/aws-sdk-go-v2/config v1.15.9
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.9.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.15.5
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.10
	github.com/aws/aws-sdk-go-v2/service/ses v1.14.6
)
