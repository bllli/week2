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
	Code  string
	Price uint
}

// daoGetProduct
// dao层wrap包装错误，抛到上一层
func daoGetProduct(code string) (*Product, error) {
	var p Product
	res := globalDB.First(&p, "code = ?", code)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, fmt.Sprintf("get product failed: %s", code))
	}
	return &p, nil
}

// apiGetProduct
// 获取指定code产品
func apiGetProduct(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		writeFailMsgCode(c, 400, "wrong product code")
		return
	}

	product, err := daoGetProduct(code)
	if err != nil {
		// 可以识别根因 根据特定业务逻辑处理
		// 如找不到时返回404 其他错误返回500
		if errors.Cause(err) == gorm.ErrRecordNotFound {
			writeFailMsgCode(c, 404, "not found")
			return
		} else {
			// 未识别的错误，可以把堆栈打log
			fmt.Printf("apiGetProduct error: \n%+v\n", err)
			writeFailMsgCode(c, 500, "error")
			return
		}
	}
	writeSuccessData(c, product)
	return
}

// apiRecommendOneProduct
// 推荐一件产品
func apiRecommendOneProduct(c *gin.Context) {
	recommendsProductCode := "NICE" // mock
	product, err := daoGetProduct(recommendsProductCode)
	if err != nil {
		// 不care错误 降级处理
		writeSuccessData(c, &Product{
			Code:  "COOL",
			Price: 100,
		})
		return
	}
	writeSuccessData(c, product)
	return
}

type apiProductCreateOrUpdateReq struct {
	Code  string `json:"code"`
	Price uint   `json:"price"`
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

func resetDB() {
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
			"msg":  msg,
		},
	})
}

func writeSuccessData(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"status": gin.H{
			"code": 200,
			"msg":  "ok",
		},
		"data": data,
	})
}

func main() {
	resetDB()
	r := gin.Default()
	r.GET("/product/get", apiGetProduct)
	r.POST("/product/update", apiCreateOrUpdateProduct)
	r.GET("/product/recommend", apiRecommendOneProduct)
	r.Run() // listen and serve on 0.0.0.0:8080
}
