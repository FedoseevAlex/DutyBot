package main

import (
	"dutybot/internal/config"
	"flag"
	"log"
	"os"

	_ "dutybot/internal/migrations"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose"
)

var (
	flags      = flag.NewFlagSet("goose", flag.ExitOnError)
	configPath = flags.String("config", config.DefaultConfigPath, "path to config file")
)

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Print("Unable to parse flags: ", err)
		return
	}

	config.ReadConfig(*configPath)

	args := flags.Args()
	command := args[0]

	db, err := goose.OpenDBWithDriver(config.Cfg.DBDriver, config.Cfg.DBConnectString)
	if err != nil {
		log.Print("Cannot connect to database: ", err)
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal("Failed to close db", err)
		}
	}()

	err = goose.Run(command, db, ".")
	if err != nil {
		log.Print("Error running migrations: ", err)
		return
	}
}
