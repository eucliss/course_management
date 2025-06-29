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


Dont make any changes to my code right now.
What do you think the best database choice would be for this particular app? Here is some information to consider:
The course structure in courses allows for long text
The users will need their own table
each user will be able to add courses and see other users courses
the users will be able to follow other golfers
we will need an activity feed based on what other golfers have done that are on their friends lists
Each course will need to consolidate reviews for an average rating across all golfers and handicap
advanced filters for all information

-- Users table
users (id, username, email, handicap, created_at, updated_at)

-- Courses table with JSONB for your complex course data
courses (id, name, address, course_data JSONB, created_by, created_at)

-- User follows/friendships
user_follows (follower_id, following_id, created_at)

-- Course reviews (separate from course data for aggregation)
course_reviews (id, course_id, user_id, overall_rating, individual_ratings JSONB, review_text, created_at)

-- User course scores
user_scores (id, course_id, user_id, score, handicap, date_played, created_at)

-- Activity feed items
activities (id, user_id, activity_type, course_id, data JSONB, created_at)