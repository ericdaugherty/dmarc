# DMARC Inbound

This package provides a Lambda method to process incoming DMARC reports via S3. To setup for usage:

- Configure your domain with a valid DMARC Entry
- Configure SES to receive email for a domain you control. You can use sub-domains if you don't want all your email to go through SES. For example, you can setup @dmarc.example.com and direct it to SES. 
- Configure SES to send emails from a domain you control.
- Edit the serverless.yml file. Change the environment variables 'MAILFROM' and 'MAILTO'. These are the To and From addresses used by the Lamdba to send notification emails via SES.
- Deploy this package to AWS via Serverless. 'make deploy'
- Setup an SES 'Rule Set' to have all email to your preferred address stored in the S3 bucket configured in the serverless.yml
- Update your domain DMARC entry with the correct mailto: so reports are routed to your SES domain.

All incoming email to your SES Address will be processed and you will receive an email any time any of your messaged are marked 'quarantine' or 'reject'.