INSERT INTO trailers (movie_id, title, youtube_id) VALUES
((SELECT id FROM movies WHERE slug = 'my-new-page5454'), 'Trailer 1', 'dQw4w9WgXcQ'),
((SELECT id FROM movies WHERE slug = 'test-page'), 'Trailer 1', 'dQw4w9WgXcQ');
