package main

import (
	"onlineBing/api"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	fmt.Println("伺服器開始運行")

	// 初始化數據庫連接
	api.InitDB()

	// 創建 Echo 實例
	e := echo.New()

	// 啟用 CORS 中間件
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // 允許所有域名，您也可以指定特定域名
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// 加載 API 路由
	api.LoadRoutes(e)

	// 啟動服務
	e.Start(":8080")
}
