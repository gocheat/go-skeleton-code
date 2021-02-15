package driver

import (
	"context"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.uber.org/fx"
	"skeleton-code/config"
	"skeleton-code/logger"
	"sync"
	"time"
)

const (
	DefaultMaxOpenConns = 25
	DefaultMaxIdleConns = 25
)

type Database struct {
	Db   *gorm.DB
	once sync.Once
}

func LoadDatabase(lifecycle fx.Lifecycle, cnf *config.Config) (*Database, error) {
	logger.Info("database connecting..", cnf.Database.Driver, cnf.Database.Source)
	db, err := gorm.Open(cnf.Database.Driver, cnf.Database.Source)
	//db.SetLogger(logger.Root())
	if err != nil {
		return nil, err
	}
	if cnf.Develop {
		db = db.Debug()
	}

	db = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8 auto_increment=1").Debug()

	if cnf.Database.MaxOpen <= 0 {
		cnf.Database.MaxOpen = DefaultMaxOpenConns
	}
	if cnf.Database.MaxIdle <= 0 {
		cnf.Database.MaxIdle = DefaultMaxIdleConns
	}
	db.DB().SetMaxOpenConns(cnf.Database.MaxOpen)
	db.DB().SetMaxIdleConns(cnf.Database.MaxIdle)
	db.DB().SetConnMaxLifetime(5 * time.Minute)

	database := &Database{
		Db: db,
	}
	lifecycle.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				database.Close()
				return nil
			},
		},
	)
	return database, nil
}

func (d *Database) Get(model interface{}) error {
	return d.Db.Where(model).First(model).Error
}

func (d *Database) Set(model interface{}) error {
	return d.Db.Create(model).Error
}

func (d *Database) Update(model interface{}) error {
	return d.Db.Model(model).Updates(model).Error
}

func (d *Database) Delete(model interface{}) error {
	return d.Db.Delete(model).Error
}

func (db *Database) Close() error {
	err := db.Db.Close()
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}