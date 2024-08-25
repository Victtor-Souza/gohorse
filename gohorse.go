package gohorse

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"
)

type GoHorse struct {
	server    *echo.Echo
	container *dig.Container
}

func New() *GoHorse {
	gh := &GoHorse{
		server:    echo.New(),
		container: dig.New(),
	}
	gh.container.Invoke(initializeViper)
	return gh
}

func (gh *GoHorse) RegisterRepository(constructor interface{}) {
	err := gh.container.Provide(constructor)
	if err != nil {
		log.Panic(err)
	}
}

func (gh *GoHorse) RegisterApplication(application interface{}) {
	err := gh.container.Provide(application)
	if err != nil {
		log.Panic(err)
	}
}

func (gh *GoHorse) RegisterController(controller interface{}) {
	err := gh.container.Invoke(controller)
	if err != nil {
		log.Panic(err)
	}
}

func initializeViper() *viper.Viper {
	v := viper.New()
	v.AddConfigPath("./configs")
	v.SetConfigType("json")
	v.SetConfigName(os.Getenv("env"))
	if err := v.ReadInConfig(); err != nil {
		log.Panic(err)
	}
	return v
}

func (gh *GoHorse) registerMongoDB(host, user, password, database string, configs ...MongoDBConfig) {

	opts := options.Client().ApplyURI(host)

	for _, cfg := range configs {
		cfg(opts)
	}

	if user != "" {
		opts.SetAuth(options.Credential{Username: user, Password: password})
	}

	err := gh.container.Provide(func() *mongo.Database {
		cli, err := newMongoClient(opts)
		if err != nil {
			return nil
		}
		return cli.Database(database)
	})

	if err != nil {
		return
	}

}
