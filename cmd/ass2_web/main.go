package main

import "fmt"
import "net/http"
import "time"
import "os"
import "ass2Shared"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Print("Starting server on port: " + port+ "\n")
	http.HandleFunc("/ex", ass2Shared.RegisterWebhookHandler)
	http.HandleFunc("/ex/", ass2Shared.AccessWebhooksHandler)
	http.HandleFunc("/ex/latest", ass2Shared.LatestCurrencyHandler)
	http.HandleFunc("/ex/average", ass2Shared.AverageCurrecyHandler)
	http.HandleFunc("/ex/evaluationtrigger", ass2Shared.EvalTriggerHandler)
	http.ListenAndServe(":" + port, nil)
}