package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	var currentDir = usr.HomeDir
	err = os.Chdir(usr.HomeDir)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan)
	go handleSigint(sigChan, usr)
	for {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Chdir(usr.HomeDir)
			fmt.Print(getDir(usr.HomeDir, usr))
		} else {
			fmt.Print(getDir(wd, usr))
		}

		input, err := reader.ReadString('\n')
		input = input[:len(input)-1]
		if err != nil {
			fmt.Println(err)
		}

		inputParts := strings.Split(input, " ")

		switch inputParts[0] {
		case "ls":
			if len(inputParts) > 1 {
				ls(currentDir, inputParts[1])
			} else {
				ls(currentDir, "")
			}
			break
		case "cd":
			changedDir, err := cd(inputParts[1], usr)
			if err != nil {
				fmt.Println(err)
				break
			}
			currentDir = changedDir
		case "exit":
			os.Exit(0)
		case "":
			break
		default:
			cmd := exec.Command(inputParts[0], inputParts[1:]...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func cd(newDir string, user *user.User) (string, error) {
	var dir string
	if newDir == "~" {
		dir = user.HomeDir
	} else {
		dir = newDir
	}
	err := os.Chdir(dir)
	if err != nil {
		return "", err
	}
	changedDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return changedDir, nil
}

func getDir(dir string, user *user.User) string {
	if dir == user.HomeDir {
		return "~ " + user.Username + "$ "
	}
	return dir + " " + user.Username + "$ "
}

func ls(currentDir string, flags string) {
	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		fmt.Println(err)
	}

	showHiddenFiles := strings.Contains(flags, "a")
	detailShow := strings.Contains(flags, "l")

	for i := 0; i < len(files); i++ {
		file := files[i]
		if string(file.Name()[0]) == "." && !showHiddenFiles {
			continue
		}
		if detailShow {
			fileTime := file.ModTime()

			fmt.Print(file.Mode().String() + " " + strconv.Itoa(int(file.Size())) + " " + fileTime.Format("2006-01-02 15:04") + " ")
		}
		fmt.Print(files[i].Name() + "\n")
	}
}

func handleSigint(sigChan chan os.Signal, usr *user.User) {
	for {
		<-sigChan
		signal := <-sigChan
		if signal.String() == "child exited" {
			continue
		}
		fmt.Fprint(os.Stdout, "\r \r")
		fmt.Print("                                                                                 ")
		fmt.Fprint(os.Stdout, "\r \r")
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Print(getDir(pwd, usr))
		fmt.Println()
		fmt.Print(getDir(pwd, usr))
	}
}
