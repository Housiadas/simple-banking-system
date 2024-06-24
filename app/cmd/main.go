// This program performs background tasks
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"

	"github.com/Housiadas/backend-system/app/cmd/commands"
	cfg "github.com/Housiadas/backend-system/business/config"
	"github.com/Housiadas/backend-system/business/data/sqldb"
	"github.com/Housiadas/backend-system/foundation/logger"
)

var build = "develop"

type Config struct {
	DB      cfg.DB
	Version cfg.Version
	Auth    struct {
		KeysFolder string
		DefaultKID string
	}
	Kafka cfg.Kafka
}

func main() {
	log := logger.New(io.Discard, logger.LevelInfo, "CMD", func(context.Context) string {
		return "00000000-0000-0000-0000-000000000000"
	})
	if err := run(log); err != nil {
		if !errors.Is(err, commands.ErrHelp) {
			fmt.Println("msg", err)
		}
		os.Exit(1)
	}
}

func run(log *logger.Logger) error {
	c, err := LoadConfig("../../")
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}
	c.Version = cfg.Version{
		Build: build,
		Desc:  "CMD",
	}
	c.Auth = struct {
		KeysFolder string
		DefaultKID string
	}{
		KeysFolder: "/keys",
		DefaultKID: "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1",
	}

	return processCommands(os.Args, log, c)
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (cfg Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigFile("config.yaml")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}

// processCommands handles the execution of the commands specified on
// the command line.
func processCommands(args []string, log *logger.Logger, c Config) error {
	dbConfig := sqldb.Config{
		User:         c.DB.User,
		Password:     c.DB.Password,
		Host:         c.DB.Host,
		Name:         c.DB.Name,
		MaxIdleConns: c.DB.MaxIdleConns,
		MaxOpenConns: c.DB.MaxOpenConns,
		DisableTLS:   c.DB.DisableTLS,
	}

	switch args[1] {
	case "seed":
		if err := commands.Seed(dbConfig); err != nil {
			return fmt.Errorf("seeding database: %w", err)
		}

	case "useradd":
		name := args[2]
		email := args[3]
		password := args[4]
		if err := commands.UserAdd(log, dbConfig, name, email, password); err != nil {
			return fmt.Errorf("adding user: %w", err)
		}

	case "genkey":
		if err := commands.GenKey(); err != nil {
			return fmt.Errorf("key generation: %w", err)
		}

	case "userevents":
		if err := commands.UserEvents(log, dbConfig, c.Kafka); err != nil {
			return fmt.Errorf("kafka consumer for user events: %w", err)
		}

	default:
		fmt.Println("seed:       add data to the database")
		fmt.Println("useradd:    add a new user to the database")
		fmt.Println("genkey:     generate a set of private/public key files")
		fmt.Println("gentoken:	 generate a JWT for a user with claims")
		fmt.Println("userevents: kafka consumer to listen to user events")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
