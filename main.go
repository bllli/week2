package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

var globalDB *gorm.DB


type Product struct {
	gorm.Model
	Code string
	Price uint
}

// daoGetProduct
func daoGetProduct(code string) (*Product, error) {
	var p Product
	res := globalDB.First(&p, "code = ?", code)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, fmt.Sprintf("get product failed: %s", code))
	}
	return &p, nil
}

func apiGetProduct(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		writeFailMsgCode(c, 400, "wrong product code")
		return
	}

	product, err := daoGetProduct(code)
	if err != nil {
		if errors.Cause(err) == gorm.ErrRecordNotFound {
			writeFailMsgCode(c, 404, "not found")
			return
		} else {
			//
			fmt.Printf("apiGetProduct error: \n%+v\n", err)
			writeFailMsgCode(c, 500, "error")
			return
		}
	}
	writeSuccessData(c, product)
	return
}

type apiProductCreateOrUpdateReq struct {
	Code string `json:"code"`
	Price uint `json:"price"`
}
func apiCreateOrUpdateProduct(c *gin.Context) {
	var req apiProductCreateOrUpdateReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		writeFailMsgCode(c, 400, "error params")
		return
	}
	product, err := daoGetProduct(req.Code)
	if err != nil {
		// ignore err
		// create
		product = &Product{
			Code:  req.Code,
			Price: req.Price,
		}
		globalDB.Create(product)
	} else {
		// update
		globalDB.Model(product).Update("Price", req.Price)
	}
	writeSuccessData(c, product)
	return
}


func resetAndInitDB() {
	dbFile := "bllli.db"
	_ = os.Remove(dbFile)
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect db")
	}

	_ = db.AutoMigrate(&Product{})
	db.Create(&Product{
		Code:  "A1",
		Price: 100,
	})
	globalDB = db
}

func writeFailMsgCode(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"status": gin.H{
			"code": code,
			"msg": msg,
		},
	})
}

func writeSuccessData(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"status": gin.H{
			"code": 200,
			"msg": "ok",
		},
		"data": data,
	})
}

func main() {
	resetAndInitDB()
	r := gin.Default()
	r.GET("/product/get", apiGetProduct)
	r.POST("/product/update", apiCreateOrUpdateProduct)
	r.Run() // listen and serve on 0.0.0.0:8080
}
