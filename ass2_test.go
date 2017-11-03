package ass2

import "fmt"
import "testing"
import "net/http"
import "bytes"
import "encoding/json"
import "net/http/httptest"
import "io/ioutil"

func Test_AccessAndDelete(t *testing.T) {
	//"STOLEN" from here: https://github.com/marni/imt2681_studentdb/blob/master/api_student_test.go#L12
	ts1 := httptest.NewServer(http.HandlerFunc(RegisterWebhookHandler))
	ts2 := httptest.NewServer(http.HandlerFunc(AccessWebhooksHandler))
	defer ts1.Close()
	defer ts2.Close()
	s := Webhook{"", "testing","EUR","NOK",float64(9.6),float64(6.9)}
	jsonBlob, err := json.Marshal(s)
	resp, err := http.Post(ts1.URL, "application/json", bytes.NewBuffer(jsonBlob))
	if err != nil {
		t.Errorf("Error doing http.Post(): %s", err)
	}
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error doing ioutil.ReadAll(): %s", err)
	}
	fmt.Println("ID is: " + string(id))
	
	resp, err = http.Get(ts2.URL + "/" + string(id))
	if err != nil {
		t.Errorf("Error doing http.Get(): %s", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	s2 := Webhook{}
	json.Unmarshal(data, &s2)
	fmt.Println(s2)
	
	req, err := http.NewRequest(http.MethodDelete, ts2.URL + "/" + string(id), nil)
	client := &http.Client{}
	client.Do(req)
}

func Test_Latest(t *testing.T) {
	ts2 := httptest.NewServer(http.HandlerFunc(LatestCurrencyHandler))
	defer ts2.Close()
	m := make(map[string]interface{})
	m["baseCurrency"] = "EUR"
	m["targetCurrency"] = "NOK"
	jsonBlob, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("JSON Error: %s\n", err)
		return
	}
	resp, err := http.Post(ts2.URL, "application/json", bytes.NewBuffer(jsonBlob))
	if err != nil {
		t.Errorf("Error doing http.Get(): %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error doing ioutil.ReadAll(): %s", err)
	}
	fmt.Println(string(body))
}

func Test_Average(t *testing.T) {
	ts2 := httptest.NewServer(http.HandlerFunc(AverageCurrecyHandler))
	defer ts2.Close()
	m := make(map[string]interface{})
	m["baseCurrency"] = "EUR"
	m["targetCurrency"] = "NOK"
	jsonBlob, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("JSON Error: %s\n", err)
		return
	}
	resp, err := http.Post(ts2.URL, "application/json", bytes.NewBuffer(jsonBlob))
	if err != nil {
		t.Errorf("Error doing http.Get(): %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error doing ioutil.ReadAll(): %s", err)
	}
	fmt.Println(string(body))
}

func Test_EvalTrigger(t *testing.T) {
	ts2 := httptest.NewServer(http.HandlerFunc(EvalTriggerHandler))
	defer ts2.Close()
	resp, err := http.Get(ts2.URL)
	if err != nil {
		t.Errorf("Error doing http.Get(): %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error doing ioutil.ReadAll(): %s", err)
	}
	fmt.Println(string(body))
}

func Test_getInitialDate(t *testing.T) {
	GetInitialDate()
}
