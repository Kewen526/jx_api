package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
)

var db *sql.DB

// Common columns shared by both tables, matching the frontend response format.
var commonColumns = []string{
	"id", "review_id", "shop_id", "shop_name", "city_name", "city_id",
	"user_id", "user_nickname", "user_face", "user_power",
	"add_time", "update_time", "star", "star_display",
	"content", "pic_count", "video_count", "pic_info",
	"shop_reply", "shop_reply_time", "reply_list",
	"raw_data", "ai_gen", "manual_confirm", "task_reply",
	"created_at", "updated_at",
}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	api := r.Group("/api/new")
	{
		api.GET("/reviews/dianping", handleReviews("review_detail_dianping", "dianping"))
		api.GET("/reviews/meituan", handleReviews("review_detail_meituan", "meituan"))
		api.GET("/reviews/all", handleAllReviews)
	}

	log.Println("Server starting on :8080")
	if err := r.Run("127.0.0.1:8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// parseParams extracts and validates shop_ids (comma-separated), page, page_size from query string.
// Also supports legacy single shop_id parameter for backward compatibility.
func parseParams(c *gin.Context) (shopIDs []int64, page, pageSize int, err error) {
	shopIDsStr := c.Query("shop_ids")
	if shopIDsStr == "" {
		shopIDsStr = c.Query("shop_id")
	}
	if shopIDsStr == "" {
		err = fmt.Errorf("shop_ids is required")
		return
	}

	parts := strings.Split(shopIDsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, parseErr := strconv.ParseInt(p, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("invalid shop_id: %s", p)
			return
		}
		shopIDs = append(shopIDs, id)
	}
	if len(shopIDs) == 0 {
		err = fmt.Errorf("shop_ids is required")
		return
	}

	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "30"))
	switch pageSize {
	case 30, 50, 100:
	default:
		pageSize = 30
	}

	return
}

// inPlaceholders returns "?, ?, ?" for n items.
func inPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}

// idsToArgs converts []int64 to []interface{} for use as SQL arguments.
func idsToArgs(ids []int64) []interface{} {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}

// handleReviews returns a handler for single-table queries (supports multiple shop_ids).
func handleReviews(table, platform string) gin.HandlerFunc {
	cols := strings.Join(commonColumns, ", ")

	return func(c *gin.Context) {
		shopIDs, page, pageSize, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		ph := inPlaceholders(len(shopIDs))
		idArgs := idsToArgs(shopIDs)

		countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE shop_id IN (%s)", table, ph)
		var total int
		if err := db.QueryRow(countSQL, idArgs...).Scan(&total); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "database error"})
			log.Printf("count error [%s]: %v", table, err)
			return
		}

		totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
		offset := (page - 1) * pageSize

		querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE shop_id IN (%s) ORDER BY add_time DESC LIMIT ? OFFSET ?", cols, table, ph)
		queryArgs := append(idArgs, pageSize, offset)
		rows, err := db.Query(querySQL, queryArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "database error"})
			log.Printf("query error [%s]: %v", table, err)
			return
		}
		defer rows.Close()

		data := scanRows(rows, platform)

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"data":        data,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		})
	}
}

// handleAllReviews queries both tables and merges results (supports multiple shop_ids).
func handleAllReviews(c *gin.Context) {
	shopIDs, page, pageSize, err := parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	cols := strings.Join(commonColumns, ", ")
	ph := inPlaceholders(len(shopIDs))
	idArgs := idsToArgs(shopIDs)

	// Count total from both tables.
	countSQL := fmt.Sprintf(`SELECT
		(SELECT COUNT(*) FROM review_detail_dianping WHERE shop_id IN (%s)) +
		(SELECT COUNT(*) FROM review_detail_meituan WHERE shop_id IN (%s))`, ph, ph)
	countArgs := append(idArgs, idArgs...)
	var total int
	if err := db.QueryRow(countSQL, countArgs...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "database error"})
		log.Printf("count error [all]: %v", err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	offset := (page - 1) * pageSize

	// UNION ALL with pagination on the merged result.
	querySQL := fmt.Sprintf(`SELECT * FROM (
		SELECT %s, 'dianping' AS platform FROM review_detail_dianping WHERE shop_id IN (%s)
		UNION ALL
		SELECT %s, 'meituan' AS platform FROM review_detail_meituan WHERE shop_id IN (%s)
	) AS combined ORDER BY add_time DESC LIMIT ? OFFSET ?`, cols, ph, cols, ph)
	queryArgs := append(idArgs, idArgs...)
	queryArgs = append(queryArgs, pageSize, offset)

	rows, err := db.Query(querySQL, queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "database error"})
		log.Printf("query error [all]: %v", err)
		return
	}
	defer rows.Close()

	data := scanRowsAll(rows)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// scanRows reads rows from a single-table query and returns a slice of maps.
func scanRows(rows *sql.Rows, platform string) []map[string]interface{} {
	var results []map[string]interface{}

	for rows.Next() {
		var (
			id           int64
			reviewID     string
			shopID       int64
			shopName     sql.NullString
			cityName     sql.NullString
			cityID       sql.NullInt64
			userID       sql.NullString
			userNickname sql.NullString
			userFace     sql.NullString
			userPower    sql.NullString
			addTime      sql.NullTime
			updateTime   sql.NullTime
			star         sql.NullInt64
			starDisplay  sql.NullString
			content      sql.NullString
			picCount     sql.NullInt64
			videoCount   sql.NullInt64
			picInfo      sql.NullString
			shopReply    sql.NullString
			shopReplyTime sql.NullTime
			replyList    sql.NullString
			rawData      sql.NullString
			aiGen        sql.NullString
			manualConfirm int
			taskReply    int
			createdAt    sql.NullTime
			updatedAt    sql.NullTime
		)

		err := rows.Scan(
			&id, &reviewID, &shopID, &shopName, &cityName, &cityID,
			&userID, &userNickname, &userFace, &userPower,
			&addTime, &updateTime, &star, &starDisplay,
			&content, &picCount, &videoCount, &picInfo,
			&shopReply, &shopReplyTime, &replyList,
			&rawData, &aiGen, &manualConfirm, &taskReply,
			&createdAt, &updatedAt,
		)
		if err != nil {
			log.Printf("scan error: %v", err)
			continue
		}

		row := buildRow(
			id, reviewID, shopID, shopName, cityName, cityID,
			userID, userNickname, userFace, userPower,
			addTime, updateTime, star, starDisplay,
			content, picCount, videoCount, picInfo,
			shopReply, shopReplyTime, replyList,
			rawData, aiGen, manualConfirm, taskReply,
			createdAt, updatedAt, platform,
		)
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	return results
}

// scanRowsAll reads rows from the UNION ALL query (includes platform column).
func scanRowsAll(rows *sql.Rows) []map[string]interface{} {
	var results []map[string]interface{}

	for rows.Next() {
		var (
			id           int64
			reviewID     string
			shopID       int64
			shopName     sql.NullString
			cityName     sql.NullString
			cityID       sql.NullInt64
			userID       sql.NullString
			userNickname sql.NullString
			userFace     sql.NullString
			userPower    sql.NullString
			addTime      sql.NullTime
			updateTime   sql.NullTime
			star         sql.NullInt64
			starDisplay  sql.NullString
			content      sql.NullString
			picCount     sql.NullInt64
			videoCount   sql.NullInt64
			picInfo      sql.NullString
			shopReply    sql.NullString
			shopReplyTime sql.NullTime
			replyList    sql.NullString
			rawData      sql.NullString
			aiGen        sql.NullString
			manualConfirm int
			taskReply    int
			createdAt    sql.NullTime
			updatedAt    sql.NullTime
			platform     string
		)

		err := rows.Scan(
			&id, &reviewID, &shopID, &shopName, &cityName, &cityID,
			&userID, &userNickname, &userFace, &userPower,
			&addTime, &updateTime, &star, &starDisplay,
			&content, &picCount, &videoCount, &picInfo,
			&shopReply, &shopReplyTime, &replyList,
			&rawData, &aiGen, &manualConfirm, &taskReply,
			&createdAt, &updatedAt, &platform,
		)
		if err != nil {
			log.Printf("scan error: %v", err)
			continue
		}

		row := buildRow(
			id, reviewID, shopID, shopName, cityName, cityID,
			userID, userNickname, userFace, userPower,
			addTime, updateTime, star, starDisplay,
			content, picCount, videoCount, picInfo,
			shopReply, shopReplyTime, replyList,
			rawData, aiGen, manualConfirm, taskReply,
			createdAt, updatedAt, platform,
		)
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	return results
}

// buildRow constructs a single result map matching the frontend JSON format.
func buildRow(
	id int64, reviewID string, shopID int64,
	shopName, cityName sql.NullString, cityID sql.NullInt64,
	userID, userNickname, userFace, userPower sql.NullString,
	addTime, updateTime sql.NullTime,
	star sql.NullInt64, starDisplay sql.NullString,
	content sql.NullString,
	picCount, videoCount sql.NullInt64,
	picInfo sql.NullString,
	shopReply sql.NullString, shopReplyTime sql.NullTime,
	replyList sql.NullString,
	rawData sql.NullString,
	aiGen sql.NullString,
	manualConfirm, taskReply int,
	createdAt, updatedAt sql.NullTime,
	platform string,
) map[string]interface{} {
	const timeFmt = "2006-01-02 15:04:05"

	row := map[string]interface{}{
		"id":             id,
		"review_id":      reviewID,
		"shop_id":        shopID,
		"shop_name":      nullStr(shopName),
		"city_name":      nullStr(cityName),
		"city_id":        nullInt(cityID),
		"user_id":        nullStr(userID),
		"user_nickname":  nullStr(userNickname),
		"user_face":      nullStr(userFace),
		"user_power":     nullStr(userPower),
		"add_time":       nullTimeStr(addTime, timeFmt),
		"update_time":    nullTimeStr(updateTime, timeFmt),
		"star":           nullInt(star),
		"star_display":   nullStr(starDisplay),
		"content":        nullStr(content),
		"pic_count":      nullInt(picCount),
		"video_count":    nullInt(videoCount),
		"pic_info":       parseJSON(picInfo),
		"shop_reply":     nullStr(shopReply),
		"shop_reply_time": nullTimeStr(shopReplyTime, timeFmt),
		"reply_list":     parseJSON(replyList),
		"raw_data":       parseJSON(rawData),
		"ai_gen":         nullStr(aiGen),
		"manual_confirm": manualConfirm,
		"task_reply":     taskReply,
		"created_at":     nullTimeStr(createdAt, timeFmt),
		"updated_at":     nullTimeStr(updatedAt, timeFmt),
		"platform":       platform,
	}
	return row
}

func nullStr(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func nullInt(ni sql.NullInt64) interface{} {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

func nullTimeStr(nt sql.NullTime, layout string) interface{} {
	if nt.Valid {
		return nt.Time.Format(layout)
	}
	return nil
}

// parseJSON attempts to unmarshal a JSON string; returns nil on failure.
func parseJSON(ns sql.NullString) interface{} {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	s := strings.TrimSpace(ns.String)

	// Try array first.
	if strings.HasPrefix(s, "[") {
		var arr []interface{}
		if json.Unmarshal([]byte(s), &arr) == nil {
			return arr
		}
	}
	// Try object.
	if strings.HasPrefix(s, "{") {
		var obj map[string]interface{}
		if json.Unmarshal([]byte(s), &obj) == nil {
			return obj
		}
	}
	return s
}
