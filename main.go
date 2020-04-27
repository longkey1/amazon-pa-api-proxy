package main

import (
	"context"
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
	version = "0.7.0"
	retryKey = "retry"
)

// Config ...
type Config struct {
	Port                          string `default:"1323"`
	AmazonAssociateTag            string `required:"true" split_words:"true"`
	AmazonAccessKey               string `required:"true" split_words:"true"`
	AmazonSecretKey               string `required:"true" split_words:"true"`
	AmazonLocale                  string `required:"true" split_words:"true"`
	AmazonRetryNumber             int    `default:"3" split_words:"true"`
	AmazonRequestDelayMillisecond int    `default:"1000" split_words:"true"`
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

	log.Print("amazon-product-json start...")
	http.HandleFunc("/items/", getItems)
	log.Fatal(http.ListenAndServe(":" + conf.Port, nil))
}

func checkVersion() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("amazon-product-json version %s\n", version)
		os.Exit(0)
	}
}

func loadConfig() {
	err := envconfig.Process("apj", conf)
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
