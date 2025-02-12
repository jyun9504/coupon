package main

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 客戶 Struct
type Customers struct {
	ID   string `gorm:"type:char(36);primaryKey"` // MySQL UUID 要使用 CHAR(36)
	Name string `gorm:"type:varchar(20);unique;not null"`
}

// 優惠券 Struct
type Coupon struct {
	ID            string    `gorm:"type:char(36);primaryKey"`
	Name          string    `gorm:"type:varchar(100);unique;not null"`
	DiscountType  string    `gorm:"type:enum('price', 'percentage');not null"`
	DiscountValue float64   `gorm:"type:decimal(10,2);not null"`
	TotalIssued   int       `gorm:"not null"`
	Remaining     int       `gorm:"not null"`
	ExpiresAt     time.Time `gorm:"type:datetime;not null"`
}

// 客戶擁有優惠券 Struct
type CustomerCoupon struct {
	ID         string     `gorm:"type:char(36);primaryKey"`
	CustomerID string     `gorm:"type:char(36);not null"`
	CouponID   string     `gorm:"type:char(36);not null"`
	Used       bool       `gorm:"not null;default:false"`
	ClaimedAt  time.Time  `gorm:"not null"`
	UsedAt     *time.Time `gorm:"default:null"`

	// 關聯
	Customers Customers `gorm:"foreignKey:CustomerID;references:ID"`
	Coupon   Coupon   `gorm:"foreignKey:CouponID;references:ID"`
}

// 生成 UUID
func (customer *Customers) BeforeCreate(tx *gorm.DB) (err error) {
	customer.ID = uuid.New().String()
	return
}

func (coupon *Coupon) BeforeCreate(tx *gorm.DB) (err error) {
	coupon.ID = uuid.New().String()
	return
}

func (customerCoupon *CustomerCoupon) BeforeCreate(tx *gorm.DB) (err error) {
	customerCoupon.ID = uuid.New().String()
	return
}

func (CustomerCoupon) TableName() string {
	return "customer_coupons"
}

// 取得環境變數
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// init 變數
var db *gorm.DB
var rdb *redis.Client
var ctx = context.Background()

func main() {
	fmt.Println("1-開始運行")

	// 取得環境變數
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "coupon_db")
	dbPort := getEnv("DB_PORT", "3306")
	rdbHost := getEnv("RDB_HOST", "localhost")
	rdbPort := getEnv("RDB_PORT", "6379")

	var err error
	// init mysql 連線
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	// Check DB connection
	fmt.Println("2-DB 連線成功")

	// 遷移資料表 Schema
	db.AutoMigrate(&Customers{}, &Coupon{}, &CustomerCoupon{})
	fmt.Println("3-遷移資料表 Schema 成功")

	// 測試用初始資料
	var count int64
	db.Model(&Customers{}).Count(&count)
	if count == 0 {
		// insert 測試客戶資料
		users := []Customers{
			{Name: "王小龜"},
			{Name: "周星星"},
		}
		db.Create(&users)

		// insert 優惠券資料
		coupons := []Coupon{
			{Name: "25% Discount", DiscountType: "percentage", DiscountValue: 25.00, TotalIssued: 5, Remaining: 5, ExpiresAt: time.Now().AddDate(0, 1, 0)},
			{Name: "NT$500 Cashback", DiscountType: "price", DiscountValue: 500.00, TotalIssued: 5, Remaining: 5, ExpiresAt: time.Now().AddDate(0, 1, 0)},
		}
		db.Create(&coupons)
		fmt.Println("*3-首次執行，注入測試資料")
	}

	// 初始化 Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", rdbHost, rdbPort),
	})

	fmt.Println("4-Redis 初始化成功")


	// 初始化 gin 框架
	r := gin.Default()

	// API
	r.GET("/customer/customers", GetCustomers)
	r.GET("/coupon/coupons", GetCoupons)
	r.GET("/customer/:customer_id/coupons", GetCustomerCoupons)
	r.POST("/coupon/claim", ClaimCoupon)
	r.POST("/coupon/coupons/use", UseCoupon)
	
	r.Run(":8081")
}

// 取得所有顧客列表 API
func GetCustomers(c *gin.Context) {
	var customers []Customers
	db.Find(&customers)
	
	c.JSON(http.StatusOK, customers)
}

// 取得所有優惠券列表 API
func GetCoupons(c *gin.Context) {
	var coupons []Coupon
	db.Find(&coupons)

	c.JSON(http.StatusOK, coupons)
}

// 查詢用戶優惠券 API
func GetCustomerCoupons(c *gin.Context) {
	// 取得 param 參數
	customerID := c.Param("customer_id")

	logAction("查詢用戶優惠券", &customerID)

	// 查詢 DB customer_id == customerID，修正關聯空資料的問題
	var coupons []CustomerCoupon
	db.Preload("Customers").Preload("Coupon").Where("customer_id = ?", customerID).Find(&coupons)

	// Response
	c.JSON(http.StatusOK, coupons)
}

// 領取優惠券 API
func ClaimCoupon(c *gin.Context) {
	// Requset 參數
	type ClaimReq struct {
		CustomerID string `json:"customer_id"`
		CouponID string `json:"coupon_id"`
	}

	// 檢查傳入參數是否格是正確，不正確就回傳 http 400 Error
	var req ClaimReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logAction("領取優惠券失敗 - 輸入格式不正確", &req.CustomerID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入格式不正確"})
		return
	}
	// 必填參數檢查
	if req.CustomerID == "" || req.CouponID == ""{
		logAction("領取優惠券失敗 - 輸入參數不能為空", &req.CustomerID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入參數不能為空"})
		return
	}

	// 建立 Redis 鎖，防止併發超發問題
	lockKey := fmt.Sprintf("lock_%d", req.CouponID)
	if !doLock(lockKey) {
		logAction("領取優惠券失敗 - 系統忙碌中", &req.CustomerID)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "目前優惠券系統忙碌中，請稍後重新嘗試"})
		return
	}

	// 檢查優惠券剩餘數量，發完就回傳 http 404 Error
	var coupon Coupon
	if err := db.Where("id = ? AND remaining > 0", req.CouponID).First(&coupon).Error; err != nil {
		logAction("領取優惠券失敗 - 優惠券已無庫存", &req.CustomerID)
		c.JSON(http.StatusNotFound, gin.H{"error": "優惠券已全數發放完畢，謝謝惠顧"})
		return
	}

	// 領取優惠券，剩餘數量 - 1
	db.Model(&Coupon{}).Where("id = ?", req.CouponID).Update("remaining", coupon.Remaining - 1)

	// 寫入客戶擁有的優惠券資料表
	customerCoupon := CustomerCoupon{CustomerID: req.CustomerID, CouponID: req.CouponID, ClaimedAt: time.Now()}
	db.Create(&customerCoupon)

	// 解鎖
	unlock(lockKey)

	logAction("領取優惠券成功", &req.CustomerID)

	// Response 領取成功，回傳 http 200 Success
	c.JSON(http.StatusOK, gin.H{"message": "恭喜您，成功領取優惠券"})
}

// 使用優惠券 API
func UseCoupon(c *gin.Context) {
	// Requset 參數
	type UseReq struct {
		CustomerID string `json:"customer_id"`
		CouponID string `json:"coupon_id"`
	}
	
	// 檢查傳入參數是否格是正確，不正確就回傳 http 400 Error
	var req UseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logAction("使用優惠券失敗 - 輸入格式不正確", &req.CustomerID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入格式不正確"})
		return
	}
	// 必填參數檢查
	if req.CustomerID == "" || req.CouponID == ""{
		logAction("使用優惠券失敗 - 輸入參數不能為空", &req.CustomerID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入參數不能為空"})
		return
	}

	// 檢查客戶是否擁有有效的優惠券
	var customerCoupon CustomerCoupon
	if err := db.Where("customer_id = ? AND coupon_id = ? AND used = false", req.CustomerID, req.CouponID).
		First(&customerCoupon).Error; err != nil {
		logAction("使用優惠券失敗 - 輸入的優惠券不可使用", &req.CustomerID)
		c.JSON(http.StatusNotFound, gin.H{"error": "您輸入的優惠券不可使用，請檢查是否已過期或已使用"})
		return
	}

	// 使用前先更新使用狀態，紀錄使用時間
	now := time.Now()
	db.Model(&CustomerCoupon{}).Where("id = ?", customerCoupon.ID).Updates(map[string]interface{}{
		"used":   true,
		"used_at": now,
	})

	logAction("成功使用優惠券", &req.CustomerID)

	// Response 使用成功，回傳 http 200 Success
	c.JSON(http.StatusOK, gin.H{"message": "成功使用優惠券"})
}

// 建立一把 Redis 鎖，五秒後會自動移除以防死鎖
func doLock(key string) bool {
	success, _ := rdb.SetNX(ctx, key, "locked", 5*time.Second).Result()
	return success
}

// 領取優惠券完成需要解鎖
func unlock(key string) {
	rdb.Del(ctx, key)
}

// 新增 LOG 紀錄操作功能
func logAction(action string, customerID *string) {
	// 取得當前時間（格式：YYYY-MM-DD HH:MM:SS）
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 構造 LOG 訊息
	if customerID != nil {
		fmt.Printf("[%s] 操作: %s | 客戶 ID: %s\n", timestamp, action, *customerID)
	} else {
		fmt.Printf("[%s] 操作: %s\n", timestamp, action)
	}
}