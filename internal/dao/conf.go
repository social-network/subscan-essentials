package dao

import (
	"fmt"
	"github.com/go-kratos/kratos/pkg/cache/redis"
	"github.com/go-kratos/kratos/pkg/database/sql"
	"github.com/social-network/netscan/util"
)

type (
	postgresConf struct {
		Conf struct {
			Host string
			User string
			Pass string
			DB   string
		}
		Api  *sql.Config
		Task *sql.Config
		Test *sql.Config
	}
	redisConf struct {
		Config *redis.Config
		DbName int
	}
)

func (dc *postgresConf) mergeEnvironment() {
	dbHost := util.GetEnv("POSTGRES_HOST", dc.Conf.Host)
	dbUser := util.GetEnv("POSTGRES_USER", dc.Conf.User)
	dbPass := util.GetEnv("POSTGRES_PASS", dc.Conf.Pass)
	dbName := util.GetEnv("POSTGRES_DB", dc.Conf.DB)
	dc.Api.DSN = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
	dc.Task.DSN = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
}

func (rc *redisConf) mergeEnvironment() {
	rc.Config.Addr = util.GetEnv("REDIS_ADDR", rc.Config.Addr)
	rc.DbName = util.StringToInt(util.GetEnv("REDIS_DATABASE", "0"))
}
