# DMARC Report Processor
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

This is a simple DMARC Report processor. It consists of multiple modules:

## Inbound
The inbound module (in /inbound) contains an AWS Lamdba function that processes inbound DMARC report emails and checks for issues. See the [Inbound README](./inbound) for details.

## Web
The web module (in /web) contains an AWS Lamdba function that serves HTML reports on the processed DMARC reports. See the [Web README](./web) for details

