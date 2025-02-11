# coupon

## 設計並實作一個優惠券系統，滿足以下需求：

* 用戶可領取特定優惠券（如滿減券、折扣券），每張優惠券數量有限。
* 用戶可以簡易設計 ex. ID, 名稱 .. 等基礎欄位就好
* 用戶在優惠券的有效期內使用，系統需驗證優惠券的可用性（例如過期或已使用則無法使用）。
* 提供查詢介面或是API，顯示使用者所有的優惠券狀態（未使用、已使用、已過期）。

## 資料庫設計

### customers（用戶表）

| 欄位  | 類型        | 描述    |
|------|-------------|---------|
| id   | UUID (PK)   | 用戶 ID |
| name | VARCHAR(20) | 用戶名稱|

### coupons（優惠券表）

| 欄位          | 類型                         | 描述            |
|---------------|-----------------------------|-----------------|
| id            | UUID (PK)                   | 優惠券唯一識別碼 |
| name          | VARCHAR(100)                | 優惠券名稱      |
| discount_type | ENUM('price', 'percentage') | 折扣類型        |
| discount_value| DECIMAL(10,2)               | 折扣數值        |
| total_issued  | INT                         | 總發行數量      |
| remaining     | INT                         | 剩餘數量        |
| expires_at    | DATETIME                    | 到期時間        |

### customers_coupons（用戶優惠券表）

| 欄位         | 類型       | 描述       |
|--------------|-----------|------------|
| id           | UUID (PK) | 記錄 ID    |
| customers_id | UUID (FK) | 用戶 ID    |
| coupon_id    | UUID (FK) | 優惠券 ID  |
| used         | BOOLEAN   | 是否已使用 |
| claimed_at   | DATETIME  | 領取時間   |
| used_at      | DATETIME  | 使用時間   |
