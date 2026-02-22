package config

import "time"

type Config struct {
	Server   ServerConfig
	Matching MatchingConfig
	Quiz     QuizConfig
	Drop     DropConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type MatchingConfig struct {
	LifestyleWeight float64
	ValuesWeight    float64
	PoliticsWeight  float64
	LookbackWeeks   int
}

type QuizConfig struct {
	TotalQuestions int
	MinAge         int
}

type DropConfig struct {
	ExpirationHours int
}

func NewDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         ":8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Matching: MatchingConfig{
			LifestyleWeight: 0.35,
			ValuesWeight:    0.40,
			PoliticsWeight:  0.25,
			LookbackWeeks:   12,
		},
		Quiz: QuizConfig{
			TotalQuestions: 66,
			MinAge:         18,
		},
		Drop: DropConfig{
			ExpirationHours: 72,
		},
	}
}
