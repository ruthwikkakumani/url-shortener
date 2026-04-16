package logger

import (
	"go.uber.org/zap"
)


const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

func InitLogger(env string) (*zap.Logger, error){
	
	switch env{
		case EnvDevelopment :
			return zap.NewDevelopment()
			
		case EnvProduction :
			return zap.NewProduction()
			
		default :
			logger, err := zap.NewProduction()
			if err != nil{
				return nil, err
			}
			
			logger.Warn("Unknown environment, defaulting to production",
				zap.String("env", env),
			)
			
			return logger, nil
	}
}