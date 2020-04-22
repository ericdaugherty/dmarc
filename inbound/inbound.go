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
	"strings"

	"github.com/DusanKasan/parsemail"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

var getEmailFunc func(context.Context, string, string) (*parsemail.Email, error)
var mailFrom string
var mailTo string

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
		err = processReport(ctx, f)
		if err != nil {
			fmt.Printf("Error processing email. Unable to process report data. %v\n", err)
		}

		// TODO: Delete file from S3
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

func processReport(ctx context.Context, f Feedback) (err error) {

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
			fmt.Printf("Processed record with no issues.\n")
		default:
			return errors.New("unknown disposition " + record.Row.PolicyEvaluated.Disposition)
		}
	}

	if message != "" {
		body := fmt.Sprintf("Processed Records with issues.\n\n%v", message)
		sendEmail(ctx, "DMARC Issues Detected", body)
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

	mailFrom = os.Getenv("MAILFROM")
	mailTo = os.Getenv("MAILTO")

	lambda.Start(handler)
}
