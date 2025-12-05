CREATE TABLE IF NOT EXISTS votes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    trailer_id INT NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    vote_value TINYINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_trailer FOREIGN KEY (trailer_id) REFERENCES trailers(id) ON DELETE CASCADE,
    CONSTRAINT unique_ip_vote UNIQUE (ip_address, trailer_id)
);
