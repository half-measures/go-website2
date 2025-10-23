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
