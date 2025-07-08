# StellarServer

<div align=center>
	<img src="web/static/favicon.ico"/>
</div>

English | [中文](./README.md)

## Introduction
StellarServer is a tool with functions such as asset mapping, subdomain enumeration, information leakage detection, vulnerability scanning, directory scanning, subdomain takeover, crawler, and page monitoring. By building multiple nodes, users can freely choose nodes to run scanning tasks. When new vulnerabilities emerge, it can quickly check whether the concerned assets have related components.

## Language
Server: Go
Scanner: Go
Front-end: Vue - vue-element-plus-admin

## Current Features
- Plugin System (Add any tool through extension)
- Subdomain Enumeration
- Subdomain Takeover Detection
- Port Scanning
- ICP Automated Collection
- APP Automated Collection
- Mini Program Automated Collection
- Asset Identification
- Directory Scanning
- Vulnerability Scanning
- Sensitive Information Leakage Detection (supports scanning PDFs)
- URL Extraction
- Crawler
- Page Monitoring
- Custom WEB Fingerprint
- POC Import
- Asset Grouping
- Multi-Node Scanning
- Webhook
- Data Export

## Installation

### Requirements
- MongoDB 4.0+
- Redis 5.0+
- Go 1.18+

### Installation Steps
1. Clone the repository
```bash
git clone https://github.com/StellarServer/StellarServer.git
```

2. Enter the project directory
```bash
cd StellarServer
```

3. Install dependencies
```bash
go mod tidy
```

4. Build the project
```bash
go build -o stellar-server cmd/main.go
```

5. Run the service
```bash
./stellar-server
```

## Configuration
When running for the first time, the system will automatically generate a configuration file `config.yaml`. You can modify the parameters in the configuration file as needed:

```yaml
system:
  timezone: Asia/Shanghai
mongodb:
  ip: 127.0.0.1
  port: 27017
  mongodb_database: StellarServer
  username: root
  password: your_password
redis:
  ip: 127.0.0.1
  port: 6379
  password: your_password
logs:
  total_logs: 1000
```

## License
All branches of this project follow AGPL-3.0, and additional terms need to be followed:
1. The commercial use of this software requires a separate commercial license.
2. Companies, organizations, and for-profit entities must obtain a commercial license before using, distributing, or modifying this software.
3. Individuals and non-profit organizations are free to use this software in accordance with the terms of AGPL-3.0. 