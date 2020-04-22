# DMARC Report Processor
This is a simple DMARC Report processor. It consists of multiple modules:

## Inbound
The inbound module (in /inbound) contains an AWS Lamdba function that processes inbound DMARC report emails and checks for issues. See the [Inbound README](./inbound) for details.

Additional modules will be added to support additional report viewing, etc.