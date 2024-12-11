package bootstrap

import "gorm.io/gorm"

type Application struct {
	Env      *Env
	Postgres *gorm.DB
}

func App() (*Application, error) {
	app := &Application{}
	var err error
	app.Env, err = NewEnv()
	if err != nil {
		return nil, err
	}
	app.Postgres, err = NewPostgresDatabase(app.Env)
	if err != nil {
		return nil, err
	}
	return app, nil
}
