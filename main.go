package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	amazonproduct "github.com/DDRBoxman/go-amazon-product-api"
	"github.com/labstack/echo"
)

const (
	version = "0.0.3"
)

// Config ...
type Config struct {
	Port      string          `toml:"port"`
	AmazonAPI AmazonAPIConfig `toml:"amazon"`
}

// AmazonAPIConfig  ...
type AmazonAPIConfig struct {
	AccessKey           string `toml:"access_key"`
	SecretKey           string `toml:"secret_key"`
	Host                string `toml:"host"`
	AssociateTag        string `toml:"associate_tag"`
	ResponseGroup       string `toml:"response_group"`
	MaxRetryNumber      int    `toml:"max_retry_number"`
	RetryDurationSecond int    `toml:"retry_duration_second"`
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

	var itemRes *amazonproduct.ItemLookupResponse
	api := getAPIClient()
	for i := 0; i < conf.AmazonAPI.MaxRetryNumber; i++ {
		// wait
		rand.Seed(time.Now().UnixNano())
		s := rand.Intn(conf.AmazonAPI.RetryDurationSecond)
		if s > 0 {
			time.Sleep(time.Duration(s) * time.Second)
		}

		res, err := api.ItemLookupWithResponseGroup(asin, conf.AmazonAPI.ResponseGroup)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		}
		itemRes = new(amazonproduct.ItemLookupResponse)
		err = xml.Unmarshal([]byte(res), itemRes)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		}
		if len(itemRes.Items.Item.ASIN) == 0 {
			continue
		}

		break
	}

	if len(itemRes.Items.Item.ASIN) == 0 {
		return ctx.String(http.StatusInternalServerError, "Error: Invalid response")
	}

	return ctx.JSON(http.StatusOK, itemRes)
}
