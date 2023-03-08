package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB_URL string
var DB_USER string
var DB_PASS string
var database *gorm.DB

func main() {
	DB_URL = Get("DB_URL", "")
	DB_USER = Get("DB_USER", "")
	DB_PASS = Get("DB_PASS", "")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:26257/defaultdb?sslmode=verify-full", DB_USER, DB_PASS, DB_URL)
	var err error
	database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/list", list)
	e.GET("/test", test)
	e.POST("/add", add)
	e.Logger.Fatal(e.Start(":8080"))
}

func list(c echo.Context) error {
	var items []Item
	db := database.Begin()
	err := db.Model(&Item{}).Select("id", "name").Find(&items).Error

	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error during select: %s\n", err))
	}

	response := &ListDto{
		Items: make([]ItemDto, 0),
	}

	for _, item := range items {
		response.Items = append(response.Items, ItemDto{Id: item.ID, Name: item.Name})
	}

	return c.JSON(http.StatusOK, response)
}

func add(c echo.Context) error {
	var itemdto ItemDto
	err := c.Bind(&itemdto)
	if err != nil {
		fmt.Printf("Error during incoming object parse")
		return c.String(http.StatusBadRequest, "Error during incoming object parse")
	}

	var item Item
	item.Name = itemdto.Name
	db := database.Begin()
	errInsert := db.Omit("id").Create(&item).Error
	if errInsert != nil {
		fmt.Printf("Cannot insert record, %s", errInsert)
		return c.String(http.StatusBadRequest, "Cannot insert record")
	}

	db.Commit()
	return c.String(http.StatusOK, "")
}

type Item struct {
	ID   string `gorm:"type:uuid"`
	Name string
}

func (Item) TableName() string {
	return "item_t"
}

type ItemDto struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ListDto struct {
	Items []ItemDto `json:"items,omitempty"`
}

func Get(key, _default string) string {
	if value, valid := os.LookupEnv(key); valid {
		if len(value) > 0 {
			return value
		}
	}

	if _default != "" {
		fmt.Printf("The '%s' variable is not set. Defaulting to '%s'\n", key, _default)
		return _default
	}

	panic(fmt.Sprintf("'%s' environment variable is missinig.\n", key))
}

func test(c echo.Context) error {
	var now time.Time
	db := database.Begin()
	db.Raw("SELECT NOW()").Scan(&now)

	return c.String(http.StatusOK, fmt.Sprintf("%s\n", now))
}
