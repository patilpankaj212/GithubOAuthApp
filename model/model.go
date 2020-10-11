package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	userURL = "https://api.github.com/user"
)

//ErrorStruct defines the error response for authorization request
type ErrorStruct struct {
	ErrorMessage string
}

//OAuthTokenResponse represents a successful token response
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

//User contains all information and API's available for the user
type User struct {
	LoginName       string `json:"login"`
	URL             string `json:"url"`
	ReposURL        string `json:"repos_url"`
	Type            string `json:"type"`
	SiteAdmin       bool   `json:"side_admin"`
	Repos           []Repo
	RepoToClone     string
	IsAuthenticated bool
	OAuthToken      string
	ClonedLocation  string
}

//Repo is the struct to define one Repo
type Repo struct {
	RepoName     string `json:"name"`
	RepoFullName string `json:"full_name"`
	Description  string `json:"description"`
	IsPrivate    bool   `json:"private"`
	Owner        User   `json:"owner"`
	RepoURL      string `json:"url"`
	CloneURL     string `json:"clone_url"`
	Language     string `json:"language"`
}

//GetUserInfo gets the user information using the bearer token
func GetUserInfo(accessToken string) (User, error) {
	user := User{}

	client := http.Client{}
	authorizationHeaderValue := fmt.Sprintf("Bearer %s", accessToken)

	request, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return user, err
	}

	request.Header.Add("Authorization", authorizationHeaderValue)

	response, err := client.Do(request)
	if err != nil {
		return user, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}

	err = setRepos(&user, client)
	if err != nil {
		return user, err
	}

	return user, nil
}

func setRepos(user *User, client http.Client) error {
	request, err := http.NewRequest("GET", user.ReposURL, nil)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	repos := []Repo{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return err
	}

	user.Repos = repos
	return nil
}

//CloneGitRepo method will clone the url to local machine
func (user *User) CloneGitRepo(repourl string) (string, error) {
	cmd := exec.Command("git", "clone", repourl)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("repository " + repourl + " cloned successfully to location: " + workingDirectory)
	return workingDirectory, nil
}
