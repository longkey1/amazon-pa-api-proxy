package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	paapi5 "github.com/spiegel-im-spiegel/pa-api"
	"github.com/spiegel-im-spiegel/pa-api/query"
    "github.com/kelseyhightower/envconfig"
)

const (
	version = "0.5.1"
)

// Config ...
type Config struct {
	Port                   string `default:"1323"`
	AmazonAssociateTag     string `required:"true" split_words:"true"`
	AmazonAccessKey        string `required:"true" split_words:"true"`
	AmazonSecretKey        string `required:"true" split_words:"true"`
	AmazonLocale           string `required:"true" split_words:"true"`
	AmazonRetryNumber      int    `default:"10" split_words:"true"`
	AmazonRetryDelaySecond int    `default:"3" split_words:"true"`
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

var conf Config

func main() {
	checkVersion()
	err := envconfig.Process("apj", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}
	e := echo.New()
	e.GET("/items/:asin", getItem)
	e.Logger.Fatal(e.Start(":" + conf.Port))
}

func checkVersion() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("amazon-product-json version %s\n", version)
		os.Exit(0)
	}
}

func getItem(ctx echo.Context) error {
	retry := 0
	if ctx.Get("retry") != nil {
		retry = ctx.Get("retry").(int)
	}
	asin := ctx.Param("asin")
	if len(asin) == 0 {
		return ctx.String(http.StatusBadRequest, "Asin is empty")
	}

	client := paapi5.New(
		paapi5.WithMarketplace(localeMap[conf.AmazonLocale]),
	).CreateClient(
		conf.AmazonAssociateTag,
		conf.AmazonAccessKey,
		conf.AmazonSecretKey,
	)

	time.Sleep(time.Second * time.Duration(1))
	q := query.NewGetItems(client.Marketplace(), client.PartnerTag(), client.PartnerType())
	q.ASINs([]string{asin}).EnableBrowseNodeInfo().EnableImages().EnableItemInfo().EnableOffers().EnableParentASIN()

	res, err := client.Request(q)
	if err != nil {
		if retry < conf.AmazonRetryNumber {
			ctx.Set("retry", retry+1)
			time.Sleep(time.Second * time.Duration(conf.AmazonRetryDelaySecond))
			ctx.Logger().Printf("Retried asin=%s. %d times. msg=%s", asin, retry, err)

			return getItem(ctx)
		}

		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
	}

	return ctx.JSONBlob(http.StatusOK, res)
}
