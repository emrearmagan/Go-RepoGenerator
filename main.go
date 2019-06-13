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
	ErrorColor = "\033[1;31m%s\033[0m\n"
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
	case answer := <-answerCh:
		answer = strings.ToLower(answer)
		if "y" == answer {
			clear()
			return nil
		} else if "n" == answer {
			return errors.New("")
		}
		return errors.New("")
	}
}

func main() {
	clear()
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

	config := GetConfig()
	absolutePath := path.Join(path.Dir(config.WorkingDirectory), repo)
	err := askPermission(absolutePath, repo)
	if err != nil {
		log.Fatal(err)
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
		c := NewCommander(config)
		clear()
		if err := c.gitInit(); err != nil {
			fmt.Println(err)
			return
		}

		if err := c.openIDE(); err != nil {
			log.Println(err)
		}
	}()

	err = createRepo(config, repo, done)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	fmt.Println("Done")
}

//Creates a Repository on Github using GoogleChrome
func createRepo(c *config, repo string, ch chan<- bool) error {
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	cdp, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}

	// run task list
	err = cdp.Run(ctxt, gitHub(c, repo))
	if err != nil {
		panic(err)
	}

	// shutdown chrome
	err = cdp.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = cdp.Wait()
	if err != nil {
		log.Fatal(err)
	}

	ch <- true
	return nil
}

func gitHub(c *config, repo string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(c.GitUrl + "/login"),
		chromedp.Sleep(2 * time.Second),
		//Fill Username, Password and Sign In
		chromedp.WaitVisible(`#login_field`, chromedp.ByID),
		chromedp.SendKeys(`#login_field`, c.Username, chromedp.ByID),
		chromedp.SendKeys(`#password`, c.Password, chromedp.ByID),
		//chromedp.Click(`input[type="submit"]`, chromedp.NodeVisible),
		chromedp.Click(`input[value="Sign in"]`, chromedp.BySearch),
		chromedp.Sleep(3 * time.Second),
		//chromedp.Sleep(5 * time.Second),
		chromedp.Navigate(c.GitUrl + "/new"),
		//// Create new Repository
		//chromedp.WaitVisible(`#new_repository`, chromedp.ByID),
		//chromedp.SendKeys(`#repository_name`, repo, chromedp.ByID),
		////chromedp.Sleep(1*time.Second),
		//chromedp.WaitEnabled(`button[class="btn btn-primary first-in-line"]`, chromedp.ByQuery),
		//chromedp.Click(`button[class="btn btn-primary first-in-line"]`, chromedp.ByQuery),
		//
		//
		//////Copy remote address
		//chromedp.WaitVisible(`clipboard-copy[for="empty-setup-push-repo-echo"]`, chromedp.ByQuery),
		//chromedp.Click(`clipboard-copy[for="empty-setup-push-repo-echo"]`, chromedp.ByQuery),
	}
}

//-------------------------------Commander------------

type Commander struct {
	cmd    *exec.Cmd
	config *config
}

func NewCommander(config *config) Commander {
	c := Commander{config: config}
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

//opens the repo with your given IDE
func (c *Commander) openIDE() error {
	c.cmd = exec.Command(c.config.IDE, ".")
	c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.cmd.Run(); err != nil {
		return errors.New("unable to openIDE directory with your Editor")
	}

	return nil
}

//Adds the copied remote server and push to master
func (c *Commander) addRemote() error {
	var remote []string
	var err error
	if remote, err = c.pasteCopy(); err != nil {
		return err
	}

	c.cmd = exec.Command("git", remote...)
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

	split := strings.Split(string(out), "\n")
	rm := strings.Split(split[0], " ")
	return rm[1:5], nil
}

//clear the Terminal
func clear() {
	cmd := exec.Command("clear")

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return
	}
}
