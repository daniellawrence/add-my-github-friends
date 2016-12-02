package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"bufio"
	"strconv"
)


type PublicKey struct {
	Id  int    `json:"id"`
	Key string `json:"key"`
}


func createSSHFilesAndDirectories(username string) (error) {
	new_user, err := user.Lookup(username)

	if err != nil {
		fmt.Printf("local user '%s' missing from this system", username)
		return err
	}

	uid, _ := strconv.Atoi(new_user.Uid)
	gid, _ := strconv.Atoi(new_user.Gid)

	ssh_directory := "/home/" + username + "/.ssh"
	authorized_keys := ssh_directory + "/authorized_keys" 

	_, err = os.Stat(ssh_directory)

	if err != nil {
		err = os.Mkdir(ssh_directory, 0700)
		if err != nil {
			fmt.Printf("mkdir failed; %s\n", err)
			return err
		}
	}

	keys, err := FetchKeysForUser(username)
	if err != nil {
		fmt.Printf("ERROR: Fetching Keys for User failed... %s\n", err)
		os.Exit(1)
	}

	f, _ := os.Create(authorized_keys)
	w := bufio.NewWriter(f)

	for _, user_public := range keys {
		w.WriteString(fmt.Sprintf("%s\n", user_public.Key))
	}	
	w.Flush()
	
	err = os.Chown(authorized_keys, uid, gid)
	err = os.Chown(ssh_directory, uid, gid)

	return nil
}


func addUser(username string) (error) {
	adduser_command := "useradd"
	adduser_args := []string{"-c", "'github user'", "-m", username}

	_, err := exec.Command(adduser_command, adduser_args...).Output()
	if err != nil {
		return err
	}
	return nil
}


func FetchKeysForUser(username string) ([]*PublicKey, error) {
	var keys []*PublicKey

	url := "http://api.github.com/users/" + username + "/keys"
	response, err := http.Get(url)
	if err != nil {
		return keys, err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return keys, err
	}

	json.Unmarshal([]byte(contents), &keys)

	return keys, nil

}


func main() {
	// username := "daniellawrence"

	numberArgs := len(os.Args)

	if numberArgs == 1 {
		fmt.Println("Usage: add-my-github-friends username [username ... username]\n")
		fmt.Println("Add a github user and public on to your system, allowing them to login.")
	}

	currentUser, _ := user.Current()
	if currentUser.Uid != "0" {
		fmt.Println("root needed, as we add users to the local system")
		os.Exit(1)
	}

	for _, value := range os.Args[1:] {
		fmt.Printf("Adding GitHub user '%s'\n", value)
		addGitHubUser(value)
	}
	
	os.Exit(0)
}


func addGitHubUser(username string) {

	err := addUser(username)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(2)
	}

	err = createSSHFilesAndDirectories(username)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(3)
	}


	os.Exit(0)
	
}
