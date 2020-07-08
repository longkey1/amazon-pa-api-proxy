package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	paapi5 "github.com/spiegel-im-spiegel/pa-api"
	"github.com/spiegel-im-spiegel/pa-api/query"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	name = "amazon-pa-api-proxy"
	version = "0.8.2"
	retryKey = "retry"
)

// Config ...
type Config struct {
	Port                          int    `default:"1323"`
	AmazonAssociateTag            string `required:"true" split_words:"true"`
	AmazonAccessKey               string `required:"true" split_words:"true"`
	AmazonSecretKey               string `required:"true" split_words:"true"`
	AmazonLocale                  string `required:"true" split_words:"true"`
	AmazonRetryNumber             int    `default:"3" split_words:"true"`
	AmazonRequestDelayMillisecond int    `default:"1000" split_words:"true"`
}

// Response ...
type ErrorResponse struct {
	Errors []Error
}
type Error struct {
	Code string
	Message string
}

// localeMap
var localeMap = map[string]paapi5.Marketplace{
	"Australia":          paapi5.LocaleAustralia,
	"Brazil":             paapi5.LocaleBrazil,
	"Canada":             paapi5.LocaleCanada,
	"France":             paapi5.LocaleFrance,
	"Germany":            paapi5.LocaleGermany,
	"India":              paapi5.LocaleIndia,
	"Italy":              paapi5.LocaleItaly,
	"Japan":              paapi5.LocaleJapan,
	"Mexico":             paapi5.LocaleMexico,
	"Spain":              paapi5.LocaleSpain,
	"Turkey":             paapi5.LocaleTurkey,
	"UnitedArabEmirates": paapi5.LocaleUnitedArabEmirates,
	"UnitedKingdom":      paapi5.LocaleUnitedKingdom,
	"UnitedStates":       paapi5.LocaleUnitedStates,
}

var conf = &Config{}
var mutex = &sync.Mutex{}

func main() {
	checkVersion()
	loadConfig()

	log.Printf("%s started on http://localhost:%d", name, conf.Port)
	http.HandleFunc("/items/", getItems)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil))
}

func checkVersion() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("%s version %s\n", name, version)
		os.Exit(0)
	}
}

func loadConfig() {
	err := envconfig.Process("apap", conf)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getItems(w http.ResponseWriter, r *http.Request) {
	retry := 0
	if r.Context().Value(retryKey) != nil {
		retry = r.Context().Value(retryKey).(int)
	}

	asin := strings.TrimPrefix(r.URL.Path, "/items/")
	if len(asin) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Asin is empty"))
		return
	}

	client := paapi5.New(
		paapi5.WithMarketplace(localeMap[conf.AmazonLocale]),
	).CreateClient(
		conf.AmazonAssociateTag,
		conf.AmazonAccessKey,
		conf.AmazonSecretKey,
	)

	q := query.NewGetItems(client.Marketplace(), client.PartnerTag(), client.PartnerType())
	q.ASINs([]string{asin}).EnableBrowseNodeInfo().EnableImages().EnableItemInfo().EnableOffers().EnableParentASIN()

	mutex.Lock()
	time.Sleep(time.Millisecond * time.Duration(conf.AmazonRequestDelayMillisecond))
	res, err := client.Request(q)
	mutex.Unlock()

	if err == nil {
		var er ErrorResponse
		err = json.Unmarshal(res, &er)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
		}
		if len(er.Errors) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("Error: %s > %s", er.Errors[0].Code, er.Errors[0].Message)))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(res)
		return
	}

	if retry >= conf.AmazonRetryNumber {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
		return
	}

	r.WithContext(context.WithValue(r.Context(), retryKey, retry + 1))
	log.Printf("Retried asin=%s. %d times. msg=%s", asin, retry, err)

	getItems(w, r)
}
