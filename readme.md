# Balance-service

Для запуска сервиса
```shell
make run
```
Для запуска тестов
```shell
make test
```
Для запуска линтера
```shell
make lint
```

## Описание методов

### GetBalance (GET)

```bash
# GetBalance
curl --location --request GET 'localhost:4444/wallet/getBalance' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ'
```
#### Example Response:
```
{
    "currency": "RUB",
    "amount": 201
}
```

### AddDeposit (POST)
```bash
# AddDeposit
curl --location --request POST 'localhost:4444/wallet/addDeposit' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "idempotence_key":1,
    "amount": 1000.50,
    "comment": "Пополнение баланса"
}
'
```
#### Response:
```
{
    "response": "OK"
}
```
### WithdrawMoney (POST)
```bash
# WithdrawMoney
curl --location --request POST 'localhost:4444/wallet/withdrawMoney' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "idempotence_key":2,
    "amount": 100.50,
    "comment": "Снятие средств"
}
'
```
#### Example Response:
```
{
    "response": "OK"
}
```
### TransferMoney (POST)
```bash
# TransferMoney
curl --location --request POST 'localhost:4444/wallet/transferMoney' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "idempotence_key":3,
    "target":333,
    "amount": 100.5,
    "comment":"Перевод"
}'
```
#### Example Response:
```
{
    "response": "OK"
}
```

### GetTransactions (GET)
params:

`?from` - Дата формата "2022-09-26T00:00:00Z"

`?to` - Дата формата "2022-09-26T00:00:00Z"

`?limit` - int

`?offset` - int

`?descending` - "true"/"false", default:"true"

`?sorting` - "amount"/"date", default:"date"


```bash
curl --location --request GET 'localhost:4444/wallet/getTransactions?from=2022-10-22T00:00:00Z&to=2022-10-23T00:00:00Z&limit=100&offset=0&descending=true&sorting=date' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ'
```
#### Example Response:
```
[
    {
        "id": 3,
        "wallet_id": 1,
        "amount": 1000.5,
        "target_wallet_id": null,
        "comment": "Пополнение баланса",
        "timestamp": "2022-09-25T18:45:38Z"
    },
    {
        "id": 2,
        "wallet_id": 1,
        "amount": -100.5,
        "target_wallet_id": null,
        "comment": "Снятие средств",
        "timestamp": "2022-09-25T18:43:43Z"
    },
    {
        "id": 1,
        "wallet_id": 1,
        "amount": 100.5,
        "target_wallet_id": null,
        "comment": "Пополнение баланса",
        "timestamp": "2022-09-25T18:42:16Z"
    }
]
```

### ReserveMoney (POST)
```bash
# ReserveMoney
curl --location --request POST 'localhost:4444/wallet/reserveMoney' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "account_id":555,
    "service_id":1,
    "order_id":111,
    "amount": 200
}'
```
#### Example Response:
```
{
    "response": "OK"
}
```

### ApplyReserve (POST)
```bash
# ApplyReserve
curl --location --request POST 'localhost:4444/wallet/applyReserve' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "account_id":555,
    "service_id":1,
    "order_id":111,
    "amount": 200
}'
```
#### Example Response:
```
{
    "response": "OK"
}
```

### CancelReserve (POST)
```bash
# CancelReserve
curl --location --request POST 'localhost:4444/wallet/cancelReserve' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "account_id":555,
    "service_id":1,
    "order_id":111,
    "amount": 200
}
'
```
#### Example Response:
```
{
    "response": "OK"
}
```

### GetReport (GET)
params:

`?month` - Дата формата (год-месяц) "2022-10"
```bash
# GetReport
curl --location --request GET 'localhost:4444/wallet/getReport?month=2022-10' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ'
```
#### Example Response:
```
ServiceTitle;Amount
Услуга;200

```