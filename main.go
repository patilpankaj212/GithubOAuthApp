package main

import (
	"fmt"
	"net/http"

	"github.com/patilpankaj212/GoTest/GitHubOAuthApp/controller"
)

func main() {
	controller.RegisterHandlers()
	fmt.Println("Starting server on port 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		fmt.Println("Server didn't start due to error: ", err.Error())
		return
	}
}
