// routes.go

package api

import (
	"github.com/labstack/echo/v4"
)

// LoadRoutes  加載 API 路由
func LoadRoutes(e *echo.Echo) {

	e.GET("/get-timeslot", GetTimeSlotLimits)
	e.GET("/get-road", GetRoadsByCityID)
	e.GET("/get-special", GetSpecificDateLimits)
	e.POST("/add-timeslot",CreateTimeSlotLimit)
	e.POST("/add-special",CreateSpecificDateLimit)
	e.PUT("/update-timeslot",UpdateTimeSlotLimit)
	e.PUT("/add-order",UpdateSpecificDateLimit)
	e.POST("/auto-add", TriggerAutoCreateLimits)
	e.POST("/start-scheduler", StartSchedulerHandler)
e.POST("/stop-scheduler", StopSchedulerHandler)
e.GET("/scheduler-status", GetSchedulerStatusHandler)




}
