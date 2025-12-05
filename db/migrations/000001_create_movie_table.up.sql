-- 1. MOVIES
CREATE TABLE IF NOT EXISTS movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE, 
    release_year INT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. TRAILERS
CREATE TABLE IF NOT EXISTS trailers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    youtube_id VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_movie FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
);

-- 3. VOTES
-- Modified for IP tracking instead of User ID
CREATE TABLE IF NOT EXISTS votes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    trailer_id INT NOT NULL,
    
    -- IPv6 can be up to 45 chars. 
    -- We index this column because we will query it frequently to check if an IP has voted.
    ip_address VARCHAR(45) NOT NULL, 
    
    vote_value TINYINT NOT NULL, -- 1 for Upvote, -1 for Downvote - better to go with single column for Audits, preventing errs, and simplicity for us
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_trailer FOREIGN KEY (trailer_id) REFERENCES trailers(id) ON DELETE CASCADE,
    
    -- Composite Index/Unique constraint
    -- Ensures one IP can only vote on a specific trailer once.
    CONSTRAINT unique_ip_vote UNIQUE (ip_address, trailer_id)
);