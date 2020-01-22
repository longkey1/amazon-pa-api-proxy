package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/dominicphillips/amazing"
	"github.com/labstack/echo"
)

const (
	version = "0.1.1"
)

// Config ...
type Config struct {
	Port      string          `toml:"port"`
	AmazonAPI AmazonAPIConfig `toml:"amazon"`
}

// AmazonAPIConfig  ...
type AmazonAPIConfig struct {
	AssociateTag  string `toml:"associate_tag"`
	AccessKey     string `toml:"access_key"`
	SecretKey     string `toml:"secret_key"`
	ServiceDomain string `toml:"service_domain"`
	ResponseGroup string `toml:"response_group"`
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
	flag.StringVar(&configPath, "c", "config.tml", "configuration file path")
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

	client, err := amazing.NewAmazing(conf.AmazonAPI.ServiceDomain, conf.AmazonAPI.AssociateTag, conf.AmazonAPI.AccessKey, conf.AmazonAPI.SecretKey)
	params := url.Values{
		"ResponseGroup": []string{conf.AmazonAPI.ResponseGroup},
	}
	res, err := client.ItemLookupAsin(asin, params)
	if err != nil {
		if retry < 2 {
			ctx.Set("retry", retry + 1)
			return getItem(ctx)
		}
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
	}
	return ctx.JSON(http.StatusOK, res)
}
