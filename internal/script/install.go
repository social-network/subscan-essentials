package script

import (
	"database/sql"
	"fmt"
	"github.com/social-network/netscan/util"
	"io"
	"os"
)

func Install(conf string) {
	// create database
	func() {
		dbHost := util.GetEnv("POSTGRES_HOST", "127.0.0.1")
		dbUser := util.GetEnv("POSTGRES_USER", "everhusk")
		dbPass := util.GetEnv("POSTGRES_PASS", "")
		dbName := util.GetEnv("POSTGRES_DB", "substrate")
		dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = db.Close()
		}()
		q := fmt.Sprintf("SELECT 'CREATE DATABASE %s WITH ENCODING = `UTF8`' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '%s')", util.NetworkNode, util.NetworkNode)
		_, err = db.Exec(q)
		if err != nil {
			panic(err)
		}
		fmt.Println("Create database", util.NetworkNode, "Success!!!")

	}()

	// conf
	_ = fileCopy(fmt.Sprintf("%s/http.toml.example", conf), fmt.Sprintf("%s/http.toml", conf))
	_ = fileCopy(fmt.Sprintf("%s/postgres.toml.example", conf), fmt.Sprintf("%s/postgres.toml", conf))
	_ = fileCopy(fmt.Sprintf("%s/redis.toml.example", conf), fmt.Sprintf("%s/redis.toml", conf))

}

func fileCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
