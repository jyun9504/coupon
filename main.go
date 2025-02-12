package main

import (
	"fmt"
	"log"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// init 變數
var db *gorm.DB

func main() {
	// START POINT
	fmt.Println("Start!")

	var err error
	// init mysql 連線
	dsn := "root:password@tcp(127.0.0.1:3306)/coupon_db?parseTime=true"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	// Check DB connection
	fmt.Println("DB connection success!")

	// 遷移資料表 Schema
	db.AutoMigrate(&Customers{}, &Coupon{}, &CustomerCoupon{})

	fmt.Println("DB migrate success!")

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
		
	}

	// 初始化 gin 框架
	r := gin.Default()

	// API
	r.GET("/customers", GetCustomers)
	r.GET("/coupon/coupons", GetCoupons)
	r.POST("/coupon/claim", ClaimCoupon)

	r.Run(":8081")
}

// 取得所有顧客列表
func GetCustomers(c *gin.Context) {
	var customers []Customers
	db.Find(&customers)

	c.JSON(http.StatusOK, customers)
}

// 取得所有優惠券列表
func GetCoupons(c *gin.Context) {
	var coupons []Coupon
	db.Find(&coupons)

	c.JSON(http.StatusOK, coupons)
}

// 領取優惠券 API
func ClaimCoupon(c *gin.Context) {
	// Requset 參數
	type ClaimReq struct {
		CustomerID string `json:"customer_id"`
		CouponID string `json:"coupon_id"`
	}

	//檢查傳入參數是否格是正確，不正確就回傳 http 400 Error
	var req ClaimReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// 必填參數檢查
	if req.CustomerID == "" || req.CouponID == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 檢查優惠券剩餘數量，發完就回傳 http 404 Error
	var coupon Coupon
	if err := db.Where("id = ? AND remaining > 0", req.CouponID).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "優惠券已全數發放完畢，謝謝惠顧"})
		return
	}

	// 領取優惠券，剩餘數量 - 1
	db.Model(&Coupon{}).Where("id = ?", req.CouponID).Update("remaining", coupon.Remaining - 1)
}