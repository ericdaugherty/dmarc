package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-chi/chi"
)

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

func (e dbEntry) BeginTimeFormatted() string {
	return time.Unix(int64(e.BeginTime), 0).UTC().Format(time.RFC3339)
}

type aggEntry struct {
	GMTDate          string
	CountAccepted    int
	CountQuarantined int
	CountRejected    int
}

type web struct {
	devMode   bool
	tmpl      *template.Template
	templates map[string]*template.Template
}

func (web *web) home(w http.ResponseWriter, r *http.Request) {
	web.initTemplates()

	entries, err := web.queryRecentReports(7)
	if err != nil {
		web.errorHandler(w, r, err.Error())
	}

	templateData := make(map[string]interface{})
	templateData["entries"] = entries

	web.renderTemplate(w, r, "home", templateData)
}

func (web *web) date(w http.ResponseWriter, r *http.Request) {
	web.initTemplates()

	date := chi.URLParam(r, "date")

	entries, err := web.getReports(date)
	if err != nil {
		web.errorHandler(w, r, err.Error())
	}

	templateData := make(map[string]interface{})
	templateData["date"] = date
	templateData["entries"] = entries

	web.renderTemplate(w, r, "date", templateData)
}

func (*web) errorHandler(w http.ResponseWriter, r *http.Request, errorDesc string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Server Error: %v", errorDesc)
}

func (*web) queryRecentReports(days int) (entries []aggEntry, err error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}

	svc := dynamodb.New(cfg)

	attributeValues := map[string]dynamodb.AttributeValue{}
	filterExpression := ""

	current := time.Now().UTC()
	for i := 0; i < days; i++ {
		current = current.AddDate(0, 0, -1)
		vName := fmt.Sprintf(":v%v", i)
		attributeValues[vName] = dynamodb.AttributeValue{S: aws.String(current.Format("2006-01-02"))}
		or := ""
		if i > 0 {
			or = " OR "
		}
		filterExpression += fmt.Sprintf("%vgmtDate = %v", or, vName)
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeValues: attributeValues,
		FilterExpression:          aws.String(filterExpression),
		TableName:                 aws.String("dmarcReports"),
	}
	req := svc.ScanRequest(input)

	result, err := req.Send(context.Background())
	if err != nil {
		return
	}

	aggEntries := map[string]aggEntry{}

	for _, r := range result.Items {
		var entry dbEntry
		err = dynamodbattribute.UnmarshalMap(r, &entry)
		if err != nil {
			return
		}
		e, ok := aggEntries[entry.GMTDate]
		if !ok {
			e = aggEntry{GMTDate: entry.GMTDate}
		}
		e.CountAccepted += entry.CountAccepted
		e.CountQuarantined += entry.CountQuarantined
		e.CountRejected += entry.CountRejected
		aggEntries[e.GMTDate] = e
	}

	for _, v := range aggEntries {
		entries = append(entries, v)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GMTDate < entries[j].GMTDate
	})

	return
}

func (*web) getReports(date string) (entries []dbEntry, err error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}

	svc := dynamodb.New(cfg)
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":d": {
				S: aws.String(date),
			},
		},
		KeyConditionExpression: aws.String("gmtDate = :d"),
		TableName:              aws.String("dmarcReports"),
	}

	req := svc.QueryRequest(input)
	result, err := req.Send(context.Background())
	if err != nil {
		return
	}

	for _, r := range result.Items {
		var entry dbEntry
		err = dynamodbattribute.UnmarshalMap(r, &entry)
		if err != nil {
			return
		}
		entries = append(entries, entry)
	}

	return
}
