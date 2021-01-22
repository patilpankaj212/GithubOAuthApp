package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/patilpankaj212/GitHubOAuthApp/model"
)

const (
	//Client ID and Client Secret for you app is required
	clientID     = ""
	clientSecret = ""

	//redirect URI as specified in you app
	redirectURI = ""
)

//using one user for the app for testing
var globalUser *model.User

//RegisterHandlers registers all http handlers
func RegisterHandlers() {

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/index", http.StatusTemporaryRedirect)
	})
	// http.Handle("/home", http.FileServer(http.Dir("html")))
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/authorize/", authorizationHandler)
	http.HandleFunc("/redirect/", redirectHandlerFunc)
	http.HandleFunc("/error/", errorHandlerFunc)
	http.HandleFunc("/success/", successHandlerFunc)
	http.HandleFunc("/clone/", cloneHandler)
	globalUser = &model.User{}
}

//indexHandler handles the root request
func indexHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Root Handler called")

	if globalUser.IsAuthenticated && !strings.EqualFold(globalUser.OAuthToken, "") {
		http.Redirect(writer, request, "/success/", http.StatusTemporaryRedirect)
	}
	templates := populateTemplates()
	requestedFile := request.URL.Path[1:]
	if strings.EqualFold(requestedFile, "") {
		requestedFile = "index"
	}
	template := templates.Lookup(requestedFile + ".html")
	if template != nil {
		err := template.Execute(writer, nil)
		if err != nil {
			log.Println(err)
		}
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

//authorizationHandler handles the authorization request
func authorizationHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Authorization Handler called")

	redirectURI := url.QueryEscape(redirectURI)
	authorizeURL := "https://github.com/login/oauth/authorize"
	scope := "repo"
	responseType := url.QueryEscape("code token")
	responseMode := "query"
	url := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s&response_type=%s&response_mode=%s", authorizeURL, clientID, redirectURI, scope, responseType, responseMode)
	http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
}

//redirectHandlerFunc redirects call to appropriate handler
func redirectHandlerFunc(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Redirect Handler called")

	requestURL := request.URL
	queryPrams := requestURL.Query()
	fmt.Println(queryPrams)
	code := queryPrams.Get("code")
	if strings.EqualFold(strings.TrimSpace(code), "") {
		//user has rejected the request. pass the request to error handler
		errorDescription := queryPrams.Get("error_description")
		errorURL := fmt.Sprintf("/error?error_description=%s", errorDescription)
		http.Redirect(writer, request, errorURL, http.StatusTemporaryRedirect)
	} else {
		//user had accepted the request. pass the request to success handler
		successURL := fmt.Sprintf("/success?code=%s", code)
		http.Redirect(writer, request, successURL, http.StatusTemporaryRedirect)
	}
}

//errorHandlerFunc handles error requests
func errorHandlerFunc(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Error Handler called")

	requestURL := request.URL
	queryPrams := requestURL.Query()
	errorDescription := queryPrams.Get("error_description")
	if strings.EqualFold(errorDescription, "") {
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
	} else {
		templates := populateTemplates()
		requestedFile := "error"
		template := templates.Lookup(requestedFile + ".html")
		if template != nil {
			errorMessage := model.ErrorStruct{ErrorMessage: errorDescription}
			err := template.Execute(writer, errorMessage)
			if err != nil {
				log.Println(err)
			}
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	}
}

//successHandlerFunc handles the successful code generate requests
func successHandlerFunc(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Success Handler called")

	if globalUser.IsAuthenticated && !strings.EqualFold(globalUser.OAuthToken, "") {
		templates := populateTemplates()
		requestedFile := "success"
		template := templates.Lookup(requestedFile + ".html")
		if template != nil {
			err := template.Execute(writer, globalUser)
			if err != nil {
				log.Println(err)
			}
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	} else {
		//go and retrieve the tokens
		requestURL := request.URL
		queryPrams := requestURL.Query()
		code := queryPrams.Get("code")
		if strings.EqualFold(code, "") {
			http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		} else {
			tokenResponse, err := getAccessToken(code)
			if err != nil {
				writer.Write([]byte(err.Error()))
				return
			}
			globalUser, err = getUser(tokenResponse.AccessToken)
			if err != nil {
				writer.Write([]byte(err.Error()))
				return
			}
			globalUser.IsAuthenticated = true
			globalUser.OAuthToken = tokenResponse.AccessToken

			templates := populateTemplates()
			requestedFile := "success"
			template := templates.Lookup(requestedFile + ".html")
			if template != nil {
				err := template.Execute(writer, globalUser)
				if err != nil {
					log.Println(err)
				}
			} else {
				writer.WriteHeader(http.StatusNotFound)
			}
		}
	}
}

//cloneHandler handles the authorization request
func cloneHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Clone Handler called")

	requestURL := request.URL
	queryPrams := requestURL.Query()
	fmt.Println(queryPrams)
	cloneURL := queryPrams.Get("url")
	cloneLocation, err := globalUser.CloneGitRepo(cloneURL)
	if err != nil {
		writer.Write([]byte(err.Error()))
		return
	}
	globalUser.RepoToClone = cloneURL
	globalUser.ClonedLocation = cloneLocation
	templates := populateTemplates()
	requestedFile := "clone"
	template := templates.Lookup(requestedFile + ".html")
	if template != nil {
		err := template.Execute(writer, globalUser)
		if err != nil {
			log.Println(err)
		}
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

//getAccessToken method accepts the authorization code and retrieves the oauth token
func getAccessToken(code string) (*model.OAuthTokenResponse, error) {
	tokenEndpointURL := "https://github.com/login/oauth/access_token"
	client := &http.Client{}
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:8000/redirect")

	req, err := http.NewRequest("POST", tokenEndpointURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	oauthToken := model.OAuthTokenResponse{}
	err = json.Unmarshal(body, &oauthToken)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &oauthToken, nil
}

func getUser(oauthToken string) (*model.User, error) {
	user, err := model.GetUserInfo(oauthToken)
	if err != nil {
		fmt.Println(err)
		return nil, http.ErrAbortHandler
	}
	return &user, nil
}

func populateTemplates() *template.Template {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(workingDirectory)
	result := template.New("html")
	template.Must(result.ParseGlob(workingDirectory + "/html/*.html"))
	return result
}
