package bootstrap

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDatabase(env *Env) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
		env.DBHost, env.DBUser, env.DBPass, env.DBName, env.DBPort)
	// logger.Info(nil, "connecting to database")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// logger.Errorf(nil, "failed to connect to database: %v", err)
		return nil, err
	}
	return db, nil
}
