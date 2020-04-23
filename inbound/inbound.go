package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DusanKasan/parsemail"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

var getEmailFunc func(context.Context, string, string) (*parsemail.Email, error)
var dynamoDBTableName string
var mailFrom string
var mailTo string

type dbEntry struct {
	GMTDate          string `json:"gmtDate"`
	OrgReportID      string `json:"orgReportId"`
	S3Bucket         string `json:"s3bucket"`
	S3Key            string `json:"s3key"`
	OrgName          string `json:"orgName"`
	ReportID         string `json:"reportId"`
	BeginTime        int    `json:"beginTime"`
	EndTime          int    `json:"endTime"`
	CountAccepted    int    `json:"countAccepted"`
	CountQuarantined int    `json:"countQuarantined"`
	CountRejected    int    `json:"countRejected"`
	XML              string `json:"xml"`
}

func handler(ctx context.Context, s3Event events.S3Event) {

	fmt.Printf("Env Config: FROM: %v, TO: %v\n", mailFrom, mailTo)

	for _, record := range s3Event.Records {
		s3 := record.S3
		fmt.Printf("Processing email from bucket: %v with key: %v\n", s3.Bucket.Name, s3.Object.Key)
		msg, err := getEmailFunc(ctx, s3.Bucket.Name, s3.Object.Key)
		if err != nil {
			fmt.Printf("Error processing email. Unable to load from S3. %v\n", err)
			return
		}

		fd, err := decodeAttachment(msg)
		if err != nil {
			fmt.Printf("Error processing email. Unable to decode attachment. %v\n", err)
			return
		}

		f, err := decodeXML(fd)
		if err != nil {
			fmt.Printf("Error processing email. Unable to decode XML. %v\n", err)
			return
		}

		err = storeReport(ctx, s3.Bucket.Name, s3.Object.Key, f, fd)
		if err != nil {
			fmt.Printf("Error processing email. Unable to process report data. %v\n", err)
		}

		err = sendNotification(ctx, f)
		if err != nil {
			fmt.Printf("Error processing email. Unable to process report data. %v\n", err)
		}
	}
}

func getMailFromS3(ctx context.Context, bucket string, key string) (m *parsemail.Email, err error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}

	svc := s3.New(cfg)
	req := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	resp, err := req.Send(ctx)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	pm, e := parsemail.Parse(resp.Body)
	return &pm, e
}

func decodeAttachment(msg *parsemail.Email) (res []byte, err error) {

	var content []byte
	ct := msg.ContentType
	ct = strings.Fields(ct)[0]
	switch ct {
	case "application/zip;":
		// parsemail will decode the file for us.
		content, err = ioutil.ReadAll(msg.Content)
		return unzip(content)
	case "multipart/mixed;":
		for _, f := range msg.Attachments {
			ext := filepath.Ext(f.Filename)
			switch ext {
			case ".gz":
				return ungzip(f.Data)
			default:
				return res, errors.New("unknown file extension " + ext)
			}
		}
	default:
		return res, errors.New("unknown content type " + ct)
	}

	return res, errors.New("no reports found in email")
}

func unzip(data []byte) (b []byte, err error) {
	r := bytes.NewReader(data)
	zr, err := zip.NewReader(r, int64(r.Len()))
	if err != nil {
		return
	}

	for _, f := range zr.File {
		if strings.Contains(f.Name, ".xml") {
			var zf io.ReadCloser
			zf, err = f.Open()
			if err != nil {
				return
			}
			return ioutil.ReadAll(zf)
		}
	}

	return b, errors.New("no xml file found in zip data")
}

func ungzip(r io.Reader) (b []byte, err error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return
	}

	return ioutil.ReadAll(zr)
}

func decodeXML(data []byte) (f Feedback, err error) {
	err = xml.Unmarshal(data, &f)
	return
}

func storeReport(ctx context.Context, s3Bucket, s3Key string, f Feedback, fd []byte) (err error) {

	var countAccepted, countQuarantined, countRejected int
	for _, record := range f.Record {
		c, err := strconv.Atoi(record.Row.Count)
		if err != nil {
			c = 1
		}
		switch record.Row.PolicyEvaluated.Disposition {
		case "quarantine":
			countQuarantined += c
		case "reject":
			countRejected += c
		case "none":
			countAccepted += c
		default:
			// Ignore Unknowns here.
			fmt.Printf("Unknown disposition encountered: %v\n", record.Row.PolicyEvaluated.Disposition)
		}
	}

	beginTime, err := strconv.Atoi(f.ReportMetadata.DateRange.Begin)
	if err != nil {
		return
	}
	endTime, err := strconv.Atoi(f.ReportMetadata.DateRange.End)
	if err != nil {
		return
	}

	unixBeginTime := time.Unix(int64(beginTime), 0).UTC()

	entry := dbEntry{
		GMTDate:          unixBeginTime.Format("2006-01-02"),
		OrgReportID:      f.ReportMetadata.OrgName + ":" + f.ReportMetadata.ReportID,
		S3Bucket:         s3Bucket,
		S3Key:            s3Key,
		OrgName:          f.ReportMetadata.OrgName,
		ReportID:         f.ReportMetadata.ReportID,
		BeginTime:        beginTime,
		EndTime:          endTime,
		CountAccepted:    countAccepted,
		CountQuarantined: countQuarantined,
		CountRejected:    countRejected,
		XML:              string(fd),
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}

	svc := dynamodb.New(cfg)
	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(dynamoDBTableName),
	}

	req := svc.PutItemRequest(input)
	_, err = req.Send(ctx)

	return
}

func sendNotification(ctx context.Context, f Feedback) (err error) {

	message := ""

	for i, record := range f.Record {
		switch record.Row.PolicyEvaluated.Disposition {
		case "quarantine":
			message += formatEmailMessage(f, i)
			fmt.Printf("Processed record with quarantine.\n")
		case "reject":
			message += formatEmailMessage(f, i)
			fmt.Printf("Processed record with reject.\n")
		case "none":
			// success path, ignore.
		default:
			return errors.New("unknown disposition " + record.Row.PolicyEvaluated.Disposition)
		}
	}

	if message != "" {
		body := fmt.Sprintf("Processed Records with issues.\n\n%v", message)
		return sendEmail(ctx, "DMARC Issues Detected", body)
	}

	return nil
}

func formatEmailMessage(f Feedback, i int) string {
	message := "%v email%v from: %v to: %v %v processed by %v and was marked %v.\n"

	r := f.Record[i]
	plural := ""
	plural2 := "was"
	if r.Row.Count != "1" {
		plural = "s"
		plural2 = "were"
	}
	return fmt.Sprintf(message,
		r.Row.Count,
		plural,
		r.Row.SourceIP,
		f.PolicyPublished.Domain,
		plural2,
		f.ReportMetadata.OrgName,
		r.Row.PolicyEvaluated.Disposition)
}

func sendEmail(ctx context.Context, subject string, body string) (err error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}

	em := ses.SendEmailInput{
		Destination: &ses.Destination{ToAddresses: []string{mailTo}},
		Source:      &mailFrom,
		Message: &ses.Message{
			Subject: &ses.Content{Data: &subject},
			Body: &ses.Body{
				Text: &ses.Content{Data: &body},
			},
		},
	}

	svc := ses.New(cfg)
	req := svc.SendEmailRequest(&em)
	_, err = req.Send(ctx)

	return
}

func main() {

	getEmailFunc = getMailFromS3

	dynamoDBTableName = os.Getenv("TABLENAME")
	mailFrom = os.Getenv("MAILFROM")
	mailTo = os.Getenv("MAILTO")

	lambda.Start(handler)
}
