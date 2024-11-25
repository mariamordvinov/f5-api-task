package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"f5.com/ha/logger"
	api_sec "f5.com/ha/pkg"
)

func registerHandlers() {
	http.HandleFunc("/register", logger.LogHandler(api_sec.Register))
	http.HandleFunc("/login", logger.LogHandler(api_sec.Login))
	http.HandleFunc("/accounts", logger.LogHandler(api_sec.Auth(api_sec.AccountsHandler)))
	http.HandleFunc("/balance", logger.LogHandler((api_sec.Auth(api_sec.BalanceHandler))))
	http.HandleFunc("/users", logger.LogHandler((api_sec.Auth(api_sec.GetUseres))))

}

func main() {
	//setup the request/response logger
	err := logger.InitLogger()
	if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

	registerHandlers()

	err = http.ListenAndServe(":8080", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
