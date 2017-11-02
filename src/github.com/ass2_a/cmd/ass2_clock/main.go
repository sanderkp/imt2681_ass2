package main

import "fmt"
import "encoding/json"
import "gopkg.in/mgo.v2"
import "time"
import "github.com/ass2"

//REMEMBER: FOR MARSHALING TO WORK MEMBER NAMES MUST START WITH CAPITAL LETTER
//FUUUUUUU GOOOO@@@@@@@@@@@1!!!!!!!!!!!!!!!!!

func getLatestCurrency() {
	fmt.Println("Connecting to DB")
	session, c := ass2Shared.DbConnect(ass2Shared.URL, ass2Shared.Db, ass2Shared.CurrencyCollection)
	defer session.Close()
	if session != nil && c != nil {
		//Get currency data
		fmt.Println("Getting latest currecy data")
		body := ass2Shared.ReadEntirePage("http://api.fixer.io/latest")
		if body != nil {
			curry := ass2Shared.Currency{}
			err := json.Unmarshal(body, &curry)
			if err != nil {
				fmt.Printf("JSON Error: %s\n", err)
			} else {
				fmt.Println("Got currency data! Now inserting into db")
				fmt.Println("Creating index")
				index := mgo.Index{Key: []string{"date"},Unique: true, DropDups: false, Background: false, Sparse: false}
				fmt.Println(index)
				fmt.Println("Ensuring index")
				err = c.EnsureIndex(index)
				fmt.Println("Index ensured")
				if err != nil {
					fmt.Printf("DB Error: %s\n", err)
				} else {
					fmt.Println("Trying to set base = 1.0")
					if curry.Rates != nil {
						fmt.Println("Base currency set")
						curry.Rates[curry.Base] = 1.0
					} else {
						fmt.Println("Failed to set base because curry.Rates is nil")
						return
					}
					err = c.Insert(curry)
					if err != nil {
						fmt.Printf("DB Error: %s", err)
					} else {
						fmt.Println("Currency data retrieved and data inserted into DB")
					}
					//TODO: Delete old currency data?
				}
			}
		}
	}
	//Check againts webhooks
	ass2Shared.TriggerWebhooks(ass2Shared.InvokeWebhook)
}

func main() {
	for {
		fmt.Println("Getting latest currency data")
		getLatestCurrency()
		time.Sleep(time.Minute * 20)
	}
}