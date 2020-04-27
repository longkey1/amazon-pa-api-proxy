# amazon-product-json

Web API Server that returns results of Amazon Product Product Advertising API's GetItems Operation with JSON

## Install

Download binary file in [release page](https://github.com/longkey1/amazon-product-json/releases)

## Usage

```
./amazon-product-json
```

```
http get http://localhost:1313/items/{asin}
```

## Configuration

### Environment variables

- `APJ_PORT` (Default: 1323)
- `APJ_AMAZON_ASSOCIATE_TAG`
- `APJ_AMAZON_ACCESS_KEY`
- `APJ_AMAZON_SECRET_KEY`
- `APJ_AMAZON_LOCALE`
- `APJ_AMAZON_RETRY_NUMBER` (Default: 3)
- `APJ_AMAZON_REQUEST_DELAY_MILLISECOND` (Default: 1000)

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
