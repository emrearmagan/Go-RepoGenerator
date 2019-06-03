# Creates a Git Repository

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

