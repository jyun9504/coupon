# Coupon (優惠券管理系統)

## 目錄
- [專案簡介](#專案簡介)
- [系統架構](#系統架構)
- [資料庫設計](#資料庫設計)
- [環境變數](#環境變數)
- [安裝與執行](#安裝與執行)
- [API 使用方式](#api-使用方式)
- [測試範例](#測試範例)

---

## 專案簡介
本專案是一個 **優惠券管理系統**，用於管理顧客與優惠券的領取與使用。

滿足以下需求：
* 用戶可領取特定優惠券（如滿減券、折扣券），每張優惠券數量有限。
* 用戶可以簡易設計 ex. ID, 名稱 .. 等基礎欄位就好
* 用戶在優惠券的有效期內使用，系統需驗證優惠券的可用性（例如過期或已使用則無法使用）。
* 提供查詢介面或是API，顯示使用者所有的優惠券狀態（未使用、已使用、已過期）。


技術棧：
- **Golang (Gin Framework)**
- **MySQL (GORM ORM)**
- **Redis (併發鎖管理)**

---

## 系統架構

```
+------------------+      +------------------+
| Gin API Server   |  ——> |       Redis      |
+------------------+      +------------------+
       |  
       v  
+------------------+
|  MySQL Database  |
+------------------+
```

---

## 資料庫設計

### **Customers (客戶表)**
| 欄位名稱 | 型態 | 附註 |
|----------|------|------|
| id       | CHAR(36) | 主鍵 (UUID) |
| name     | VARCHAR(20) | 唯一且不可為 NULL |

### **Coupons (優惠券表)**
| 欄位名稱 | 型態 | 附註 |
|----------|------|------|
| id            | CHAR(36) | 主鍵 (UUID) |
| name          | VARCHAR(100) | 優惠券名稱 |
| discount_type | ENUM ('price', 'percentage') | 折扣類型 |
| discount_value | DECIMAL(10,2) | 折扣值 |
| total_issued | INT | 發行總數 |
| remaining | INT | 剩餘數量 |
| expires_at | DATETIME | 到期日 |

### **CustomerCoupons (顧客優惠券表)**
| 欄位名稱 | 型態 | 附註 |
|----------|------|------|
| id         | CHAR(36) | 主鍵 (UUID) |
| customer_id | CHAR(36) | 外鍵 (Customers.id) |
| coupon_id   | CHAR(36) | 外鍵 (Coupons.id) |
| used        | BOOLEAN | 是否已使用 |
| claimed_at  | DATETIME | 領取時間 |
| used_at     | DATETIME | 使用時間 |

---

## 環境變數
在 `docker-compose.yml` 檔案內設定以下變數：
```
DB_USER=root
DB_PASSWORD=password
DB_NAME=coupon_db
DB_PORT=3306
RDB_PORT=6379
```

---

## 安裝與執行

1. **下載**
```sh
git clone https://github.com/jyun9504/coupon.git
```


2. **本地端執行**
```sh
$ cd ./coupon
$ go mod tidy
$ go run main.go 
```

or

2. **快速啟動 Docker 服務**
```sh
$ cd ./coupon
$ docker-compose up -d
```
- **預設執行在 http://localhost:8081**

---

## API 使用方式

### **取得所有顧客列表**
**GET** `/customer/customers`
```json
[
    {
        "ID": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
        "Name": "周星星"
    },
    {
        "ID": "0be41668-e040-4894-92bb-aa045f35b0b3",
        "Name": "王小龜"
    }
]
```

### **取得所有優惠券列表**
**GET** `/coupon/coupons`
```json
[
    {
        "ID": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5",
        "Name": "25% Discount",
        "DiscountType": "percentage",
        "DiscountValue": 25,
        "TotalIssued": 5,
        "Remaining": 4,
        "ExpiresAt": "2025-03-11T18:57:52.195Z"
    },
    {
        "ID": "b9e9f42f-d410-4fb4-b6af-d118c39f9893",
        "Name": "NT$500 Cashback",
        "DiscountType": "price",
        "DiscountValue": 500,
        "TotalIssued": 5,
        "Remaining": 5,
        "ExpiresAt": "2025-03-11T18:57:52.229Z"
    }
]
```

### **查詢顧客擁有的優惠券**
**GET** `/customer/{customer_id}/coupons`
```json
[
    {
        "ID": "1cc1c738-8279-4b6f-8495-36a7c6f1fc55",
        "CustomerID": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
        "CouponID": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5",
        "Used": false,
        "ClaimedAt": "2025-02-12T05:36:11.707Z",
        "UsedAt": null,
        "Customers": {
            "ID": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
            "Name": "周星星"
        },
        "Coupon": {
            "ID": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5",
            "Name": "25% Discount",
            "DiscountType": "percentage",
            "DiscountValue": 25,
            "TotalIssued": 5,
            "Remaining": 4,
            "ExpiresAt": "2025-03-11T18:57:52.195Z"
        }
    }
]
```

### **領取優惠券**
**POST** `/coupon/claim`
**Body:**
```json
{
    "customer_id": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
    "coupon_id": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5"
}
```

**回應:**
```json
{ "message": "恭喜您，成功領取優惠券" }
```

### **使用優惠券**
**POST** `/coupon/coupons/use`
**Body:**
```json
{
    "customer_id": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
    "coupon_id": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5"
}
```

**回應:**
```json
{ "message": "成功使用優惠券" }
```

---

## 測試範例

### **1. 使用 `cURL` 測試 API**

## 注意事項
- **每位顧客只能領取相同的優惠券一次**
- **輸入參數 id 請以 API 查詢的結果為主，範例中所示的 id 並非實際產生的 id**

#### **取得所有顧客列表**
```sh
curl -X GET "http://localhost:8081/customer/customers"
```
```json
[
    {
        "ID": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
        "Name": "周星星"
    },
    {
        "ID": "0be41668-e040-4894-92bb-aa045f35b0b3",
        "Name": "王小龜"
    }
]
```

#### **取得所有優惠券列表**
```sh
curl -X GET "http://localhost:8081/coupon/coupons"
```
```json
[
    {
        "ID": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5",
        "Name": "25% Discount",
        "DiscountType": "percentage",
        "DiscountValue": 25,
        "TotalIssued": 5,
        "Remaining": 4,
        "ExpiresAt": "2025-03-11T18:57:52.195Z"
    },
    {
        "ID": "b9e9f42f-d410-4fb4-b6af-d118c39f9893",
        "Name": "NT$500 Cashback",
        "DiscountType": "price",
        "DiscountValue": 500,
        "TotalIssued": 5,
        "Remaining": 5,
        "ExpiresAt": "2025-03-11T18:57:52.229Z"
    }
]
```

#### **查詢顧客擁有的優惠券**
```sh
curl -X GET "http://localhost:8081/customer/6f21c203-97c9-4318-a7b9-ccb6d41b9044/coupons"
```
```json
[
    {
        "ID": "1cc1c738-8279-4b6f-8495-36a7c6f1fc55",
        "CustomerID": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
        "CouponID": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5",
        "Used": false,
        "ClaimedAt": "2025-02-12T05:36:11.707Z",
        "UsedAt": null,
        "Customers": {
            "ID": "6f21c203-97c9-4318-a7b9-ccb6d41b9044",
            "Name": "周星星"
        },
        "Coupon": {
            "ID": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5",
            "Name": "25% Discount",
            "DiscountType": "percentage",
            "DiscountValue": 25,
            "TotalIssued": 5,
            "Remaining": 4,
            "ExpiresAt": "2025-03-11T18:57:52.195Z"
        }
    }
]
```

#### **領取優惠券**
```sh
curl -X POST "http://localhost:8081/coupon/claim" \
     -H "Content-Type: application/json" \
     -d '{"customer_id": "6f21c203-97c9-4318-a7b9-ccb6d41b9044", "coupon_id": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5"}'
```
**回應:**
```json
{ "message": "恭喜您，成功領取優惠券" }
```

#### **使用優惠券**
```sh
curl -X POST "http://localhost:8081/coupon/coupons/use" \
     -H "Content-Type: application/json" \
     -d '{"customer_id": "6f21c203-97c9-4318-a7b9-ccb6d41b9044", "coupon_id": "2ab9d49b-d058-4f6f-b5e0-a9f2be351ac5"}'
```
**回應:**
```json
{ "message": "成功使用優惠券" }
```

---