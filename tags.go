package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func ConfigureFollowedTags() {
	db := GetDatabase()
	fmt.Println("At the moment you follow the following tags:")
	// Get the tags that the user is following
	tags := GetFollowedTags(db)
	// Show the tags
	for _, tag := range tags {
		fmt.Print(tag, " ")
	}
	fmt.Println()
	fmt.Println("Enter the tags you want to follow, separated by spaces\nTo remove tags, prefix them with a minus sign: ")
	newtags := Readline()
	// Split the tags into an array and insert them into database DB
	// using table "followed_tags"
	tagArray := strings.Split(newtags, " ")
	for _, tag := range tagArray {
		tag = strings.Trim(tag, " \n")
		if !strings.HasPrefix(tag, "-") {
			_, err := db.Exec("INSERT INTO followed_tags(tag) VALUES(?)", tag)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err := db.Exec("DELETE FROM followed_tags WHERE tag=?", tag[1:])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func GetFollowedTags(db *sql.DB) []string {
	// Get the tags from the database
	rows, err := db.Query("SELECT tag FROM followed_tags")
	if err != nil {
		fmt.Println(err)
	}
	var tags []string
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			fmt.Println(err)
		}
		tags = append(tags, tag)
	}
	return tags
}
