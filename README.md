# GitHub Repo Generator
Creates a repository for you on github using Chromedp and opens it with your given IDE.

## Installation
The easiest way to use the HVV API in your Go project is to install it using **go get**:
```go
go get https://github.com/emrearmagan/Go-RepoGenerator
```
Before running, you should set the config values in **config.go**
```go
func GetConfig() *config {
	return &config{
		WorkingDirectory: "/Users/your/work/space/src/",
		GitUrl:           "https://github.com",
		Username:         "USERNAME",
		Password:         "PASSWORD",
		IDE:              "/usr/local/bin/goland",
	}
}
```

## Usage

* Go to your home directory using ```cd ~```
* Create a new bash script with touch ```.my_commands.sh```

Open the file you've just created in any text editor you like and paste the following code:

```
#!/bin/bash

#Creates a Git Repository
function init() {
  cd ~/path/to-your-work/space/Go-RepoGenerator
  go run main.go $1 
  cd ~/path/to-your-work/space/$1
}
```

* nano ~/.bash_profile and paste the following code: ```source ~/.my_commands.sh```
* Run ```init your_repository_name``` and it will automatically create and init a git repository for you :) 

## Author
Emre Armagan, emre.armagan@hotmail.de