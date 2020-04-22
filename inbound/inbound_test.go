package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DusanKasan/parsemail"
	"github.com/aws/aws-lambda-go/events"
)

func TestS3KeyParsing(t *testing.T) {

	event, err := getEvent(simpleEmailS3Event)
	if err != nil {
		t.Errorf("Error parsing JSON into S3Event. %v", err)
		return
	}

	requestedBucket := ""
	requestedKey := ""

	getEmailFunc =
		func(_ context.Context, bucket string, key string) (*parsemail.Email, error) {
			requestedBucket = bucket
			requestedKey = key
			m, err := parsemail.Parse(strings.NewReader(simpleEmailS3))
			return &m, err
		}
	handler(context.Background(), event)

	expected := "sesdmarcemailbody"
	if requestedBucket != expected {
		t.Errorf("Expected %v but got %v", expected, requestedBucket)
	}

	expected = "daakmgeqjkdmoloqru42elii1nma38fhguq2c601"
	if requestedKey != expected {
		t.Errorf("Expected %v but got %v", expected, requestedKey)
	}
}

func TestDecodeAttachmentGoogle(t *testing.T) {
	msg, err := parsemail.Parse(strings.NewReader(googleSampleZipped))
	if err != nil {
		t.Errorf("Error parsing email message. %v", err)
		return
	}

	b, err := decodeAttachment(&msg)
	if err != nil {
		t.Errorf("Error decoding attachment. %v", err)
		return
	}

	expected := googleSampleZippedXML
	if string(b) != expected {
		t.Errorf("Expectd %v\n but got %v\n", string(b), expected)
	}
}

func TestDecodeAttachmentAmazonSES(t *testing.T) {
	msg, err := parsemail.Parse(strings.NewReader(amazonsesEmail))
	if err != nil {
		t.Errorf("Error parsing email message. %v", err)
		return
	}

	b, err := decodeAttachment(&msg)
	if err != nil {
		t.Errorf("Error decoding attachment. %v", err)
		return
	}

	expected := amazonsesEmailXML
	if string(b) != expected {
		t.Errorf("Expectd %v\n but got %v\n", expected, string(b))
	}
}

func TestDecodeXMLGoogle(t *testing.T) {

	f, err := decodeXML([]byte(googleSampleZippedXML))
	if err != nil {
		t.Errorf("Error decoding XML: %v", err)
	}

	expected := "google.com"
	value := f.ReportMetadata.OrgName
	if f.ReportMetadata.OrgName != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	expected = "209.85.220.41"
	value = f.Record[0].Row.SourceIP
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}
}

func TestDecodeXMLAmazon(t *testing.T) {
	f, err := decodeXML([]byte(amazonsesEmailXML))
	if err != nil {
		t.Errorf("Error decoding XML: %v", err)
	}

	expected := "AMAZON-SES"
	value := f.ReportMetadata.OrgName
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	expected = "209.85.167.48"
	value = f.Record[0].Row.SourceIP
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	expected = "209.85.208.180"
	value = f.Record[1].Row.SourceIP
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	expected = "209.85.167.47"
	value = f.Record[2].Row.SourceIP
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	expected = "209.85.208.179"
	value = f.Record[3].Row.SourceIP
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	expected = "209.85.208.175"
	value = f.Record[4].Row.SourceIP
	if value != expected {
		t.Errorf("Expected %v but got %v", expected, value)
	}

	if len(f.Record) != 6 {
		t.Errorf("Expected %v but got %v", 6, len(f.Record))
	}

}
func TestZipXMLAttachment(t *testing.T) {

	event, err := getEvent(simpleEmailS3Event)
	if err != nil {
		t.Errorf("Error parsing JSON into S3Event. %v", err)
		return
	}

	getEmailFunc = func(_ context.Context, bucket string, key string) (*parsemail.Email, error) {
		m, err := parsemail.Parse(strings.NewReader(googleSampleZipped))
		return &m, err
	}
	handler(context.Background(), event)
}

func TestFormatEmailMessage(t *testing.T) {

	f, err := decodeXML([]byte(amazonsesEmailXML))
	if err != nil {
		t.Errorf("Error decoding XML: %v", err)
	}

	value := formatEmailMessage(f, 0)

	expected := "1 email from: 209.85.167.48 to: ericdaugherty.com was processed by AMAZON-SES and was marked none.\n"
	if value != expected {
		t.Errorf("Expected \n%v\n but got: \n%v\n", expected, value)
	}

	f.Record[0].Row.Count = "2"
	f.Record[0].Row.PolicyEvaluated.Disposition = "quarantine"
	value = formatEmailMessage(f, 0)

	expected = "2 emails from: 209.85.167.48 to: ericdaugherty.com were processed by AMAZON-SES and was marked quarantine.\n"
	if value != expected {
		t.Errorf("Expected \n%v\n but got: \n%v\n", expected, value)
	}
}

func getEvent(s string) (ses events.S3Event, e error) {
	e = json.Unmarshal([]byte(s), &ses)
	return
}

const simpleEmailS3Event = `{
    "Records": [
        {
            "eventVersion": "2.1",
            "eventSource": "aws:s3",
            "awsRegion": "us-east-1",
            "eventTime": "2020-04-18T23:24:14.093Z",
            "eventName": "ObjectCreated:Put",
            "userIdentity": {
                "principalId": "AWS:AIDAIE26RTG3F45XIHQFI"
            },
            "requestParameters": {
                "sourceIPAddress": "10.163.102.244"
            },
            "responseElements": {
                "x-amz-id-2": "eCv53HLlI2GMQa+XT5fAJ0NBzvzGDPt2WPmd9IzgXpn6KUiw7ASi3+WM+ZTYQUyKc+75mv2R7L854ho9wrQIwOb+vTMX6QCeMQBIOD1r2qM=",
                "x-amz-request-id": "51F5056F0A6797E7"
            },
            "s3": {
                "s3SchemaVersion": "1.0",
                "configurationId": "69ebe2e5-8d5a-4d6b-8819-353ed269594e",
                "bucket": {
                    "name": "sesdmarcemailbody",
                    "ownerIdentity": {
                        "principalId": "A28BG39FWST96Y"
                    },
                    "arn": "arn:aws:s3:::sesdmarcemailbody"
                },
                "object": {
                    "key": "daakmgeqjkdmoloqru42elii1nma38fhguq2c601",
                    "size": 4134,
                    "urlDecodedKey": "",
                    "versionId": "",
                    "eTag": "7b393b4a51dc6804008e3a838ee86ed5",
                    "sequencer": "005E9B8C22E2F274AD"
                }
            }
        }
    ]
}`

const simpleEmailS3 = `Return-Path: <eric@ericdaugherty.com>
Received: from mail-lj1-f175.google.com (mail-lj1-f175.google.com [209.85.208.175])
 by inbound-smtp.us-east-1.amazonaws.com with SMTP id e2mkee9fsi7998cqbar8imbaqcp7uvp7nqcgmd81
 for dmarc@dmarc.ericdaugherty.com;
 Sat, 18 Apr 2020 23:15:43 +0000 (UTC)
X-SES-Spam-Verdict: PASS
X-SES-Virus-Verdict: PASS
Received-SPF: pass (spfCheck: domain of ericdaugherty.com designates 209.85.208.175 as permitted sender) client-ip=209.85.208.175; envelope-from=eric@ericdaugherty.com; helo=mail-lj1-f175.google.com;
Authentication-Results: amazonses.com;
 spf=pass (spfCheck: domain of ericdaugherty.com designates 209.85.208.175 as permitted sender) client-ip=209.85.208.175; envelope-from=eric@ericdaugherty.com; helo=mail-lj1-f175.google.com;
 dkim=pass header.i=@ericdaugherty.com;
 dmarc=pass header.from=ericdaugherty.com;
X-SES-RECEIPT: AEFBQUFBQUFBQUFHaUpxL1ppZ0pLYzFuWmd0MmJKZUpPdkhvaFlnMG5DaXNsY1JVSzNobmRIMjlJMHdmNGh3alU2U1h5QW5HNVg2M25pTGkzcTVjQWNtQ0tqS1Y3UURBRHR1M21MOHkxMEczQTB0Rkh3Nm9tYjM2Wktoem1DWmh1aHNVZ3hHZHBTUmFyVS9lbmxRVThPSEhNdXVZTjhRU3RDa0Z0Y2REb2tMZ3pCejB6VHBwUmJUbmVlSkxyT3JDM2tpRmE5MXJlQTcrV0l5QkFnNytld2x3MlZaVE1sOGlmUXRpSVpxdGZmenpXbFZXWTA5S0JCb1dWMFhpbm1lWHNQWXg1Vkw2aERzZGErcVlmd3k3RG03OXJXSXlLVllObnRrRFZORVRKUHM0Q3k5ZDJMTkdOV0E9PQ==
X-SES-DKIM-SIGNATURE: a=rsa-sha256; q=dns/txt; b=AgLkZ4qp7iIvT/CJ/7gUFB4dZo+4THQo7AGUQk0mDY9ReVizfaaYGo6ijnv9dZ33CKVXetIif4vIlxpUPRlKKZdXVLgFbIFdrbdPDnA5DrnaqDVRHYY2sJqNCyn0i9VLDu+cP0E+HIiJZfxKac0bDlCNDZUz2EvUkoB+Z1QWyMM=; c=relaxed/simple; s=224i4yxa5dv7c2xz3womw6peuasteono; d=amazonses.com; t=1587251743; v=1; bh=EkitOlvyCRl2WGfsgUpKUGfoUkyDeB6DdTzuc2eqb44=; h=From:To:Cc:Bcc:Subject:Date:Message-ID:MIME-Version:Content-Type:X-SES-RECEIPT;
Received: by mail-lj1-f175.google.com with SMTP id q19so5968639ljp.9
        for <dmarc@dmarc.ericdaugherty.com>; Sat, 18 Apr 2020 16:15:42 -0700 (PDT)
DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=ericdaugherty.com; s=google;
        h=mime-version:from:date:message-id:subject:to;
        bh=EkitOlvyCRl2WGfsgUpKUGfoUkyDeB6DdTzuc2eqb44=;
        b=JDiWoZssnvq1UYsz3ztbFoc63fVBZB3epaESARjr0AARK9wfBL4IuPqLErLNoY8fE5
         X9DYyW+DbD4qwiVIT5GHF0bwvgCM/n00XzfDU/BAadV8Z1885p7nCsBltbvf5Jcr7lzL
         I+XO6rxBF4V8F6gxt0qEn+UHDcdKQyTN6diS2gMyM8BxfxBTvTqMx4DhssSFgAyHk4BQ
         xp615B4LJyqg1rLncGPnHEBSKdV0FSO5voIwlThuKl8Ab9SkKmk7KMxdG5TS7PIPbgIr
         z0QH4bBLkElw9XltajfbxTHUFO5uoOzS6AOaiOZFTAtfLLRj78BYbOigT/Jb0hHwt3mB
         OnlQ==
X-Google-DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=1e100.net; s=20161025;
        h=x-gm-message-state:mime-version:from:date:message-id:subject:to;
        bh=EkitOlvyCRl2WGfsgUpKUGfoUkyDeB6DdTzuc2eqb44=;
        b=c+Z1k4FRhvbPgjU/UmcRQ97LjAea24pYFx+2RZWhUBFObD2yjOXZ5EghI5jzea+ZUc
         +VnMTHZoSFpwuUKwiMc0a4x0IUV709QFndlutBSXffen1/mG1etChDrQwNfbLF3I6qRb
         dQrL3h+Iglgbzk1ff1x9mIOlGVh5rbpBq9SoXUK3t5itZMM6SnE6O/7drgktA50RT38t
         Nml+zA5M3XlCJLoLkhZCJFC0cTlDJPiE/HQ3mQR62nko7UHgAKYD5cFOzO2KictMF5O+
         1hfGBpDC7FSI1wobmNq8xvaoIuPKvbtn0q1b95Ow2rJhWkTdndBY4HyzoNFlRQk1un5i
         1i+w==
X-Gm-Message-State: AGi0PubL8zcjLLDiSet8guAoNn5dg+vNr1N1BHjiLFgXrWIeoMUGArr/
	LTUMrU6B7NoVfKVftQQO/0hdNHu3fKabiKblSOCSBMZotBQ=
X-Google-Smtp-Source: APiQypLDPsKSP8/HmEabTfZHo1is12n1KxbcUNDyI1x6j++9kyqXN7Q0dRnauX1RovW1EXEOPpuAY3ODKFcDdvD6szU=
X-Received: by 2002:a2e:a292:: with SMTP id k18mr238863lja.263.1587251741390;
 Sat, 18 Apr 2020 16:15:41 -0700 (PDT)
MIME-Version: 1.0
From: Eric Daugherty <eric@ericdaugherty.com>
Date: Sat, 18 Apr 2020 17:15:30 -0600
Message-ID: <CACLpbHEuusUBVr64zAR9zg0H2SdYmuE5L__13538QreYxrCGpA@mail.gmail.com>
Subject: Test S3 Email Saving
To: dmarc@dmarc.ericdaugherty.com
Content-Type: multipart/alternative; boundary="00000000000037097005a398d868"

--00000000000037097005a398d868
Content-Type: text/plain; charset="UTF-8"

Entire email, including this body, should get saved in S3.

Eric

--00000000000037097005a398d868
Content-Type: text/html; charset="UTF-8"

<div dir="ltr"><div>Entire email, including this body, should get saved in S3.</div><div><br></div><div>Eric<br></div></div>

--00000000000037097005a398d868--
`

const googleSampleZipped = `Delivered-To: dmarc@ericdaugherty.com
Received: by 2002:ab3:5f89:0:0:0:0:0 with SMTP id w9csp1556585ltc;
        Sat, 18 Apr 2020 02:38:22 -0700 (PDT)
X-Received: by 2002:a05:620a:1362:: with SMTP id d2mr5544665qkl.256.1587202702054;
        Sat, 18 Apr 2020 02:38:22 -0700 (PDT)
ARC-Seal: i=1; a=rsa-sha256; t=1587202702; cv=none;
        d=google.com; s=arc-20160816;
        b=XveQMoAmgarfZtjhJm1pOHkoGkJsIIbpLGQpfekZyJ/d9t8OgYA8OARg7/J89gypeG
         XGsTxOCSg8aw2X38JYrgx+l40MkUO0BXNZViSxR32IpaHngCnKlEQlGKawp2zZzNGJA+
         s2/FOU8/lFgrq/mCqalboY4X8LBJK5Kb47RYq9c3FEOU6riRYJFsJjwEx70Eyp6WbJD1
         uAx7xGtEJ071gUz1a4DW6UYDSQYVT4hPFgNtX42gK3ncDapbKfo5oT/k/GMWmez23soQ
         hZENmGJQGHiBavSx2rvlziRH2D8dvoiPJOdRuONcXFBoMO0VRG1DvIBcOG/gUKP8Woi5
         dIJQ==
ARC-Message-Signature: i=1; a=rsa-sha256; c=relaxed/relaxed; d=google.com; s=arc-20160816;
        h=content-transfer-encoding:content-disposition:to:from:subject
         :message-id:date:mime-version:dkim-signature;
        bh=lULni/JLk38jsOt6IMAW8ZWwSUi3J1nknUMO3EUrPQU=;
        b=CkEzQKv76YIluJLVCgIpRKQfU1+7bfp135zdsyyS7RaMeymAW0CD62s7RbIcNsMk4l
         v65Sv+Od5LfeP1PDQvl5Fukof51oF71yrLKnm9JJAIaTkK68an5R0D75Uxkr1PVCbUpa
         0/p+Uj5QhmsIoPnjJiVuHWEKfmlRE4w7YDASzxAOmp2zhnrkCc1GEEfZe1czI+sLO+/u
         h19rBS9p0x+n14TI6mhe4l4QGIEXHNvYBvGBYz1WkKojK3kUx1+nix7N4bEN0Ki8OmFe
         5eUJHqjWVOI3/aq8O138+iBn0ehaCpnU1MOvGfVFx4onoe2uVJvn4jfmK1jQs8J5/x5O
         STEg==
ARC-Authentication-Results: i=1; mx.google.com;
       dkim=pass header.i=@google.com header.s=20161025 header.b="CBI2/hlP";
       spf=pass (google.com: domain of noreply-dmarc-support@google.com designates 209.85.220.73 as permitted sender) smtp.mailfrom=noreply-dmarc-support@google.com;
       dmarc=pass (p=REJECT sp=REJECT dis=NONE) header.from=google.com
Return-Path: <noreply-dmarc-support@google.com>
Received: from mail-sor-f73.google.com (mail-sor-f73.google.com. [209.85.220.73])
        by mx.google.com with SMTPS id o35sor24364754qtd.48.2020.04.18.02.38.20
        for <dmarc@ericdaugherty.com>
        (Google Transport Security);
        Sat, 18 Apr 2020 02:38:22 -0700 (PDT)
Received-SPF: pass (google.com: domain of noreply-dmarc-support@google.com designates 209.85.220.73 as permitted sender) client-ip=209.85.220.73;
Authentication-Results: mx.google.com;
       dkim=pass header.i=@google.com header.s=20161025 header.b="CBI2/hlP";
       spf=pass (google.com: domain of noreply-dmarc-support@google.com designates 209.85.220.73 as permitted sender) smtp.mailfrom=noreply-dmarc-support@google.com;
       dmarc=pass (p=REJECT sp=REJECT dis=NONE) header.from=google.com
DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=google.com; s=20161025;
        h=mime-version:date:message-id:subject:from:to:content-disposition
         :content-transfer-encoding;
        bh=lULni/JLk38jsOt6IMAW8ZWwSUi3J1nknUMO3EUrPQU=;
        b=CBI2/hlP/n7a7Oq0x9vgxkR+rxpoKGMD16dCEE5XTIdgMgL8z5sudNhKOagLPlYJ+t
         4IPhwdFg7P+iGXa2u+/OmzRQkGh90QM4mD5geQzVWpW0dvjyJJeFdEJ2c1rfvXm8JplC
         xuOTCSPwPbtPKayPuMCZpEj56iSsWHMARMpDLu68FlO3YtP8RHdVOKbbJ+0Uv9UkIilc
         OFvwMXfBpRBKMXpKe9iCdSUyQknVWTNLlE9nyTfkFCshBg2P+Mmtwr0mXGKnt/aR70U/
         aDj8KMl2Wty9bOmNuhazPLr/Fssr/CFKrSDT5csyxin+0lxHDEpujFrWDCkBGd7D47Cg
         8BwQ==
X-Google-DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=1e100.net; s=20161025;
        h=x-gm-message-state:mime-version:date:message-id:subject:from:to
         :content-disposition:content-transfer-encoding;
        bh=lULni/JLk38jsOt6IMAW8ZWwSUi3J1nknUMO3EUrPQU=;
        b=U9h70OrHN75EU5yxdTuFYasnyG4ZVzrs50IuKw17LS4J6lzYf43Pt03nqoBguACaw2
         ylsXVkx+o6giNZFA29ddYM2PbafoVhi5dqtikFPmp3temN2JZk4+ZD0CU8xoywfL/L3N
         iPrLB0/JcPWiorSalQVcTvXMZTf64Dzgwto8gpQcy+t+W6y9lVt7L6OvBKV+pVWLwQs2
         Cl98ovtoDfPq13ZyuaBu9Lis5KVzoSiKTNQRAgS0vU0tx7jhxUqJeOXEkmAVut9PC0eJ
         Inq29JGFmHtSJDAeJTqmTNk2Zk5URApcPZUvB54Qix2psu4qSJ2r6CXQzA98ZKeEF/Mr
         Bcmw==
X-Gm-Message-State: AGi0PuZ8M/Do/fsKMOopmLcFagxj5AIzZKqBaqe2/mBTTnfM9UOB39qZ
	x0IuDt9DlQXgI8HF+gzcaw==
X-Google-Smtp-Source: APiQypLnDGTmF2cW3sqweFpE/jP2xbv0/Ye5tGYBNnzUCTqJHG2swN5gOzjnguJh/Ug+D5DuYKEo3L/qxgbm4g==
MIME-Version: 1.0
X-Received: by 2002:ac8:646:: with SMTP id e6mr7041467qth.191.1587202700870;
 Sat, 18 Apr 2020 02:38:20 -0700 (PDT)
Date: Fri, 17 Apr 2020 16:59:59 -0700
Message-ID: <8047041664389952409@google.com>
Subject: Report domain: ericdaugherty.com Submitter: google.com Report-ID: 8047041664389952409
From: noreply-dmarc-support@google.com
To: dmarc@ericdaugherty.com
Content-Type: application/zip; 
	name="google.com!ericdaugherty.com!1587081600!1587167999.zip"
Content-Disposition: attachment; 
	filename="google.com!ericdaugherty.com!1587081600!1587167999.zip"
Content-Transfer-Encoding: base64

UEsDBAoAAAAIACxKklAx8Xul7wEAAL4EAAA2AAAAZ29vZ2xlLmNvbSFlcmljZGF1Z2hlcnR5LmNv
bSExNTg3MDgxNjAwITE1ODcxNjc5OTkueG1srVTLcqMwELznKyjfjcCLMWwpSk75gt0zJUsD1gYk
lSSS+O9XWOKxTg457AnRM9Mz3RrATx9Dn7yBsULJx12eZrsEJFNcyO5x9/vXy77aJU/kAbcA/EzZ
K3lIEmxAK+OaARzl1NEJ86gyXSPpAKRTqushZWrAaAFDDgxU9EQqz9Bf93yghu3tqCe6521ZyIs1
H87QhinpKHONkK0iF+e0/YlQLE3XUkQRlfYdDDoUZXmsMs/1uT4QRxmCkyorTlmRl2Xxo6rr46HI
aozWcEj3UqExVHZRjIfO0AlJ8mN1yqq8zHyzgMxxkPwWzctTXXvK6T2QoX/Zlm5bT7FWvWDXRo/n
XtgLLIMo744kYATjdOwuYNw12BYjIY3yVzEQg1E4RNDq9oZNzwBpYuAPMIeRjohdITtjmjmSTwqn
w23ir6bznjJl5kGNel+ssGo0DBqhySGr0+qYHg5ZWuS+wxKYU5kapSMFRuEww7EfvNF+9ObxOTA5
IqxWVji/xH65JHgnNsgmbzJCU2t9wuJJlNzGwGLMRuNdT39bszIsOEgnWuE/oaXsApSDaVqjhq9u
aRuOfJ9YMB3dpTFgx96txHdDf28R4q5PTFFifNmoh97ftjLx0/UmzMDixLY13nj038bYOO/X8079
lBwWC6P1V/QXUEsBAgoACgAAAAgALEqSUDHxe6XvAQAAvgQAADYAAAAAAAAAAAAAAAAAAAAAAGdv
b2dsZS5jb20hZXJpY2RhdWdoZXJ0eS5jb20hMTU4NzA4MTYwMCExNTg3MTY3OTk5LnhtbFBLBQYA
AAAAAQABAGQAAABDAgAAAAA=
`

const googleSampleZippedXML = `<?xml version="1.0" encoding="UTF-8" ?>
<feedback>
  <report_metadata>
    <org_name>google.com</org_name>
    <email>noreply-dmarc-support@google.com</email>
    <extra_contact_info>https://support.google.com/a/answer/2466580</extra_contact_info>
    <report_id>8047041664389952409</report_id>
    <date_range>
      <begin>1587081600</begin>
      <end>1587167999</end>
    </date_range>
  </report_metadata>
  <policy_published>
    <domain>ericdaugherty.com</domain>
    <adkim>r</adkim>
    <aspf>r</aspf>
    <p>reject</p>
    <sp>reject</sp>
    <pct>100</pct>
  </policy_published>
  <record>
    <row>
      <source_ip>209.85.220.41</source_ip>
      <count>4</count>
      <policy_evaluated>
        <disposition>none</disposition>
        <dkim>pass</dkim>
        <spf>pass</spf>
      </policy_evaluated>
    </row>
    <identifiers>
      <header_from>ericdaugherty.com</header_from>
    </identifiers>
    <auth_results>
      <dkim>
        <domain>ericdaugherty.com</domain>
        <result>pass</result>
        <selector>google</selector>
      </dkim>
      <spf>
        <domain>ericdaugherty.com</domain>
        <result>pass</result>
      </spf>
    </auth_results>
  </record>
</feedback>
`

const amazonsesEmail = `Delivered-To: dmarc@ericdaugherty.com
Received: by 2002:ab3:5f89:0:0:0:0:0 with SMTP id w9csp2664421ltc;
        Sun, 19 Apr 2020 05:00:08 -0700 (PDT)
X-Google-Smtp-Source: APiQypKAFFDHnZoJFv7ABte8A47ERDsvqwvUgbXj818m+4sHElt6cSvMQ+snxlZHUnZg8LckYoXu
X-Received: by 2002:a37:9ad0:: with SMTP id c199mr10792529qke.472.1587297608035;
        Sun, 19 Apr 2020 05:00:08 -0700 (PDT)
ARC-Seal: i=1; a=rsa-sha256; t=1587297608; cv=none;
        d=google.com; s=arc-20160816;
        b=kGKuczgvileHlg8Hln5VO1L8lQ4QOA8Q2StDzFZA5JrKqxIK6YsUL/1Jsu48J7QIAG
         deuzsRnEGzTckIQSiFMd03GG02ePZddEw5NT6TRyoLYXt8ozHKpqvo6GWRGZOT2Le6Dg
         5eUU7/naiwF5MdmqSSHEnI1/NzT/sracK/QRwIeFRdpPQ4N0ArpjJHiiK1w/PVFegIF0
         7fkadMD1qDnGOB7Bxn9A/A5sFQOtB511WyVRCSTfb/A6i0h747d1sHn6kwpcQSHo1jRq
         RuFFZledt9l5dnBbZvtb2kL4cI1b6fPxWaUSu5KmyhDgIX2YowUuqZebSsoEqReNHV66
         9vlA==
ARC-Message-Signature: i=1; a=rsa-sha256; c=relaxed/relaxed; d=google.com; s=arc-20160816;
        h=feedback-id:mime-version:subject:message-id:to:from:date
         :dkim-signature:dkim-signature;
        bh=8W0gW80iTo1m6DIkNrxAPRjLU3vrqcbZPmAHsrZzzlM=;
        b=m9IKI1mFh5B5hUDKBpPJyvmSCUyJdPqU3qyXErp3U+F3XGvsDmJqtLEcoLS0J5t7HT
         QhMShfnyWTbFLHZw8FAM2YW3/NOcGl6mkmj15lA16ySisKO0frZzuQSvPF3B6lGw0RMI
         9Bw/O2rjXyhHHgzZzMxHR52YBhg/MMrXzLqR/J947c9eHHsQUzQt9kUY0hUFw/HVossv
         sE7THvNuoS8QarUM/8hYlCFjjU1HC1LVPCRueuDzzPdBSrrtfvsPtn6tu0adzZCk06v9
         XiSWBVqZwzGIor9Ekikii1Jk3+2VVvNj33tPikIjJz6bKy2XTDPKr4tewDDNlgXCWPHx
         nUag==
ARC-Authentication-Results: i=1; mx.google.com;
       dkim=pass header.i=@amazonses.com header.s=lq57vmtncuz5rm5xc53ilgrihvexyjk2 header.b=W+51IlYP;
       dkim=pass header.i=@amazonses.com header.s=224i4yxa5dv7c2xz3womw6peuasteono header.b="JX5/pC3K";
       spf=pass (google.com: domain of 01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@amazonses.com designates 54.240.14.87 as permitted sender) smtp.mailfrom=01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@amazonses.com;
       dmarc=pass (p=QUARANTINE sp=QUARANTINE dis=NONE) header.from=amazonses.com
Return-Path: <01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@amazonses.com>
Received: from a14-87.smtp-out.amazonses.com (a14-87.smtp-out.amazonses.com. [54.240.14.87])
        by mx.google.com with ESMTPS id h15si12855326qtc.140.2020.04.19.05.00.07
        for <dmarc@ericdaugherty.com>
        (version=TLS1_2 cipher=ECDHE-ECDSA-AES128-SHA bits=128/128);
        Sun, 19 Apr 2020 05:00:08 -0700 (PDT)
Received-SPF: pass (google.com: domain of 01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@amazonses.com designates 54.240.14.87 as permitted sender) client-ip=54.240.14.87;
Authentication-Results: mx.google.com;
       dkim=pass header.i=@amazonses.com header.s=lq57vmtncuz5rm5xc53ilgrihvexyjk2 header.b=W+51IlYP;
       dkim=pass header.i=@amazonses.com header.s=224i4yxa5dv7c2xz3womw6peuasteono header.b="JX5/pC3K";
       spf=pass (google.com: domain of 01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@amazonses.com designates 54.240.14.87 as permitted sender) smtp.mailfrom=01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@amazonses.com;
       dmarc=pass (p=QUARANTINE sp=QUARANTINE dis=NONE) header.from=amazonses.com
DKIM-Signature: v=1; a=rsa-sha256; q=dns/txt; c=relaxed/simple;
	s=lq57vmtncuz5rm5xc53ilgrihvexyjk2; d=amazonses.com; t=1587297606;
	h=Date:From:To:Message-ID:Subject:MIME-Version:Content-Type;
	bh=SyfNsHkl4G9aUvxkp14+6QQQEvrLPBBamPP+/SY4A/Y=;
	b=W+51IlYPSrvG0juBQkFmCDPyOmMRgAPIufcxv2KlLPxsFobYtgaOk6Rr/mSumN6t
	mqLQbX2moEgY7rUmsG+ZxjbLpuGvLuwBtvdbInSUZ52IL6oPzU9yvXcw3MK+oLd+X0o
	cxuIwYzSvJnUfOUp1rAB+lkdFASkVdt4C31kXYK8=
DKIM-Signature: v=1; a=rsa-sha256; q=dns/txt; c=relaxed/simple;
	s=224i4yxa5dv7c2xz3womw6peuasteono; d=amazonses.com; t=1587297606;
	h=Date:From:To:Message-ID:Subject:MIME-Version:Content-Type:Feedback-ID;
	bh=SyfNsHkl4G9aUvxkp14+6QQQEvrLPBBamPP+/SY4A/Y=;
	b=JX5/pC3KA6VPCeGw60fPAfu+T3HW0bEmOx61GtQSqJVshilb2PcRomHROmKyzU+Q
	JNRbS7CNrBy9JxoIDD60cTMrdTpj9FMJFdmTJ7cWwbr4uR5GIQZkX6CoGIlTrxH7dY1
	56ax4EIflZZQndvkyT2FU5JhQUbidsbTFIBRLCKg=
Date: Sun, 19 Apr 2020 12:00:06 +0000
From: postmaster@amazonses.com
To: dmarc@ericdaugherty.com
Message-ID: <01000171924f5b62-0dddbc87-dec5-4db3-9e07-64efe12d4aea-000000@email.amazonses.com>
Subject: Dmarc Aggregate Report Domain: {ericdaugherty.com}  Submitter:
 {Amazon SES}  Date: {2020-04-18}  Report-ID:
 {fd22fed8-be9b-4788-9fb8-ce8eb4804161}
MIME-Version: 1.0
Content-Type: multipart/mixed; 
	boundary="----=_Part_40304_1252609814.1587297606493"
X-SES-Outgoing: 2020.04.19-54.240.14.87
Feedback-ID: 1.us-east-1.CTa/CO4t1eWkL0VlHBu5/eINCZhxZraAIsQC/FZHIgk=:AmazonSES

------=_Part_40304_1252609814.1587297606493
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

This MIME email was sent through Amazon SES.
------=_Part_40304_1252609814.1587297606493
Content-Type: application/octet-stream; 
	name=amazonses.com!ericdaugherty.com!1587168000!1587254400.xml.gz
Content-Transfer-Encoding: base64
Content-Disposition: attachment; 
	filename=amazonses.com!ericdaugherty.com!1587168000!1587254400.xml.gz

H4sIAAAAAAAAAO2Xy27cIBSG15OnqLL3VZ4xI1HSLLpsu8iuGwvj4xkaGxDgadOnL2DmFqXSROqu
3tjwn8O58cmS8cOvcfhwAG24FB/vizS/fyB3uAfoWsqeyd0KRyPJ0wJnx43TNSipbTOCpR211Ekr
LPWuEXQE8vjl8fu3r8nT5yecnUTvASPlA1HS2JEaC/oTHelvKQyYlMkRZ7Pde8b4vCN9V5Y9dChp
YdsmVY1Qsu1blDBA0FYor4qNK+3s70+7kqDRVOxC2hVuYccFKdaoLjYoz3OczUowguiCqVxXlTf5
vQ+SXUU5pbhoGSs5cPbSqKkduNlDTC5dF4KA5qyj024P2r7M7UWLd6LdMx+Jxtm8CJJRfVD82wuK
aPgBzOJMhb05C2ZWFLOk8DX7hRd6SdzOPX3Bb1Tn5sqknuvU8uc8ACMnzaDhipT5NkXrtNjUaYVc
lpMh+DE5CZcPZ/MiaDEHHOgwuXGFyH4G3Lhb5tbTIqQA1/uFEn1844oa44xxBqHJPopxDOdGrpK4
+5jrx7wDYXnPHZzHCz3AIBU0vZbjW/dw7RDO7IF2oP964tIckr9Kiulk940GMw02VnHR0y1EBOb9
8dh93MwDOMXCx6H8g6BHzLJXtXu3CMltuJQ5SguUL7wsvNz+eakXXBZcbv+81NuFl4WXd/CyXnhZ
eHkHL9XCy//Oi/txOv75/gFL5aJ/Gg8AAA==
------=_Part_40304_1252609814.1587297606493--
`

const amazonsesEmailXML = `<?xml version="1.0"?>
<feedback>
	<version>0.1</version>
	<report_metadata>
		<org_name>AMAZON-SES</org_name>
		<email>postmaster@amazonses.com</email>
		<report_id>fd22fed8-be9b-4788-9fb8-ce8eb4804161</report_id>
		<date_range>
			<begin>1587168000</begin>
			<end>1587254400</end>
		</date_range>
	</report_metadata>
	<policy_published>
		<domain>ericdaugherty.com</domain>
		<adkim>r</adkim>
		<aspf>r</aspf>
		<p>reject</p>
		<sp>reject</sp>
		<pct>100</pct>
		<fo>0</fo>
	</policy_published>
	<record>
		<row>
			<source_ip>209.85.167.48</source_ip>
			<count>1</count>
			<policy_evaluated>
				<disposition>none</disposition>
				<dkim>pass</dkim>
				<spf>pass</spf>
			</policy_evaluated>
		</row>
		<identifiers>
			<envelope_from>ericdaugherty.com</envelope_from>
			<header_from>ericdaugherty.com</header_from>
		</identifiers>
		<auth_results>
			<dkim>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</dkim>
			<spf>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</spf>
		</auth_results>
	</record>
	<record>
		<row>
			<source_ip>209.85.208.180</source_ip>
			<count>1</count>
			<policy_evaluated>
				<disposition>none</disposition>
				<dkim>pass</dkim>
				<spf>pass</spf>
			</policy_evaluated>
		</row>
		<identifiers>
			<envelope_from>ericdaugherty.com</envelope_from>
			<header_from>ericdaugherty.com</header_from>
		</identifiers>
		<auth_results>
			<dkim>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</dkim>
			<spf>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</spf>
		</auth_results>
	</record>
	<record>
		<row>
			<source_ip>209.85.167.47</source_ip>
			<count>1</count>
			<policy_evaluated>
				<disposition>none</disposition>
				<dkim>pass</dkim>
				<spf>pass</spf>
			</policy_evaluated>
		</row>
		<identifiers>
			<envelope_from>ericdaugherty.com</envelope_from>
			<header_from>ericdaugherty.com</header_from>
		</identifiers>
		<auth_results>
			<dkim>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</dkim>
			<spf>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</spf>
		</auth_results>
	</record>
	<record>
		<row>
			<source_ip>209.85.208.179</source_ip>
			<count>1</count>
			<policy_evaluated>
				<disposition>none</disposition>
				<dkim>pass</dkim>
				<spf>pass</spf>
			</policy_evaluated>
		</row>
		<identifiers>
			<envelope_from>ericdaugherty.com</envelope_from>
			<header_from>ericdaugherty.com</header_from>
		</identifiers>
		<auth_results>
			<dkim>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</dkim>
			<spf>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</spf>
		</auth_results>
	</record>
	<record>
		<row>
			<source_ip>209.85.208.175</source_ip>
			<count>1</count>
			<policy_evaluated>
				<disposition>none</disposition>
				<dkim>pass</dkim>
				<spf>pass</spf>
			</policy_evaluated>
		</row>
		<identifiers>
			<envelope_from>ericdaugherty.com</envelope_from>
			<header_from>ericdaugherty.com</header_from>
		</identifiers>
		<auth_results>
			<dkim>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</dkim>
			<spf>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</spf>
		</auth_results>
	</record>
	<record>
		<row>
			<source_ip>209.85.208.174</source_ip>
			<count>1</count>
			<policy_evaluated>
				<disposition>none</disposition>
				<dkim>pass</dkim>
				<spf>pass</spf>
			</policy_evaluated>
		</row>
		<identifiers>
			<envelope_from>ericdaugherty.com</envelope_from>
			<header_from>ericdaugherty.com</header_from>
		</identifiers>
		<auth_results>
			<dkim>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</dkim>
			<spf>
				<domain>ericdaugherty.com</domain>
				<result>pass</result>
			</spf>
		</auth_results>
	</record>
</feedback>`

var noAttachment = `{"Records":[{"eventVersion":"1.0","eventSource":"aws:ses","ses":{"mail":{"commonHeaders":{"from":["Eric Daugherty \u003ceric@ericdaugherty.com\u003e"],"to":["dmarc@dmarc.ericdaugherty.com"],"returnPath":"eric@ericdaugherty.com","messageId":"\u003cCACLpbHG3ObFM_3-=JzF9J9fby=L=XLMwFcg2O_ZEfv6Cf2su6A@mail.gmail.com\u003e","date":"Sat, 18 Apr 2020 16:55:30 -0600","subject":"Test Email No Attachment"},"source":"eric@ericdaugherty.com","timestamp":"2020-04-18T22:55:43.486Z","destination":["dmarc@dmarc.ericdaugherty.com"],"headers":[{"name":"Return-Path","value":"\u003ceric@ericdaugherty.com\u003e"},{"name":"Received","value":"from mail-lj1-f174.google.com (mail-lj1-f174.google.com [209.85.208.174]) by inbound-smtp.us-east-1.amazonaws.com with SMTP id udf7vbt59mb26e76uknu0p5s6na9j6ejv31l2ro1 for dmarc@dmarc.ericdaugherty.com; Sat, 18 Apr 2020 22:55:43 +0000 (UTC)"},{"name":"X-SES-Spam-Verdict","value":"PASS"},{"name":"X-SES-Virus-Verdict","value":"PASS"},{"name":"Received-SPF","value":"pass (spfCheck: domain of ericdaugherty.com designates 209.85.208.174 as permitted sender) client-ip=209.85.208.174; envelope-from=eric@ericdaugherty.com; helo=mail-lj1-f174.google.com;"},{"name":"Authentication-Results","value":"amazonses.com; spf=pass (spfCheck: domain of ericdaugherty.com designates 209.85.208.174 as permitted sender) client-ip=209.85.208.174; envelope-from=eric@ericdaugherty.com; helo=mail-lj1-f174.google.com; dkim=pass header.i=@ericdaugherty.com; dmarc=pass header.from=ericdaugherty.com;"},{"name":"X-SES-RECEIPT","value":"AEFBQUFBQUFBQUFISHlZSXlMN1NkaEpOZStld0NMWXVHUmwwdEhrcGs1N3V6bVF3Q05mcWJzU1Z0RkRwNnQ3eVpocytuWlpsVDhWYzNnWEh5R0FUNTNSTlpXcUxWT2YxN1F3aEF3cGttQzczb1lTWnlOTVlNYktpSXplMGlyUjh4NGpTWDJZbW4wTVZUWEwwZUorZHVIRGZya0NGTXhJK1hMUzFra2dVZDZ0blZHNkx2V1lMOWpVQ1R6TDc3bVRUME5HZ1hEQytNZVhTZEtKOVZmbHZuSXFSUVhUcjh2UXE0MCswRWVHV0lTVWZWdHhPWG84LzdsTFU2U0pUcWtxQmpmSEZPbHpEYnZHbVpTbFptd3lXYzQxcGpWb1FlaTFGc3RzdGpmQXI5dXVkQllHQXU3em9weEE9PQ=="},{"name":"X-SES-DKIM-SIGNATURE","value":"a=rsa-sha256; q=dns/txt; b=TPZ3gqRuoACEmT07PupS3WFBAozVjccPiWPViq0H2aTSoIHCZlFtgb0fcgqVYXAfEsFWSNMiQGnIPjWejSbEvro5WZF+Zm2Tnwg2B4cY+Q4midn2SkjTBF7sR8nL7yRXPDRW0QsGwIiUdFp0hyipqOb3SF0ysQCpFZ92MYFvUbE=; c=relaxed/simple; s=224i4yxa5dv7c2xz3womw6peuasteono; d=amazonses.com; t=1587250543; v=1; bh=OOQwHaAXXzD7sNpwhaiD+3TvedySdjHIf/EgHglxg6w=; h=From:To:Cc:Bcc:Subject:Date:Message-ID:MIME-Version:Content-Type:X-SES-RECEIPT;"},{"name":"Received","value":"by mail-lj1-f174.google.com with SMTP id u6so5968053ljl.6 for \u003cdmarc@dmarc.ericdaugherty.com\u003e; Sat, 18 Apr 2020 15:55:43 -0700 (PDT)"},{"name":"DKIM-Signature","value":"v=1; a=rsa-sha256; c=relaxed/relaxed; d=ericdaugherty.com; s=google; h=mime-version:from:date:message-id:subject:to; bh=OOQwHaAXXzD7sNpwhaiD+3TvedySdjHIf/EgHglxg6w=; b=LTf4I0uTqn/zNsteNM5jLtx8gV3Kt8NKab5heBlphEi2HcNJ5+NUAtEY46j745uyNOWGBkvuFmoMQjuUjLqucNM0zTiT3v+BW9YfK6fXCJYspd6ESslqvKmsZIuZKj8MQeg65ZUTirWaUBpvBOpqIFbJe1cq05UcuqMMedMbaoNJg9gQPt1B7EaMwDSXXRoe1vzb2wCxyfvquvnZRCe1/lhZyhCuneaWTRXs7tyYTE2F4tNncNVjdFz4BGRb5Af5wSi8j3/AI6WAWxaM/cYfI9z2wxT1grmXjaAybeExAA20apnivld2DBYlF0C+TkcrxrY5KhC7EREo9IOi4t8sJw=="},{"name":"X-Google-DKIM-Signature","value":"v=1; a=rsa-sha256; c=relaxed/relaxed; d=1e100.net; s=20161025; h=x-gm-message-state:mime-version:from:date:message-id:subject:to; bh=OOQwHaAXXzD7sNpwhaiD+3TvedySdjHIf/EgHglxg6w=; b=DNnzcEsyn/U40ZiVzDIMFF00fnlunnO1GkAGOZT2zmwoG1V5QIWvGDvRE6wqYO+lY2 M+o5NHVNefFgczxXvn+Au2PAAfo3trcihjkZ3GyJoiJyhcYevNat7xcK1NRksvoww/4Z wvru2pOprMA2jGnCiOwsnvyPiJyzVbLvMQ4zgH7ZyU7td0sBvX9LV2xClGH63Hjt2wrV /jqTT4NfSnqpEYyQ/CkdlKWmS/qWWhQoxi+XsdN7lZax/nTbZ7zt+uCBywIszexGO4UD wDIO7QX5mQDI4SxQAWUVPfiSXhYPsYv8RB+8HL8mhRw3MbPfKGjnTpAnPXojE7rKZd71 2unA=="},{"name":"X-Gm-Message-State","value":"AGi0PuauF5Ldk6Jd9pJqls0MrLU7jZ1hdRvgoreHW38JgdaMZ2CkqK9d GZ5P4OLNIb9y7SdGhlpVPAW4qYG1QKmCpRd/IRC+Ae7Cu/w="},{"name":"X-Google-Smtp-Source","value":"APiQypL5HEoSupPcfnYIL5VcmRo3OZW/4B/mhT5fWA1St88t2mfdRirtfQJcSjkK1CKK6pCHQ7j89EnMhq0U0Do4Uz4="},{"name":"X-Received","value":"by 2002:a2e:b4f1:: with SMTP id s17mr5633493ljm.283.1587250541789; Sat, 18 Apr 2020 15:55:41 -0700 (PDT)"},{"name":"MIME-Version","value":"1.0"},{"name":"From","value":"Eric Daugherty \u003ceric@ericdaugherty.com\u003e"},{"name":"Date","value":"Sat, 18 Apr 2020 16:55:30 -0600"},{"name":"Message-ID","value":"\u003cCACLpbHG3ObFM_3-=JzF9J9fby=L=XLMwFcg2O_ZEfv6Cf2su6A@mail.gmail.com\u003e"},{"name":"Subject","value":"Test Email No Attachment"},{"name":"To","value":"dmarc@dmarc.ericdaugherty.com"},{"name":"Content-Type","value":"multipart/alternative; boundary=\"000000000000b697d605a3989019\""}],"headersTruncated":false,"messageId":"udf7vbt59mb26e76uknu0p5s6na9j6ejv31l2ro1"},"receipt":{"recipients":["dmarc@dmarc.ericdaugherty.com"],"timestamp":"2020-04-18T22:55:43.486Z","spamVerdict":{"status":"PASS"},"dkimVerdict":{"status":"PASS"},"dmarcVerdict":{"status":"PASS"},"dmarcPolicy":"","spfVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"},"action":{"type":"Lambda","invocationType":"Event","functionArn":"arn:aws:lambda:us-east-1:231107391174:function:dmarc-reporting-dev-inbound"},"processingTimeMillis":400}}}]}`
