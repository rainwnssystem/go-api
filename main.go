package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	// "github.com/joho/godotenv"
)

type User struct {
	gorm.Model
	ID    uint   `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Name  string `json:"name"`
	Email string `json:"email" gorm:"index:idx_user,unique"`
}

func main() {

	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&User{})

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	// POST /users
	r.POST("/users", func(c *gin.Context) {
		var u User
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var exists bool
		err = db.Model(&User{}).
			Select("count(*) > 0").
			Where("email = ?", u.Email).
			Find(&exists).
			Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		if err := db.Create(&u).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
		c.JSON(http.StatusCreated, u)
	})

	// GET /users/:id
	r.GET("/users/:id", func(c *gin.Context) {
		var u User
		if err := db.First(&u, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, u)
	})

	// GET /users
	r.GET("/users", func(c *gin.Context) {
		var u []User
		if _u := db.Find(&u); _u.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": _u.Error})
			return
		} else if _u.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"error": "there are no users"})
			return
		}

		c.JSON(http.StatusOK, u)
	})

	r.Run()
}
