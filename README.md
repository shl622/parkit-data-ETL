# NYC Parking Meters ETL Pipeline

This project is a serverless ETL pipeline that extracts parking meter data from NYC Open Data, transforms it into a structured format, and loads it into MongoDB.

## Architecture

- **Extract**: Fetches parking meter data from NYC Open Data API
- **Transform**: Processes and normalizes the data (meter types, operational hours, locations)
- **Load**: Stores the processed data in MongoDB
- **Schedule**: Runs automatically every 7 days

## Prerequisites

- Go 1.x
- Node.js & npm
- AWS CLI configured with appropriate credentials
- MongoDB instance

## Required Environment Variables

Create a `.env` file with the following variables:

```env
NYC_API_URL=https://data.cityofnewyork.us/resource/693u-uax6.json
NYC_API_APP_TOKEN=your_api_token
MONGODB_URI=your_mongodb_connection_string
MONGODB_DATABASE=your_database_name
BATCH_SIZE=1000
```

## AWS Setup

1. Configure AWS credentials:
```bash
aws configure
```

2. Store environment variables in AWS Parameter Store:
```bash
aws ssm put-parameter --name "/parkit/nyc_api_url" --value "https://data.cityofnewyork.us/resource/693u-uax6.json" --type "String"
aws ssm put-parameter --name "/parkit/nyc_api_token" --value "your_token" --type "SecureString"
aws ssm put-parameter --name "/parkit/mongodb_uri" --value "your_uri" --type "SecureString"
aws ssm put-parameter --name "/parkit/mongodb_database" --value "your_database" --type "SecureString"
aws ssm put-parameter --name "/parkit/batch_size" --value "1000" --type "String"
```

## Installation

1. Install dependencies:
```bash
npm install
```

2. Build the Go binary:
```bash
go build -o cmd/sync/main ./cmd/sync
```

3. Deploy to AWS:
```bash
serverless deploy
```

## Development

The main components are:
- `internal/nyc/client.go`: NYC Open Data API client
- `internal/models/parking.go`: Data models
- `internal/database/mongodb.go`: MongoDB operations
- `internal/service/sync.go`: Main ETL logic
- `serverless.yml`: AWS Lambda configuration

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
