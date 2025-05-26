# Course Management

A golf course management system that tracks course ratings, reviews, and scores.

## Local Development

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your environment variables
3. Run `go mod download` to install dependencies
4. Run `go run main.go` to start the server

## Deployment

This application can be deployed using Railway:

1. Create a new project on [Railway](https://railway.app/)
2. Connect your GitHub repository
3. Railway will automatically detect the Dockerfile and deploy your application
4. Set your environment variables in Railway's dashboard
5. Railway will provide you with a public URL for your application

## Environment Variables

- `PORT`: The port number for the server (default: 42069)
- Add any additional environment variables here

## Tech Stack

- Go
- Echo Framework
- HTML Templates
- Docker 
```
air
```
