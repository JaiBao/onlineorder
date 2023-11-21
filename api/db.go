//db.go
package api

import (
    "fmt"
    "os"
    "github.com/joho/godotenv"
    "time"
    "database/sql"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB() {
    // 載入環境變量
    err := godotenv.Load()
    if err != nil {
        log.Fatal("環境變量讀取失敗")
    }

   
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASS"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )

    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("已連結至資料庫")
}


// TimeSlotLimit  表示時段限制的結構
type TimeSlotLimit struct {
    TimeSlot   string `json:"time_slot"`
    LimitCount int    `json:"limit_count"`
}
// TimeSlotLimits 是多個 TimeSlotLimit 的切片
type TimeSlotLimits []TimeSlotLimit

// SpecificDateLimit  表示特定日期的時段限制結構
type SpecificDateLimit struct {
    Date       string                  `json:"date"`
    TimeLimits map[string]int          `json:"time_limits"`
}

// Road 表示路名和城市 ID 的結構
type Road struct {
    Name    string `json:"name"`
    CityID  int    `json:"city_id"`
}

// Roads 是多個 Road 的切片
type Roads []Road


// FetchRoadsByCityID 從數據庫中獲取特定城市 ID 的所有路名
func FetchRoadsByCityID(cityID int) (Roads, error) {
    var roads Roads

    // 查詢來選擇特定城市 ID 的路名
    rows, err := db.Query("SELECT name, city_id FROM roads WHERE city_id = ?", cityID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var road Road
        if err := rows.Scan(&road.Name, &road.CityID); err != nil {
            return nil, err
        }
        roads = append(roads, road)
    }

    return roads, nil
}


// FetchTimeSlotLimits 從數據庫中獲取所有時段限制
func FetchTimeSlotLimits() ([]TimeSlotLimit, error) {
    var limits []TimeSlotLimit
    rows, err := db.Query("SELECT TimeSlot, LimitCount FROM TimeSlotLimits")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var limit TimeSlotLimit
        if err := rows.Scan(&limit.TimeSlot, &limit.LimitCount); err != nil {
            return nil, err
        }
        limits = append(limits, limit)
    }

    return limits, nil
}

// FetchSpecificDateLimits 從數據庫中獲取特定日期的時段限制
func FetchSpecificDateLimits() (map[string]map[string]map[string]int, error) {
    today := time.Now().Format("2006-01-02")

    rows, err := db.Query("SELECT Date, TimeSlot, LimitCount FROM DateLimits WHERE Date >= ? ORDER BY Date, TimeSlot", today)
  
    // rows, err := db.Query("SELECT Date, TimeSlot, LimitCount FROM DateLimits ORDER BY Date, TimeSlot")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    monthlyLimits := make(map[string]map[string]map[string]int)
    var date, timeSlot string
    var limitCount int

    for rows.Next() {
        err := rows.Scan(&date, &timeSlot, &limitCount)
        if err != nil {
            return nil, err
        }

        // 分解日期以获取年份和月份
        yearMonth := date[:7] // 获取日期的前7个字符（例如，“2023-11”）

        if _, exists := monthlyLimits[yearMonth]; !exists {
            monthlyLimits[yearMonth] = make(map[string]map[string]int)
        }

        if _, dayExists := monthlyLimits[yearMonth][date]; !dayExists {
            monthlyLimits[yearMonth][date] = make(map[string]int)
        }

        monthlyLimits[yearMonth][date][timeSlot] = limitCount
    }

    return monthlyLimits, nil
}

func InsertTimeSlotLimits(limits TimeSlotLimits) error {
    for _, limit := range limits {
        stmt, err := db.Prepare("INSERT INTO TimeSlotLimits (TimeSlot, LimitCount) VALUES (?, ?)")
        if err != nil {
            return err 
        }

        _, err = stmt.Exec(limit.TimeSlot, limit.LimitCount)
        stmt.Close() 

        if err != nil {
            return err 
        }
    }
    return nil
}


func UpdateExistingTimeSlotLimits(limits map[string]int) error {
    for timeSlot, limitCount := range limits {
        stmt, err := db.Prepare("UPDATE TimeSlotLimits SET LimitCount = ? WHERE TimeSlot = ?")
        if err != nil {
            return err
        }
        defer stmt.Close()

        _, err = stmt.Exec(limitCount, timeSlot)
        if err != nil {
            return err
        }
    }
    return nil
}

//增加設定日期
func insertDateIfNeeded(date string) error {
    // 檢查日期是否已存在
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Dates WHERE Date = ?)", date).Scan(&exists)
    if err != nil {
        return err
    }

    // 如果日期不存在，則插入
    if !exists {
        stmt, err := db.Prepare("INSERT INTO Dates (Date) VALUES (?)")
        if err != nil {
            return err
        }
        defer stmt.Close()

        _, err = stmt.Exec(date)
        if err != nil {
            return err
        }
    }

    return nil
}
func InsertSpecificDateLimit(dateLimit SpecificDateLimit) error {
    // 首先檢查並可能插入日期到Dates表
    if err := insertDateIfNeeded(dateLimit.Date); err != nil {
        return err 
    }

    // 然後插入時段限制到DateLimits表
    for timeSlot, limit := range dateLimit.TimeLimits {
        stmt, err := db.Prepare("INSERT INTO DateLimits (Date, TimeSlot, LimitCount) VALUES (?, ?, ?)")
        if err != nil {
            return err 
        }

        _, err = stmt.Exec(dateLimit.Date, timeSlot, limit)
        stmt.Close() // 立即關閉語句

        if err != nil {
            return err 
        }
    }
    return nil
}




func UpdateExistingSpecificDateLimit(dateLimit SpecificDateLimit) error {
    for timeSlot, limit := range dateLimit.TimeLimits {
        stmt, err := db.Prepare("UPDATE DateLimits SET LimitCount = ? WHERE Date = ? AND TimeSlot = ?")
        if err != nil {
            return err
        }
        defer stmt.Close()

        _, err = stmt.Exec(limit, dateLimit.Date, timeSlot)
        if err != nil {
            return err
        }
    }
    return nil
}

// AutoCreateNextTwoMonthsLimits 自動新增特定時間範圍的限制
func AutoCreateNextTwoMonthsLimits(period string) error {
    // 取得現有設定日期
    existingLimits, err := FetchSpecificDateLimits()
    if err != nil {
        return err
    }

    // 取得預設
    initialLimits, err := FetchTimeSlotLimits()
    if err != nil {
        return err
    }

    // 轉格式
    initialTimeLimits := make(map[string]int)
    for _, limit := range initialLimits {
        initialTimeLimits[limit.TimeSlot] = limit.LimitCount
    }

    // 時間範圍
    startDate := time.Now()
    var endDate time.Time
    switch period {
    case "oneWeek":
        endDate = startDate.AddDate(0, 0, 7)
    case "twoWeeks":
        endDate = startDate.AddDate(0, 0, 14)
    case "oneMonth":
        endDate = startDate.AddDate(0, 1, 0)
    case "twoMonths":
        endDate = startDate.AddDate(0, 2, 0)
    default:
        return fmt.Errorf("不支援的時間範圍: %s", period)
    }

    // 日期循環
    for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
        dateStr := d.Format("2006-01-02")

        // 檢查有無設定
        if _, exists := existingLimits[dateStr]; !exists {
            dateLimit := SpecificDateLimit{
                Date:       dateStr,
                TimeLimits: initialTimeLimits,
            }

            // 沒有設定的插入預設
            if err := InsertSpecificDateLimit(dateLimit); err != nil {
                return err
            }
        }
    }

    return nil
}




