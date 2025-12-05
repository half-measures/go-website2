package models

import (
	"database/sql"
	"time"
)

// Movie represents a movie in the database.
type Movie struct {
	ID          int
	Title       string
	Slug        string
	ReleaseYear int
	Description string
	CreatedAt   time.Time
}

// Trailer represents a movie trailer.
type Trailer struct {
	ID        int
	MovieID   int
	Title     string
	YouTubeID string
	CreatedAt time.Time
	Votes     int // This field will be populated by a custom query
}

// Vote represents a vote on a trailer.
type Vote struct {
	ID         int
	TrailerID  int
	IPAddress  string
	VoteValue  int
	CreatedAt  time.Time
}

// TrailerModel wraps the database connection.
type TrailerModel struct {
	DB *sql.DB
}

// GetTrailersByMovieSlug fetches all trailers for a given movie slug,
// along with their vote counts.
func (m *TrailerModel) GetTrailersByMovieSlug(slug string) ([]*Trailer, error) {
	// SQL query to get trailers and their vote counts
	stmt := `
        SELECT t.id, t.movie_id, t.title, t.youtube_id, t.created_at, COALESCE(SUM(v.vote_value), 0) AS votes
        FROM trailers t
        JOIN movies m ON t.movie_id = m.id
        LEFT JOIN votes v ON t.id = v.trailer_id
        WHERE m.slug = ?
        GROUP BY t.id
        ORDER BY votes DESC
    `

	rows, err := m.DB.Query(stmt, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trailers []*Trailer
	for rows.Next() {
		t := &Trailer{}
		err := rows.Scan(&t.ID, &t.MovieID, &t.Title, &t.YouTubeID, &t.CreatedAt, &t.Votes)
		if err != nil {
			return nil, err
		}
		trailers = append(trailers, t)
	}

	return trailers, nil
}

// Vote applies a vote to a trailer.
func (m *TrailerModel) Vote(trailerID int, ipAddress string, voteValue int) error {
	// Using INSERT ... ON DUPLICATE KEY UPDATE to handle new votes and changes to existing votes
	stmt := `
        INSERT INTO votes (trailer_id, ip_address, vote_value)
        VALUES (?, ?, ?)
        ON DUPLICATE KEY UPDATE vote_value = ?
    `

	_, err := m.DB.Exec(stmt, trailerID, ipAddress, voteValue, voteValue)
	return err
}
