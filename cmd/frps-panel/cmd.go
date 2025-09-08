package main

import (
	"database/sql"
	"errors"
	"fmt"
	"frps-panel/pkg/server"
	"frps-panel/pkg/server/controller"
	"frps-panel/pkg/server/model"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	gormysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const version = "2.0.0"

var (
	showVersion bool
	configFile  string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of frps-panel")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./frps-panel.toml", "config file of frps-panel")
}

var rootCmd = &cobra.Command{
	Use:   "frps-panel",
	Short: "frps-panel is the server plugin of frp to support multiple users.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			log.Println(version)
			return nil
		}
		executable, err := os.Executable()
		if err != nil {
			log.Printf("error get program path: %v", err)
			return err
		}
		rootDir := filepath.Dir(executable)

		config, tls, err := parseConfigFile(configFile)
		if err != nil {
			log.Printf("fail to start frps-panel : %v", err)
			return err
		}

		// Database initialization
		if config.Database.Enable {
			log.Println("Database is enabled, connecting to database...")
			dsn := config.Database.Dsn
			cfg, err := mysql.ParseDSN(dsn)
			if err != nil {
				log.Fatalf("failed to parse database DSN: %v", err)
			}
			dbName := cfg.DBName
			cfg.DBName = ""
			sqlDB, err := sql.Open("mysql", cfg.FormatDSN())
			if err != nil {
				log.Fatalf("failed to connect to database server: %v", err)
			}
			_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
			if err != nil {
				sqlDB.Close()
				log.Fatalf("failed to create database: %v", err)
			}
			sqlDB.Close()
			log.Printf("Database '%s' created or already exists.", dbName)
			db, err := gorm.Open(gormysql.Open(dsn), &gorm.Config{})
			if err != nil {
				log.Fatalf("failed to connect database: %v", err)
			}
			config.DB = db
			log.Println("Database connection successful.")

			// Auto migrate the schema
			err = db.AutoMigrate(&model.UserToken{}, &model.ServerInfo{})
			if err != nil {
				log.Fatalf("failed to auto migrate database schema: %v", err)
			}
			log.Println("Database schema migrated.")

			// Load tokens from database
			var userTokens []model.UserToken
			result := db.Find(&userTokens)
			if result.Error != nil {
				log.Fatalf("failed to load tokens from database: %v", result.Error)
			}

			config.Tokens = make(map[string]controller.TokenInfo)
			for _, ut := range userTokens {
				ti, err := controller.ToTokenInfo(ut)
				if err != nil {
					log.Printf("error converting user token %s to token info: %v", ut.User, err)
					continue
				}
				config.Tokens[ti.User] = ti
			}
			log.Printf("Loaded %d tokens from database.", len(config.Tokens))

		} else {
			log.Println("Database is disabled, using token file.")
			configDir := filepath.Dir(configFile)
			tokensFile := filepath.Join(configDir, "frps-tokens.toml")
			config.TokensFile = tokensFile

			var tokens controller.Tokens
			_, err = toml.DecodeFile(tokensFile, &tokens)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					tokens = controller.Tokens{Tokens: make(map[string]controller.TokenInfo)}
				} else {
					log.Fatalf("decode token file %v error: %v", tokensFile, err)
				}
			}
			config.Tokens = tokens.Tokens
			log.Printf("Loaded %d tokens from file.", len(config.Tokens))
		}

		s, err := server.New(
			rootDir,
			config,
			tls,
		)
		if err != nil {
			return err
		}
		err = s.Run()
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func parseConfigFile(configFile string) (controller.HandleController, server.TLS, error) {
	var commonCfg controller.Common
	_, err := toml.DecodeFile(configFile, &commonCfg)
	if err != nil {
		log.Fatalf("decode config file %v error: %v", configFile, err)
	}

	tls := server.TLS{
		Enable:   commonCfg.Common.TlsMode,
		Protocol: "HTTP",
		Cert:     commonCfg.Common.TlsCertFile,
		Key:      commonCfg.Common.TlsKeyFile,
	}

	if tls.Enable {
		tls.Protocol = "HTTPS"

		if strings.TrimSpace(tls.Cert) == "" || strings.TrimSpace(tls.Key) == "" {
			tls.Enable = false
			tls.Protocol = "HTTP"
			log.Printf("fail to enable tls: tls cert or key not exist, use http as default.")
		}
	}

	return controller.HandleController{
		CommonInfo:            commonCfg.Common,
		Version:               version,
		ConfigFile:            configFile,
		CurrentDashboardIndex: 0, // Default to the first dashboard
		Database:              commonCfg.Database,
	}, tls, nil
}
