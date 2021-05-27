package golang_tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func JsonDecode(data io.Reader) (map[string]interface{}, error) {
	var responseData map[string]interface{}

	dec := json.NewDecoder(data)
	dec.UseNumber()
	err := dec.Decode(&responseData)

	return responseData, err
}

func JsonDecodeArray(data io.Reader) ([]map[string]interface{}, error) {
	var responseData []map[string]interface{}

	dec := json.NewDecoder(data)
	dec.UseNumber()
	err := dec.Decode(&responseData)

	return responseData, err
}

func CheckRequiredParams(data map[string]interface{}, filter []string) error {
	var missingParams []string
	for _, filterKey := range filter {
		if strings.Contains(filterKey, "|") {
			dependParams := strings.Split(filterKey, "|")
			found := false
			for _, dependParamsKey := range dependParams {
				_, ok := data[dependParamsKey]
				if ok {
					found = true
					break
				}
			}
			if !found {
				missingParams = append(missingParams, filterKey)
			}
		} else {
			val, ok := data[filterKey]
			if !ok || val == nil {
				missingParams = append(missingParams, filterKey)
			}
		}
	}

	if len(missingParams) > 0 {
		return errors.New(fmt.Sprintf("Missing %s param", missingParams))
	}

	return nil
}

func ConvertMap(arr map[string][]string) map[string]interface{} {
	data := make(map[string]interface{})
	if arr != nil {
		for key, val := range arr {
			data[key] = val[0]
		}
	}
	return data
}

func ShowError(err error, message string, w http.ResponseWriter) {
	log.Printf("%s: %s\n", message, err)
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "application/json")

	var tmpMsg = err.Error()
	if len(tmpMsg) > 0 && tmpMsg[0] == '"' {
		tmpMsg = tmpMsg[1:]
	}
	if len(tmpMsg) > 0 && tmpMsg[len(tmpMsg)-1] == '"' {
		tmpMsg = tmpMsg[:len(tmpMsg)-1]
	}
	msg, err := json.Marshal(fmt.Sprintf("%s: %s", message, tmpMsg))
	if err != nil {
		return
	}
	w.Write(msg)
}

func GetHttpError(data io.Reader) error {
	body, _ := ioutil.ReadAll(data)
	return errors.New(string(body))
}

func WriteToFile(filename string, text string) error {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(text); err != nil {
		return err
	}
	return nil
}

func MakeRequestLog(r *http.Request) error {
	currentTime := time.Now()
	url := r.URL.Path
	method := r.Method

	var body []byte
	if method == "POST" {
		body, _ = ioutil.ReadAll(r.Body)
	}

	err := WriteToFile("api.log", "["+currentTime.Format("2006.01.02 15:04:05")+"] "+url+" "+method+" "+string(body)+"\n")
	if err != nil {
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		return err
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	return nil
}
