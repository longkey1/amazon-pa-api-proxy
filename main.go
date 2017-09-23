package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	amazonproduct "github.com/DDRBoxman/go-amazon-product-api"
	xj "github.com/basgys/goxml2json"
	"github.com/labstack/echo"
)

const (
	version = "0.0.1"
)

// Config ...
type Config struct {
	Port      string          `toml:"port"`
	AmazonAPI AmazonAPIConfig `toml:"amazon"`
}

// AmazonAPIConfig  ...
type AmazonAPIConfig struct {
	AccessKey     string `toml:"access_key"`
	SecretKey     string `toml:"secret_key"`
	Host          string `toml:"host"`
	AssociateTag  string `toml:"associate_tag"`
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
	if len(os.Args) > 1 && os.Args[1] == "version" {
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

func getAPIClient() amazonproduct.AmazonProductAPI {
	var api amazonproduct.AmazonProductAPI

	api.AccessKey = conf.AmazonAPI.AccessKey
	api.SecretKey = conf.AmazonAPI.SecretKey
	api.Host = conf.AmazonAPI.Host
	api.AssociateTag = conf.AmazonAPI.AssociateTag
	api.Client = &http.Client{} // optional

	return api
}

func getItem(ctx echo.Context) error {
	asin := ctx.Param("asin")
	if string(asin) == "" {
		return ctx.String(http.StatusBadRequest, "Asin is invalid")
	}

	api := getAPIClient()
	res, err := api.ItemLookupWithResponseGroup(asin, conf.AmazonAPI.ResponseGroup)
	if err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
	}

	buf, err := xj.Convert(strings.NewReader(res))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
	}

	return ctx.JSONBlob(http.StatusOK, buf.Bytes())
}
