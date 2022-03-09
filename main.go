package main

import (
	"math/rand"
	"mime/multipart"
	"os"
	"time"

	_ "fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Upload struct {
  gorm.Model
  Url string
  Name  string
  Uploader string
  FileID int
  File File
}

type File struct {
  gorm.Model
  ReferenceCount int
  Hash string
  Path string
}

func findOrCreateFile(db *gorm.DB, fileData *multipart.FileHeader) (*File, error) {
  fileHash, err := fileHash(fileData)
  if err != nil {
    return nil, err
  }
  
  var file File
  res := db.Find(&file, "hash = ?", fileHash)
  if res.Error != nil {
    return nil, res.Error
  }

  if res.RowsAffected == 0 {
    file = File {
      ReferenceCount: 1,
      Hash: fileHash,
      Path: "./data/" + fileHash,
    }
    db.Create(&file)
    return &file, nil
  } else {
    file.ReferenceCount += 1
    db.Model(&file).Update("reference_count", file.ReferenceCount)
    return &file, nil
  }
}

func main() {
  rand.Seed(time.Now().UnixNano())

  if err := os.MkdirAll("./data", os.ModePerm); err != nil {
    panic("Failed to create data directory: " + err.Error())
  }

  db, err := gorm.Open(sqlite.Open("fileman.db"), &gorm.Config{})
  if err != nil {
    panic("Failed to connect database.")
  }

  db.AutoMigrate(&Upload{}, &File{})

  r := gin.Default()
  r.SetTrustedProxies(nil)

  fileCRUD := r.Group("/file")
  {
    fileCRUD.GET("/:url", func(c *gin.Context) {
      var upload Upload
      res := db.Preload("File").First(&upload, "url = ?", c.Param("url"))
      if res.Error != nil {
        c.JSON(404, gin.H{"message": "File not found"})
      } else {
        c.FileAttachment(upload.File.Path, upload.Name)
      }
    })

    fileCRUD.GET("/:url/raw", func(c *gin.Context) {
      var upload Upload
      res := db.Preload("File").First(&upload, "url = ?", c.Param("url"))
      if res.Error != nil {
        c.JSON(404, gin.H{"message": "File not found"})
      } else {
        c.File(upload.File.Path)
      }
    })

    fileCRUD.POST("", func(c *gin.Context) {
      fileData, err := c.FormFile("file")
      if err != nil {
        c.JSON(500, gin.H{"message": "Error reading file."})
        return
      }

      file, err := findOrCreateFile(db, fileData)
      if err != nil {
        c.JSON(500, gin.H{"message": "Error uploading file."})
        return
      } else if file.ReferenceCount == 1 {
        c.SaveUploadedFile(fileData, file.Path)
      }

      upload := Upload {
        Url: randID(8),
        Name: fileData.Filename,
        Uploader: c.ClientIP(),
        FileID: int(file.ID),
      }
      db.Create(&upload)

      c.JSON(200, gin.H{
        "name": upload.Name,
        "path": "/file/" + upload.Url,
      })
    })

    fileCRUD.DELETE("/:url", func(c *gin.Context) {
      var upload Upload
      res := db.Preload("File").First(&upload, "url = ?", c.Param("url"))
      if res.Error != nil {
        c.JSON(404, gin.H{"message": "File not found"})
      } else {
        upload.File.ReferenceCount -= 1
        if upload.File.ReferenceCount == 0 {
          os.Remove(upload.File.Path)
          db.Delete(&upload.File)
        } else {
          db.Model(&upload.File).Update("reference_count", upload.File.ReferenceCount)
        }
        db.Delete(&upload)
        c.JSON(200, gin.H{"message": "File deleted successfully."})
      }
    })
  }

  r.Run()
}
