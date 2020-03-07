package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
	paapi5 "github.com/spiegel-im-spiegel/pa-api"
	"github.com/spiegel-im-spiegel/pa-api/query"
)

const (
	version = "0.4.0"
)

// Config ...
type Config struct {
	Port      string          `toml:"port"`
	AmazonAPI AmazonAPIConfig `toml:"amazon"`
}

// AmazonAPIConfig  ...
type AmazonAPIConfig struct {
	AssociateTag     string `toml:"associate_tag"`
	AccessKey        string `toml:"access_key"`
	SecretKey        string `toml:"secret_key"`
	Locale           string `toml:"locale"`
	RetryNumber      int    `toml:"retry_number"`
	RetryDelaySecond int    `toml:"retry_delay_second"`
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
	loadConfig()
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

func loadConfig() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.toml", "configuration file path")
	flag.Parse()

	if _, err := toml.DecodeFile(configPath, &conf); err != nil {
		panic(err)
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
		paapi5.WithMarketplace(localeMap[conf.AmazonAPI.Locale]),
	).CreateClient(
		conf.AmazonAPI.AssociateTag,
		conf.AmazonAPI.AccessKey,
		conf.AmazonAPI.SecretKey,
	)

	q := query.NewGetItems(client.Marketplace(), client.PartnerTag(), client.PartnerType())
	q.ASINs([]string{asin}).EnableBrowseNodeInfo().EnableImages().EnableItemInfo().EnableOffers().EnableParentASIN()

	res, err := client.Request(q)
	if err != nil {
		if retry < conf.AmazonAPI.RetryNumber {
			ctx.Set("retry", retry+1)
			time.Sleep(time.Second * time.Duration(conf.AmazonAPI.RetryDelaySecond))
			ctx.Logger().Printf("Retried asin=%s. %d times. msg=%s", asin, retry, err)

			return getItem(ctx)
		}

		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
	}

	return ctx.JSONBlob(http.StatusOK, res)
}
