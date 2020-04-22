package main

import "encoding/xml"

// Feedback maps the DMARC XML report to a struct
type Feedback struct {
	XMLName        xml.Name `xml:"feedback"`
	Text           string   `xml:",chardata"`
	Version        string   `xml:"version"`
	ReportMetadata struct {
		Text      string `xml:",chardata"`
		OrgName   string `xml:"org_name"`
		Email     string `xml:"email"`
		ReportID  string `xml:"report_id"`
		DateRange struct {
			Text  string `xml:",chardata"`
			Begin string `xml:"begin"`
			End   string `xml:"end"`
		} `xml:"date_range"`
	} `xml:"report_metadata"`
	PolicyPublished struct {
		Text   string `xml:",chardata"`
		Domain string `xml:"domain"`
		Adkim  string `xml:"adkim"`
		Aspf   string `xml:"aspf"`
		P      string `xml:"p"`
		Sp     string `xml:"sp"`
		Pct    string `xml:"pct"`
		Fo     string `xml:"fo"`
	} `xml:"policy_published"`
	Record []struct {
		Text string `xml:",chardata"`
		Row  struct {
			Text            string `xml:",chardata"`
			SourceIP        string `xml:"source_ip"`
			Count           string `xml:"count"`
			PolicyEvaluated struct {
				Text        string `xml:",chardata"`
				Disposition string `xml:"disposition"`
				Dkim        string `xml:"dkim"`
				Spf         string `xml:"spf"`
			} `xml:"policy_evaluated"`
		} `xml:"row"`
		Identifiers struct {
			Text         string `xml:",chardata"`
			EnvelopeFrom string `xml:"envelope_from"`
			HeaderFrom   string `xml:"header_from"`
		} `xml:"identifiers"`
		AuthResults struct {
			Text string `xml:",chardata"`
			Dkim struct {
				Text   string `xml:",chardata"`
				Domain string `xml:"domain"`
				Result string `xml:"result"`
			} `xml:"dkim"`
			Spf struct {
				Text   string `xml:",chardata"`
				Domain string `xml:"domain"`
				Result string `xml:"result"`
			} `xml:"spf"`
		} `xml:"auth_results"`
	} `xml:"record"`
}
