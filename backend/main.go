package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Video struct {
	gorm.Model
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
}

var db *gorm.DB

func init() {
	var err error
	dsn := "host=localhost user=youruser dbname=yourdb sslmode=disable password=yourpassword"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&Video{})
}

func main() {
	r := gin.Default()
	r.POST("/upload", uploadHandler)
	r.Run(":8080")
}

func uploadHandler(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	// Upload to AWS S3
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String("your-region"),
		Credentials: credentials.NewStaticCredentials("your-access-key", "your-secret-key", ""),
	})
	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("your-bucket-name"),
		Key:    aws.String(file.Filename),
		Body:   src,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save metadata to PostgreSQL
	video := Video{FileName: file.Filename, FileURL: result.Location}
	db.Create(&video)

	c.JSON(http.StatusOK, gin.H{"message": "Upload successful", "file_url": result.Location})
}
