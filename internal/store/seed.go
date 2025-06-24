package store

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"math/rand"
)

var usernames = []string{
	"alice", "john", "emma", "mike", "olivia", "david", "sophia", "james", "mia", "daniel",
	"ava", "william", "isabella", "ben", "amelia", "lucas", "harper", "logan", "ella", "jack",
	"grace", "liam", "chloe", "noah", "zoe", "ethan", "lily", "alex", "nora", "ryan",
	"ruby", "nathan", "leah", "tyler", "hannah", "cole", "sara", "kevin", "ivy", "brian",
	"elena", "jake", "aurora", "sean", "violet", "mark", "bella", "josh", "maya", "nick",
}

var titles = []string{
	"Getting Started with Go",
	"Why REST APIs Matter",
	"Top 10 VSCode Extensions",
	"How to Write Clean Code",
	"Understanding Concurrency in Go",
	"Deploying with Docker",
	"Writing Unit Tests in Go",
	"Exploring SQL Joins",
	"Building a Blog API",
	"JWT Authentication Explained",
	"Working with JSON in Go",
	"Pagination in REST APIs",
	"Debugging Tips for Go Developers",
	"Intro to Gorilla Mux",
	"Database Migrations with Goose",
	"Creating Middleware in Go",
	"Logging Best Practices",
	"Error Handling Strategies",
	"Working with Context in Go",
	"Serving Static Files with Go",
}

var texts = []string{
	"This post walks you through setting up your first Go project, from installation to running a simple program.",
	"Learn why RESTful APIs are the standard for web services and how they help build scalable systems.",
	"A curated list of must-have VSCode extensions that improve productivity for Go and web developers.",
	"Clean code is maintainable code. Here are tips and examples on how to write cleaner Go code.",
	"Go's concurrency model using goroutines and channels is powerful — this post explains how to use it effectively.",
	"See how Docker can simplify your deployment process and help you ship faster.",
	"Learn how to write unit tests in Go using the `testing` package and best testing practices.",
	"This post explains the differences between INNER, LEFT, RIGHT, and FULL joins in SQL with practical examples.",
	"A step-by-step guide to building a simple blog backend with Go and PostgreSQL.",
	"Understand how JWT tokens work, and how to implement authentication and authorization with them.",
	"Learn how to encode and decode JSON data in Go using the `encoding/json` package.",
	"Implementing pagination in your API can improve performance and usability — here’s how to do it.",
	"Struggling to debug Go code? Here are some tools and techniques that can make the process easier.",
	"A beginner’s intro to Gorilla Mux — a powerful router for building flexible HTTP servers in Go.",
	"Database migrations help manage schema changes. This post shows how to use Goose with Go projects.",
	"Middleware allows you to handle common concerns like logging and authentication in one place. Here’s how to write your own.",
	"Discover logging libraries and learn how to structure your logs for better observability.",
	"Good error handling is critical. This guide walks through idiomatic error patterns in Go.",
	"Learn how to use context for managing request-scoped values, cancellations, and timeouts in Go.",
	"Serve static files like images, CSS, and JavaScript efficiently using Go’s standard library.",
}

var tags = []string{
	"golang",
	"web",
	"api",
	"docker",
	"database",
	"sql",
	"rest",
	"json",
	"testing",
	"jwt",
	"concurrency",
	"security",
	"deployment",
	"clean-code",
	"middleware",
	"postgres",
	"router",
	"debugging",
	"context",
	"logging",
}

func (s *Store) Seed(store *Store, db *sqlx.DB) {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTxx(ctx, nil)

	for _, user := range users {
		if err := store.Users.CreateUser(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("error creating user:", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Errorf("error occured: %w", err)
		return
	}
	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.CreatePost(ctx, post); err != nil {
			log.Println("error creating post:", err)
			return
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("error creating comment:", err)
			return
		}
	}

	log.Println("Success seeding the DB!")
}

func generateUsers(num int) []*User {
	users := make([]*User, num)

	for i := 0; i < num; i++ {
		users[i] = &User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@test.com",
		}
	}

	return users
}

func generatePosts(num int, users []*User) []*Post {
	posts := make([]*Post, num)

	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]

		posts[i] = &Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: texts[rand.Intn(len(texts))],
			Tags: []string{
				titles[rand.Intn(len(tags))],
				titles[rand.Intn(len(tags))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*User, posts []*Post) []*Comment {
	cms := make([]*Comment, num)
	for i := 0; i < num; i++ {
		cms[i] = &Comment{
			PostID:  posts[rand.Intn(len(posts))].ID,
			UserID:  users[rand.Intn(len(users))].ID,
			Content: texts[rand.Intn(len(texts))],
		}
	}

	return cms
}
