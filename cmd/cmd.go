package cmd

import (
	"fmt"
	"wpbrute/pkg/config"
	"wpbrute/pkg/models"

	cli "github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config is the global configuration object
var Config *config.Config
var DBConnection *gorm.DB

// BeforeCommand sets the global flags before any commands are run
func BeforeCommand(c *cli.Context) error {
	// Load Config
	Config = loadConfigFromClientContext(c)

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s port=%d host=%s sslmode=require", Config.DBConfig.Username,
		Config.DBConfig.Password, Config.DBConfig.Database, Config.DBConfig.Port, Config.DBConfig.Host)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
		// This should probably be a config option
		// PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return err
	}
	DBConnection = db
	mdb, err := DBConnection.DB()
	if err != nil {
		panic(err)
	}
	mdb.SetMaxOpenConns(10)

	fmt.Println("Connected!")

	shouldMigrate := false
	if c.Bool("migrate") {
		shouldMigrate = c.Bool("migrate")
	}

	if shouldMigrate {
		fmt.Println("Migrating targets...")
		err = DBConnection.AutoMigrate(&models.Target{})
		if err != nil {
			return err
		}

		fmt.Println("Migrating creds...")
		err = DBConnection.AutoMigrate(&models.WordpressCredential{})
		if err != nil {
			return err
		}

		fmt.Println("Migrating proxies...")
		err = DBConnection.AutoMigrate(&models.Proxy{})
		if err != nil {
			return err
		}
	}

	return nil
}
func AfterCommand(c *cli.Context) error {
	// if DBConnection != nil {
	// 	return DBConnection.Close()
	// }
	return nil
}

func loadConfigFromClientContext(c *cli.Context) *config.Config {
	path := c.String("config")
	return config.NewFromYaml(path)
}
