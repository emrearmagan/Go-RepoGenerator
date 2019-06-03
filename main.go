package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	ErrorColor       = "\033[1;31m%s\033[0m\n"
	GitUrl           = "https://github.com/login"

	workingDirectory = "/Users/your/work/space/src/"
	Username         = "USERNAME"
	Password         = "PASSWORD"
	GoLand = "/usr/local/bin/goland" //need to create a command with Intellj - GoLand OR use your favourite IDE :)
)
func askPermission(path, repo string) error {
	//Check if directory already exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf(ErrorColor, "Repository/Directory already exists")
	}

	//starting a timer, when no input exit programm
	timer := time.NewTicker(time.Duration(10) * time.Second)

	fmt.Printf(ErrorColor, "Do you want to create a new Git Repository: "+repo+" ? (y/n)")
	fmt.Println(strings.Repeat("-", 30))

	answerCh := make(chan string)
	go func() {
		var answer string
		fmt.Scanf("%s\n", &answer)
		answerCh <- answer
	}()
	select {
	// time out
	case <-timer.C:
		return fmt.Errorf(ErrorColor, "No answer. Time out")
		//revieved an answer
	case answer := <-answerCh:
		answer = strings.ToLower(answer)
		if "y" == answer {
			return nil
		} else if "n" == answer {
			return errors.New("")
		}
		return errors.New("")
	}
}

func main() {
	//Checks if git is installed ?
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Printf(ErrorColor, "Git $Path not found")
		os.Exit(0)
	}

	input := os.Args[1:]
	if len(input) < 1 {
		fmt.Printf(ErrorColor, "Select Repository name")
		os.Exit(0)
	}
	repo := strings.Join(input, " ")

	absolutePath := path.Join(path.Dir(workingDirectory), repo)
	err := askPermission(absolutePath, repo)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	var wg sync.WaitGroup
	done := make(chan bool)
	//create directory and move there
	go func() {
		wg.Add(1)
		defer wg.Done()

		err := os.Mkdir(absolutePath, 0755)
		if err != nil {
			log.Fatal(err)
			return
		}

		//Change directory
		if err := os.Chdir(absolutePath); err != nil {
			log.Fatal(err)
			return
		}

		//Create README file
		if _, err := os.Create("README.md"); err != nil {
			fmt.Printf(ErrorColor, "README.md failed to create")
		}

		//wait for channel
		<-done
		clear()
		//Init Git Repository and add remote
		c := NewCommander()
		clear()
		if err := c.gitInit(); err != nil {
			fmt.Println(err)
			return
		}

		if err := c.openIDE();err !=nil{
			log.Println(err)
		}
	}()

	time.Sleep(1*time.Second)
	createRepo(repo, done)
	wg.Wait()
	fmt.Println("Done")
}

//Creates a Repository on Github using GoogleChrome
func createRepo(repo string, ch chan<- bool) error {
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}

	// run task list
	var res string
	err = c.Run(ctxt, gitHub(Username, Password, repo, &res))
	if err != nil {
		log.Fatal(err)
	}

	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	ch <- true
	return nil
}

func gitHub(u, p, repo string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(GitUrl),
		chromedp.Sleep(2 * time.Second),
		//Fill Username, Password and Sign In
		chromedp.WaitVisible(`#login_field`, chromedp.ByID),
		chromedp.SendKeys(`#login_field`, u, chromedp.ByID),
		chromedp.SendKeys(`#password`, p, chromedp.ByID),
		chromedp.Click(`input[name="commit"]`, chromedp.ByQuery),

		// Create new Repository
		chromedp.WaitVisible(`a[class="btn btn-sm btn-primary text-white"]`, chromedp.ByQuery),
		chromedp.Click(`a[class="btn btn-sm btn-primary text-white"]`, chromedp.ByQuery),
		//chromedp.Navigate("https://github.com/new"),
		chromedp.WaitVisible(`#new_repository`, chromedp.ByID),
		chromedp.SendKeys(`#repository_name`, repo, chromedp.ByID),
		//chromedp.Sleep(1*time.Second),
		chromedp.WaitEnabled(`button[class="btn btn-primary first-in-line"]`, chromedp.ByQuery),
		chromedp.Click(`button[class="btn btn-primary first-in-line"]`, chromedp.ByQuery),


		//Copy remote address
		chromedp.WaitVisible(`clipboard-copy[for="empty-setup-push-repo-echo"]`, chromedp.ByQuery),
		chromedp.Click(`clipboard-copy[for="empty-setup-push-repo-echo"]`, chromedp.ByQuery),
	}
}

//-------------------------------Commander------------

type Commander struct {
	cmd *exec.Cmd
}

func NewCommander() Commander {
	c := Commander{}
	return c
}

//Inits a new git repo
func (c *Commander) gitInit() error {
	c.cmd = exec.Command("git", "init")
	c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.cmd.Run(); err != nil {
		return err
	}

	if err := c.addRemote(); err != nil {
		return err
	}

	return nil
}

func (c *Commander) openIDE() error {
	c.cmd = exec.Command(GoLand, ".")
	c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.cmd.Run(); err != nil {
		return errors.New("unable to openIDE directory with your Editor")
	}

	return nil
}

//Adds the copied remote server and pushes to master
func (c *Commander) addRemote() error {
	var remote []string
	var err error
	if remote, err = c.pasteCopy(); err != nil {
		return err
	}

	c.cmd = exec.Command("git",remote...)
	c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = c.cmd.Run(); err != nil {
		return err
	}

	fmt.Println("Remote Server added")
	return nil
}

//Gets the copied value -remote...
func (c *Commander) pasteCopy() ([]string, error) {
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return nil, err
	}

	split := strings.Split(string(out),"\n")
	rm := strings.Split(split[0]," ")
	return rm[1:5],nil
}

func clear(){
	cmd := exec.Command("clear")

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return
	}
}