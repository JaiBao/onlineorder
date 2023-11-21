//handlers.go
package api

import (
    "time"
    "strconv"
    "net/http"
    "github.com/labstack/echo/v4"
)

// GetRoadsByCityID 根據城市 ID 獲取路名
func GetRoadsByCityID(c echo.Context) error {
    cityIDParam := c.QueryParam("city_id")
    cityID, err := strconv.Atoi(cityIDParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "無效的城市 ID"})
    }

    roads, err := FetchRoadsByCityID(cityID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "無法獲取路名數據"})
    }

    return c.JSON(http.StatusOK, roads)
}


// GetTimeSlotLimits 處理函數， 獲取所有時段限制
func GetTimeSlotLimits(c echo.Context) error {
    limits, err := FetchTimeSlotLimits()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "無法獲取時段限制資料"})
    }

    // 将切片转换为映射
    limitsMap := make(map[string]int)
    for _, limit := range limits {
        limitsMap[limit.TimeSlot] = limit.LimitCount
    }

    return c.JSON(http.StatusOK, limitsMap)
}



// GetSpecificDateLimits 處理函數，獲取特定日期的時段限制
func GetSpecificDateLimits(c echo.Context) error {
    yearMonth := c.QueryParam("month") // 從查詢參數中獲取月份，例如 "2024-01"
    specificDate := c.QueryParam("date") // 新增：從查詢參數中獲取具體日期，例如 "2024-01-02"

    allLimits, err := FetchSpecificDateLimits()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "無法獲取特定日期的時間限制數據"})
    }

    // 如果提供了具體日期參數，則只返回該日期的數據
    if specificDate != "" {
        yearMonthOfDate := specificDate[:7] // 從日期獲取年月
        if dates, ok := allLimits[yearMonthOfDate]; ok {
            if dateLimits, ok := dates[specificDate]; ok {
                return c.JSON(http.StatusOK, map[string]map[string]int{specificDate: dateLimits})
            }
        }
        return c.JSON(http.StatusNotFound, map[string]string{"error": "未找到指定日期的數據"})
    }

    // 如果提供了月份參數，則只返回該月份的數據
    if yearMonth != "" {
        if limits, ok := allLimits[yearMonth]; ok {
            return c.JSON(http.StatusOK, limits)
        }
        return c.JSON(http.StatusNotFound, map[string]string{"error": "未找到指定月份的數據"})
    }

    // 如果沒有提供月份或日期參數，返回所有數據
    return c.JSON(http.StatusOK, allLimits)
}




// CreateTimeSlotLimit  創建新的時段限制
func CreateTimeSlotLimit(c echo.Context) error {
    var limit TimeSlotLimit
    if err := c.Bind(&limit); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "格式錯誤"})
    }

    // 将单个对象转换为切片
    limits := TimeSlotLimits{limit}

    if err := InsertTimeSlotLimits(limits); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "無法創建時段限制"})
    }
    return c.JSON(http.StatusCreated, limit)
}


// UpdateTimeSlotLimit  更新現有的時段限制
func UpdateTimeSlotLimit(c echo.Context) error {
    var limits map[string]int
    if err := c.Bind(&limits); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "无效输入"})
    }

    if err := UpdateExistingTimeSlotLimits(limits); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "更新失败"})
    }
    return c.JSON(http.StatusOK, map[string]string{"result": "更新成功"})
}

// CreateSpecificDateLimit  創建特定日期的時段限制
func CreateSpecificDateLimit(c echo.Context) error {
    var dateLimits map[string]map[string]int
    if err := c.Bind(&dateLimits); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "格式錯誤"})
    }

    for date, limits := range dateLimits {
        dateLimit := SpecificDateLimit{
            Date:       date,
            TimeLimits: limits,
        }
        if err := InsertSpecificDateLimit(dateLimit); err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "無法創建特定日期的時段限制"})
        }
    }
    return c.JSON(http.StatusCreated, dateLimits)
}

// UpdateSpecificDateLimit 更新特定日期的一個時段或多個限制

func UpdateSpecificDateLimit(c echo.Context) error {
    var dateLimits map[string]map[string]int
    if err := c.Bind(&dateLimits); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "格式錯誤"})
    }

    for date, timeLimits := range dateLimits {
        for timeSlot, limitCount := range timeLimits {
            // 檢查原本是否有這時段紀錄
            if exists, err := checkDateLimitExists(date, timeSlot); err != nil {
                return c.JSON(http.StatusInternalServerError, map[string]string{"error": "檢查時發生錯誤"})
            } else if exists {
                // 有的話就改
                err := UpdateDateLimit(date, timeSlot, limitCount)
                if err != nil {
                    return c.JSON(http.StatusInternalServerError, map[string]string{"error": "加入失敗"})
                }
            } else {
                // 沒有的話給錯誤或插入新時段
                return c.JSON(http.StatusNotFound, map[string]string{"error": "沒有這個時段設定"})
                // 或者可以选择插入新记录
                // err := InsertDateLimit(date, timeSlot, limitCount)
                // if err != nil {
                //     return c.JSON(http.StatusInternalServerError, map[string]string{"error": "插入新時段"})
                // }
            }
        }
    }
    return c.JSON(http.StatusOK, map[string]string{"result": "已加入訂單"})
}

// checkDateLimitExists 檢查日期時段有無
func checkDateLimitExists(date string, timeSlot string) (bool, error) {
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM DateLimits WHERE Date = ? AND TimeSlot = ?)", date, timeSlot).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}


// UpdateDateLimit 更新特定時段
func UpdateDateLimit(date string, timeSlot string, limitCount int) error {
    stmt, err := db.Prepare("UPDATE DateLimits SET LimitCount = ? WHERE Date = ? AND TimeSlot = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(limitCount, date, timeSlot)
    if err != nil {
        return err
    }

    return nil
}


// TriggerAutoCreateLimits 自動創建預設日期
func TriggerAutoCreateLimits(c echo.Context) error {
    period := c.QueryParam("add") // 從查詢參數獲取時間範圍oneWeek、twoWeeks、oneMonth

    // 如果沒有提供時間範圍，則預設為兩個月
    if period == "" {
        period = "twoMonths"
    }

    err := AutoCreateNextTwoMonthsLimits(period)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, map[string]string{"message": "已成功新增時段預設"})
}


var schedulerActive bool
var ticker *time.Ticker

// StartSchedulerHandler 定時新增
func StartSchedulerHandler(c echo.Context) error {
    if !schedulerActive {
        StartScheduler()
        schedulerActive = true
        return c.JSON(http.StatusOK, map[string]string{"message": "定時自動新增已啟動"})
    }
    return c.JSON(http.StatusBadRequest, map[string]string{"error": "定時任務進行中"})
}

// StopSchedulerHandler 停止定時
func StopSchedulerHandler(c echo.Context) error {
    if schedulerActive {
        StopScheduler()
        schedulerActive = false
        return c.JSON(http.StatusOK, map[string]string{"message": "定時任務已停止"})
    }
    return c.JSON(http.StatusBadRequest, map[string]string{"error": "定時任務未運作"})
}

func StartScheduler() {
    ticker = time.NewTicker(24 * time.Hour)
    go func() {
        for {
            select {
            case <-ticker.C:
                AutoCreateNextTwoMonthsLimits("twoWeeks")
            }
        }
    }()
}

func StopScheduler() {
    if ticker != nil {
        ticker.Stop()
    }
}

// GetSchedulerStatusHandler 定時的狀態
func GetSchedulerStatusHandler(c echo.Context) error {
    status := map[string]bool{"schedulerActive": schedulerActive}
    return c.JSON(http.StatusOK, status)
}




