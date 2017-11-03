package main

import "fmt"
import "net/http"
import "os"
import "github.com/sanderkp/imt2681_ass2"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Print("Starting server on port: " + port+ "\n")
	http.HandleFunc("/ex", ass2.RegisterWebhookHandler)
	http.HandleFunc("/ex/", ass2.AccessWebhooksHandler)
	http.HandleFunc("/ex/latest", ass2.LatestCurrencyHandler)
	http.HandleFunc("/ex/average", ass2.AverageCurrecyHandler)
	http.HandleFunc("/ex/evaluationtrigger", ass2.EvalTriggerHandler)
	http.ListenAndServe(":" + port, nil)
}