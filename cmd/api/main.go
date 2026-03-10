// File: cmd/api/main.go

package main 

import (
	"flag"
	"log/slog"
	"os"
	"time"
	"math/rand"
	"github.com/spector-asael/banking/cmd/api/handler"
	"github.com/spector-asael/banking/internal/data"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
)

func openDB(settings handler.ServerConfig) (*sql.DB, error) {
    // open a connection pool
    db, err := sql.Open("postgres", settings.DB.DSN)
    if err != nil {
        return nil, err
    }
    
    // set a context to ensure DB operations don't take too long
    ctx, cancel := context.WithTimeout(context.Background(),
                                       5 * time.Second)
    defer cancel()
    // let's test if the connection pool was created
    // we trying pinging it with a 5-second timeout
    err = db.PingContext(ctx)
    if err != nil {
        db.Close()
        return nil, err
    }


    // return the connection pool (sql.DB)
    return db, nil

} 

func main() {
	rand.Seed(time.Now().UnixNano())
	var settings handler.ServerConfig 

	flag.IntVar(&settings.Port, "port", 4000, "Server port")
	flag.StringVar(&settings.Environment, "env", "development", "Environment(development|staging|production)")
	flag.StringVar(&settings.DB.DSN, "db-dsn", "", "PostgreSQL DSN")
    flag.Float64Var(&settings.Limiter.RPS, "limiter-rps", 2,
                  "Rate Limiter maximum requests per second")

    flag.IntVar(&settings.Limiter.Burst, "limiter-burst", 5,
                  "Rate Limiter maximum burst")

    flag.BoolVar(&settings.Limiter.Enabled, "limiter-enabled", true,
                  "Enable rate limiter")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// release the database resources before exiting
	// defer db.Close()
	defer db.Close()

	logger.Info("database connection pool established")


	appInstance := &handler.ApplicationDependencies {
		Config: settings,
		Logger: logger,
		Models: data.Models{
		Deposits: data.DepositModel{DB: db},
		Balances: data.BalanceModel{DB: db},
		History: data.HistoryModel{DB: db},
	},
	}

    err = appInstance.Serve(&settings, appInstance)
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }

}