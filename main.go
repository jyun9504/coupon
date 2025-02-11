package main

import (
	"fmt"
	"log"
	"time"

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
	ExpiresAt     time.Time `gorm:"not null"`
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
	dsn := "root:password@tcp(127.0.0.1:3306)/coupon_db"
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
	db.Model(&Customers{}).Count(&count)
	if count == 0 {
		// insert 測試客戶資料
		users := []User{
			{Name: "王小龜"},
			{Name: "周星星"},
		}
		db.Create(&users)

		// insert 優惠券資料
		coupons := []Coupon{
			{Name: "25% Discount", Type: "percentage", DiscountValue: 25, TotalIssued: 5, Remaining: 5, ExpiresAt: time.Now().AddDate(0, 1, 0)},
			{Name: "NT$500 Cashback", Type: "price", DiscountValue: 500, TotalIssued: 5, Remaining: 5, ExpiresAt: time.Now().AddDate(0, 1, 0)},
		}
		db.Create(&coupons)
		
	}
}