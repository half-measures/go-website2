# Movie Trailer Voting System

A web application that allows users to post YouTube links for movie trailers, vote on them, and view all trailers per movie without duplicates.

## Features
- User authentication (optional)
- Post YouTube trailer links with descriptions
- Vote system for each trailer
- Dedicated pages for each movie with embedded trailers
- No duplicate entries

## Tech Stack
- **Backend**: Go (Gin framework)
- **Frontend**: HTML, CSS, JavaScript
- **Database**: MYSQL

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/movie-trailer-voting.git
   ```
2. Navigate to project directory:
   ```bash
   cd movie-trailer-voting
   ```
3. Install dependencies:
   ```bash
   go get ./...
   ```
4. Set up database :
   - DB is created via Docker - 'movie_trailers'
   - migrate is used to create the correct schema via migrate -path db/migrations -database "mysql://root:secret@tcp(localhost:3306)/movie_trailers" up

## Usage
1. Start the server:
   ```bash
   go run main.go
   ```
2. Open browser at http://localhost:8080

## Contributing
Pull requests are welcome! Please follow these guidelines:
- Use descriptive commit messages
- Add tests for new features
- Maintain consistent code style

## License
This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.
