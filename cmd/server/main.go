package main

import (
	"log"

	"github.com/arthurlee945/Docrilla/config"
	"github.com/arthurlee945/Docrilla/db"
)

var cfg *config.Config

func init() {
	if err := config.Initialize(".env"); err != nil {
		log.Println("No .env file found")
	}
	cfg = config.New()
}

func main() {
	dbConn, err := db.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer dbConn.Close()
	// db.InitializeTable(dbConn)
	// dbConn.Exec(`
	// DROP Table IF EXISTS account, field, project, session, usr, verification_token;
	// DROP Type IF EXISTS user_role, role, project_type, type;
	// `)
}