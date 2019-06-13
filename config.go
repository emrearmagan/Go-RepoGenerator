package main

type config struct {
	WorkingDirectory string
	GitUrl           string
	Username         string
	Password         string
	IDE              string
}

func GetConfig() *config {
	return &config{
		WorkingDirectory: "/Users/your/work/space/src/",
		GitUrl:           "https://github.com",
		Username:         "USERNAME",
		Password:         "PASSWORD",
		IDE:              "/usr/local/bin/goland",
	}
}
