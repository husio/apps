package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	loginFl := flag.String("login", "", "Account login. Required.")
	passFl := flag.String("password", "", "Account password, Required.")
	roleFl := flag.String("role", "", "Account role, Required.")
	pgFl := flag.String("postgres", "", "PostgreSQL connection string.")
	hashCostFl := flag.Int("hash-cost", bcrypt.DefaultCost+2, "Password hash cost.")
	flag.Parse()

	if *loginFl == "" {
		printUsage()
		os.Exit(2)
	}
	if *loginFl == "" {
		printUsage()
		os.Exit(2)
	}
	if *passFl == "" {
		printUsage()
		os.Exit(2)
	}
	if *roleFl == "" {
		printUsage()
		os.Exit(2)
	}

	pg := os.Getenv("POSTGRES")
	if *pgFl != "" {
		pg = *pgFl
	}
	if pg == "" {
		pg = "host=localhost port=5432 user=postgres dbname=postgres sslmode=disable"
	}

	db, err := sql.Open("postgres", pg)
	if err != nil {
		log.Fatalf("cannot create database: %s", err)
	}
	defer db.Close()

	passHash, err := bcrypt.GenerateFromPassword([]byte(*passFl), *hashCostFl)
	if err != nil {
		log.Fatalf("cannot hash password: %s", err)
	}

	now := time.Now()

	row := db.QueryRow(`
		INSERT INTO accounts (login, password_hash, role, created_at, valid_till)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, valid_till
	`, *loginFl, string(passHash), *roleFl, now, now.Add(24*90*time.Hour))

	var (
		accID     int64
		validTill time.Time
	)
	if err := row.Scan(&accID, &validTill); err != nil {
		log.Fatalf("cannot create account: %s", err)
	}

	log.Printf("account registered %d, valid till %s", accID, validTill)
}

func printUsage() {
	fmt.Fprintf(os.Stdout, `
Usage: %s [options]

`, os.Args[0])
	flag.PrintDefaults()
}
