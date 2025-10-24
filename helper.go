package main

//Meant to have one off stuff
import (
	"fmt"
	"regexp"
)

var illegalCharPattern = regexp.MustCompile(`[^a-z0-9_-]`) //our good dictionary

func charchecker(name string) error { //returns nil if no bad characters are found
	if illegalCharPattern.MatchString(name) {
		return fmt.Errorf("name contains bad characters")
	}
	return nil
}

// Regex to find a YouTube video ID from various URL formats.
var youtubeRegex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?(?:youtube\.com\/(?:watch\?v=|embed\/)|youtu\.be\/)([a-zA-Z0-9\-_]+)`)

// processYouTubeURL finds the first YouTube URL in a block of text
// and converts it to a standard embeddable URL.
// If no URL is found, it returns an empty string.
func processYouTubeURL(text string) string {
	matches := youtubeRegex.FindStringSubmatch(text)

	// matches[0] is the full matched URL, matches[1] is the video ID (the capturing group)
	if len(matches) > 1 {
		videoID := matches[1]
		return "https://www.youtube.com/embed/" + videoID
	}

	return ""
}
