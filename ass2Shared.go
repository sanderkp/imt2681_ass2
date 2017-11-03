package ass2

import "fmt"
import "strings"
import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"
import "encoding/json"
import "net/http"
import "io"
import "io/ioutil"
import "bytes"
import "time"
import "strconv"

//URL ...
const URL = "mongodb://admin:admin@ds119685.mlab.com:19685/imt2681_ass2"
//Db ...
const Db = "imt2681_ass2"
//WebhookCollection ...
const WebhookCollection = "webhook"
//CurrencyCollection ...
const CurrencyCollection = "currency"

//REMEMBER: FOR MARSHALING TO WORK MEMBER NAMES MUST START WITH CAPITAL LETTER
//FUUUUUUU GOOOO@@@@@@@@@@@1!!!!!!!!!!!!!!!!!

//Webhook ...
type Webhook struct {
	ID bson.ObjectId `json:"-" bson:"_id,omitempty"`
	WebhookURL string `json:"webhookURL"`
	BaseCurrency string `json:"baseCurrency"`
	TargetCurrency string `json:"targetCurrency"`
	MinTriggerValue float64 `json:"minTriggerValue"`
	MaxTriggerValue float64 `json:"maxTriggerValue"`
}
//Currency ...
type Currency struct {
	Base string `json:"base"`
	Date string `json:"date"`
	Rates map[string]interface{} `json:"rates"`
}

//DbConnect ...
func DbConnect(url string, db string, col string) (*mgo.Session, *mgo.Collection){
	s := strings.Split(url, "/")
	dbName := s[len(s) - 1]
	fmt.Print("Connecting to DB: " + dbName + "\n")
	session, err := mgo.Dial(url)
	if err != nil {
		fmt.Printf("DB Error: %s\n", err)
		return nil, nil
	}
	fmt.Print("Establishing session to DB\n")
	c := session.DB(db).C(col)
	return session, c
}
//DbGetCount ...
func DbGetCount(c *mgo.Collection) int {
	count, err := c.Count()
	if err != nil {
		fmt.Printf("DB Error: %s", err)
		return 0
	}
	return count
}
//GetCurrencyRate ...
func GetCurrencyRate(curry Currency, s Webhook) float64 {
	var res float64
	if curry.Rates[s.TargetCurrency] != nil && curry.Rates[s.BaseCurrency] != nil {
		targ, ok := curry.Rates[s.TargetCurrency].(float64)
		base, ok := curry.Rates[s.BaseCurrency].(float64)
		if !ok {
			fmt.Println("Type Assertion Error, target or base currency is not of type float")
			return res
		}
		res =  targ / base
	} else {
		fmt.Printf("ERROR, Target or BaseCurrency was not found in rates, TargetCurrency: '%s' BaseCurrency: '%s'\n", s.TargetCurrency, s.BaseCurrency)
	}
	return res
}
//DoInvokeWebhook ...
func DoInvokeWebhook(currentRate float64, s Webhook) {
	m := make(map[string]interface{})
	m["baseCurrency"] = s.BaseCurrency
	m["targetCurrency"] = s.TargetCurrency
	m["currentRate"] = currentRate
	m["minTriggerValue"] = s.MinTriggerValue
	m["maxTriggerValue"] = s.MaxTriggerValue
	jsonBlob, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("JSON Error: %s\n", err)
		return
	}
	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewBuffer(jsonBlob))
	if err != nil {
		fmt.Printf("Http POST Error: %s\n", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("IO Error: %s\n", err)
		return
	}
	fmt.Printf("Response: StatusCode: %d %s\n", resp.StatusCode, body)
}
//InvokeWebhook ...
func InvokeWebhook(currentRate float64, s Webhook) {
	fmt.Println("Invoking trigger: " + s.WebhookURL)
	if currentRate < s.MinTriggerValue || currentRate > s.MaxTriggerValue {
		DoInvokeWebhook(currentRate, s)
	} else {
		fmt.Println("The webhook was not notified because the current rate is not outside the min/max trigger values")
	}
}
//TriggerWebhooks ...
func TriggerWebhooks(f func(float64, Webhook)) error {
	fmt.Println("Invoking triggers")
	s := []Webhook{}
	fmt.Print("Connecting to DB\n")
	session, err := mgo.Dial(URL)
	if err != nil {
		fmt.Printf("DB Error: %s\n", err)
		return err
	}
	defer session.Close()
	fmt.Print("Establishing session to DB\n")
	c1 := session.DB(Db).C(WebhookCollection)
	c2 := session.DB(Db).C(CurrencyCollection)
	if session != nil {
		latestCurry := Currency{}
		err = c1.Find(nil).All(&s)
		err = c2.Find(nil).Sort("-$natural").One(&latestCurry)
		fmt.Println(latestCurry.Date)
		if err != nil {
			fmt.Printf("DB Error: %s\n", err)
			return err
		}
		for i := 0; i < len(s); i++ {
			f(GetCurrencyRate(latestCurry, s[i]), s[i])
		}
	}
	return nil
}
//ReadHTTPBody ...
func ReadHTTPBody(r io.ReadCloser) []byte {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Printf("IO Error: %s\n", err)
		return nil
	}
	return body
}
//ReadEntirePage ...
func ReadEntirePage(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("HTTP Error: %s\n", err)
		return nil
	}
	defer resp.Body.Close()
	return ReadHTTPBody(resp.Body)
}
//GetInitialDate ...
func GetInitialDate() time.Time {
	fmt.Print("Connecting to DB\n")
	session, err := mgo.Dial(URL)
	latestCurrencyTime := time.Now().Local()
	if err != nil {
		fmt.Printf("DB Error: %s", err)
	} else {
		defer session.Close()
		fmt.Print("Establishing session to DB\n")
		c := session.DB(Db).C(CurrencyCollection)
		if err != nil {
			fmt.Printf("DB Error: %s\n", err)
		} else {
			count, err := c.Count()
			if err != nil {
				fmt.Printf("DB Error: %s", err)
			} else {
				if count != 0 {
					curry := Currency{}
					err = c.Find(nil).Sort("-$natural").One(&curry)
					if err != nil {
						fmt.Printf("Query Error: %s", err)
					} else {
						newTime,err := time.Parse("2006-01-02", curry.Date)
						if err != nil {
							fmt.Printf("Time Error: %s\n", err)
						}
						diff := latestCurrencyTime.Sub(newTime)
						if diff.Hours() >= 24 {
							return latestCurrencyTime
						}
						return newTime
					}
				}
			}
		}
	}
	return latestCurrencyTime
}
//RegisterWebhookHandler ...
func RegisterWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		data := ReadHTTPBody(r.Body)
		if data != nil {
			s := Webhook{}
			err := json.Unmarshal(data, &s)
			if err != nil {
				fmt.Printf("JSON Error: %s\n", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Println(s)
			if s.WebhookURL == "" || s.TargetCurrency == "" || s.BaseCurrency == "" {
				fmt.Println("WebhookURL, TargetCurrency or BaseCurrency is empty, this is not allowed")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			session, c := DbConnect(URL, Db, WebhookCollection)
			if session != nil && c != nil {
				defer session.Close()
				fmt.Print("Ensuring index\n")
				err = c.EnsureIndex(mgo.Index{Key: []string{"webhookurl", "basecurrency", "targetcurrency"},Unique: true,DropDups: false,Background: false,Sparse: false})
				if err != nil {
					fmt.Printf("DB Error: %s\n", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				fmt.Print("Inserting webhook data\n")
				s.ID = bson.NewObjectId()
				err = c.Insert(s)
				if err != nil {
					fmt.Printf("DB Error: %s", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				fmt.Print("Webhook data registered!\n")
				w.Write([]byte(s.ID.Hex()))
			} else {
				fmt.Printf("DB Error: session or collection is nil")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		} else {
			fmt.Printf("IO Error: data is nil")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		fmt.Println("Ilegal method")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
//AccessWebhooksHandler ...
func AccessWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	s := strings.Split(r.URL.Path, "/")
	id := s[len(s) - 1]
	session, c := DbConnect(URL, Db, WebhookCollection)
	if r.Method == http.MethodGet {
		if session != nil && c != nil {
			defer session.Close()
			fmt.Println("Accessing registered webhook")
			res := Webhook{}
			if bson.IsObjectIdHex(id) {
				err := c.Find(bson.M{"_id":bson.ObjectIdHex(id)}).One(&res)
				if err != nil {
					fmt.Printf("DB Error: %s\n", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				fmt.Println(res)
				json.NewEncoder(w).Encode(res)
			} else {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
			}
		}
	} else if r.Method == http.MethodDelete {
		fmt.Println("Deleting registered webhook")
		err := c.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
		if err != nil {
			fmt.Printf("DB Error: %s\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Println("Webhook deleted!")
	} else {
		fmt.Println("Ilegal method")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
//LatestCurrencyHandler ...
func LatestCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s := Webhook{}
		data := ReadHTTPBody(r.Body)
		if data != nil {
			err := json.Unmarshal(data, &s)
			if err != nil {
				fmt.Printf("JSON Error: %s\n", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Println(s)
			fmt.Print("Connecting to DB\n")
			session, c := DbConnect(URL, Db, CurrencyCollection)
			if session != nil && c != nil {
				defer session.Close()
				fmt.Print("Finding currency db\n")
				curry := Currency{}
				//https://stackoverflow.com/questions/38127583/get-last-inserted-element-from-mongodb-in-golang
				count := DbGetCount(c)
				if count == 0 {
					fmt.Println("DB Warning: Database is empty");
					return
				}
				var exchange float64
				if count > 0 {
					err = c.Find(nil).Skip(count-1).One(&curry)
					if err != nil {
						fmt.Printf("DB Error: %s", err)
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					exchange = GetCurrencyRate(curry, s)
					fmt.Print("Got currency data")
				}
				w.Write([]byte(strconv.FormatFloat(exchange, 'f', 6, 64)))
			} else {
				fmt.Printf("DB Error: session or collection is nil")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	} else {
		fmt.Println("Ilegal method")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
//AverageCurrecyHandler ...
func AverageCurrecyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s := Webhook{}
		data := ReadHTTPBody(r.Body)
		if data != nil {
			err := json.Unmarshal(data, &s)
			if err != nil {
				fmt.Printf("JSON Error: %s\n", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Println(s)
			session, c := DbConnect(URL, Db, CurrencyCollection)
			if session != nil && c != nil {
				defer session.Close()
			}
			var curry []Currency
			err = c.Find(nil).Sort("-$natural").Limit(3).All(&curry)
			avrg := 0.0
			count := len(curry)
			fmt.Printf("Number of days to averge: %d\n", count)
			for i := range curry {
				fmt.Println(curry[i].Date)
				avrg += GetCurrencyRate(curry[i], s)
			}
			if count > 0 {
				avrg /= float64(count)
			}
			fmt.Println(curry)
			w.Write([]byte(strconv.FormatFloat(avrg, 'f', 6, 64)))
		}
	} else {
		fmt.Println("Ilegal method")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
//EvalTriggerHandler ...
func EvalTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := TriggerWebhooks(InvokeWebhook)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("Ilegal method")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
