# amazon-pa-api-proxy

Amazon Product Advertising API(PA-API) Proxy Server.

## Install

Download binary file in [release page](https://github.com/longkey1/amazon-pa-api-proxy/releases)

## Usage

```
$ ./amazon-pa-api-proxy
```

```
http get http://localhost:1313/items/{asin}
```

## Configuration

### Environment variables

- `APAP_PORT` (Default: 1323)
- `APAP_AMAZON_ASSOCIATE_TAG`
- `APAP_AMAZON_ACCESS_KEY`
- `APAP_AMAZON_SECRET_KEY`
- `APAP_AMAZON_LOCALE`
- `APAP_AMAZON_RETRY_NUMBER` (Default: 3)
- `APAP_AMAZON_REQUEST_DELAY_MILLISECOND` (Default: 1000)

### Locale Map

- Australia
- Brazil
- Canada
- France
- Germany
- India
- Italy
- Japan
- Mexico
- Spain
- Turkey
- UnitedArabEmirates
- UnitedKingdom
- UnitedStates
