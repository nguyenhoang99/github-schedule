package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Hoang_0610"
	dbname   = "teaching_schedule"
)

type CreateClassRequest struct {
	Classname string `json:"classname"`
}

type UpdateClassRequest struct {
	Id        int    `json:"id"`
	Classname string `json:"classname"`
}

type DeleteClassRequest struct {
	Id int `json:"id"`
}

type Class struct {
	Id        int    `json:"id"`
	Classname string `json:"classname"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Teacher struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	TeacherName string `json:"teacher_name"`
}

type Claims struct {
	Username string `json:"username"`
	Id       int    `json:"id"`
	jwt.StandardClaims
}

var jwtKey = []byte("m_byte_123")

type UserSignUpRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	TeacherName string `json:"teacher_name"`
}

type CreateScheduleRequest struct {
	ClassId   int    `json:"class_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type UpdateScheduleRequest struct {
	Id        int    `json:"id"`
	ClassId   int    `json:"class_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type DeleteScheduleRequest struct {
	Id int `json:"id"`
}

type TeachingSchedule struct {
	Id        int    `json:"id"`
	TeacherId int    `json:"teacher_id"`
	ClassId   int    `json:"class_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	r := gin.Default()
	r.Use(authMiddleware())
	r.POST("/login", func(c *gin.Context) {
		var userLoginRequest UserLoginRequest
		if err := c.ShouldBindJSON(&userLoginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := fmt.Sprintf("SELECT COUNT(*)FROM teachers WHERE username ='%s' and password = '%s'", userLoginRequest.Username, userLoginRequest.Password)
		row, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		var count int
		row.Next()
		err = row.Scan(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "wrong username or password",
			})
		} else {
			var teacherInfo Teacher
			query = fmt.Sprintf("SELECT *FROM teachers WHERE username = '%s'", userLoginRequest.Username)
			row, err := db.Query(query)
			defer row.Close()
			row.Next()
			err = row.Scan(&teacherInfo.Id, &teacherInfo.Username, &teacherInfo.Password, &teacherInfo.TeacherName)
			if err != nil {
				log.Fatal(err)
			}
			expirationTime := time.Now().Add(5 * time.Minute)
			claims := &Claims{
				Id:       teacherInfo.Id,
				Username: userLoginRequest.Username,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(jwtKey)
			if err != nil {
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, gin.H{
				"token": tokenString,
			})
		}
	})
	r.POST("/signup", func(c *gin.Context) {
		var userSignUpRequest UserSignUpRequest
		if err := c.ShouldBindJSON(&userSignUpRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := fmt.Sprintf("SELECT COUNT(*)FROM teachers WHERE username = '%s'", userSignUpRequest.Username)
		row, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer row.Close()
		var count int
		row.Next()
		err = row.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
		if count == 0 {
			query = fmt.Sprintf("INSERT INTO teachers(username, password, teacher_name)  VALUES ('%s', '%s', '%s')", userSignUpRequest.Username, userSignUpRequest.Password, userSignUpRequest.TeacherName)
			_, err = db.Exec(query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "account already exists",
			})
		}
	})
	r.POST("/create-class", func(c *gin.Context) {
		var createClassRequest CreateClassRequest
		if err := c.ShouldBindJSON(&createClassRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := fmt.Sprintf("INSERT INTO class(classname) VALUES ('%s')", createClassRequest.Classname)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})

	})
	r.PUT("/update-class", func(c *gin.Context) {
		var updateClassRequest UpdateClassRequest
		if err := c.ShouldBindJSON(&updateClassRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := fmt.Sprintf("SELECT COUNT(*)FROM class WHERE id = %d ", updateClassRequest.Id)
		row, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer row.Close()
		row.Next()
		var count int
		err = row.Scan(&count)
		if count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "User wrong update",
			})
			return
		}
		query = fmt.Sprintf("UPDATE class SET classname = '%s' WHERE id = %d", updateClassRequest.Classname, updateClassRequest.Id)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.DELETE("/delete-class", func(c *gin.Context) {
		var deleteClassRequest DeleteClassRequest
		if err := c.ShouldBindJSON(&deleteClassRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var count int
		query := fmt.Sprintf("SELECT COUNT(*)FROM class WHERE id = %d", deleteClassRequest.Id)
		row, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer row.Close()
		row.Next()
		err = row.Scan(&count)
		if count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "User not delete",
			})
			return
		}
		query = fmt.Sprintf("DELETE FROM class WHERE id = %d", deleteClassRequest.Id)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.GET("/get-class", func(c *gin.Context) {
		classes := make([]*Class, 0)
		rows, err := db.Query("SELECT  *FROM class ")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			class := new(Class)
			if err := rows.Scan(&class.Id, &class.Classname); err != nil {
				panic(err)
			}
			classes = append(classes, class)
		}
		c.JSON(http.StatusOK, classes)

	})
	r.POST("/create-schedule", func(c *gin.Context) {
		id, ok := c.Get("id")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Wrong userid",
			})
		}
		var createScheduleRequest CreateScheduleRequest
		if err := c.ShouldBindJSON(&createScheduleRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		check := checkTimeCreate(db, createScheduleRequest.StartTime, createScheduleRequest.EndTime, id.(int))
		if !check {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Trung thoi gian"})
			return
		}
		query := fmt.Sprintf("INSERT INTO teaching_schedule(teacher_id, class_id, start_time, end_time)"+
			"VALUES(%d, %d, '%s', '%s')", id, createScheduleRequest.ClassId, createScheduleRequest.StartTime, createScheduleRequest.EndTime)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.PUT("/update-schedule", func(c *gin.Context) {
		id, ok := c.Get("id")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Wrong id",
			})
		}
		var updateScheduleRequest UpdateScheduleRequest
		if err := c.ShouldBindJSON(&updateScheduleRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := fmt.Sprintf("SELECT COUNT(*)FROM teaching_schedule WHERE id = %d and teacher_id = %d", updateScheduleRequest.Id, id)
		row, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer row.Close()
		row.Next()
		var count int
		err = row.Scan(&count)
		if count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Teaching Schedule wrong update",
			})
			return
		}
		check := checkTimeUpdate(db, updateScheduleRequest.StartTime, updateScheduleRequest.EndTime, id.(int))
		if !check {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Trung thoi gian"})
			return
		}
		query = fmt.Sprintf("UPDATE teaching_schedule SET class_id = %d, start_time = '%s', end_time = '%s' WHERE id = %d and teacher_id = %d",
			updateScheduleRequest.ClassId, updateScheduleRequest.StartTime, updateScheduleRequest.EndTime, updateScheduleRequest.Id, id)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.DELETE("/delete-schedule", func(c *gin.Context) {
		id, ok := c.Get("id")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Wrong id",
			})
		}
		var deleteScheduleRequest DeleteScheduleRequest
		if err := c.ShouldBindJSON(&deleteScheduleRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := fmt.Sprintf("SELECT COUNT(*)FROM teaching_schedule WHERE id = %d and teacher_id = %d", deleteScheduleRequest.Id, id)
		row, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer row.Close()
		row.Next()
		var count int
		err = row.Scan(&count)
		if count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Teaching Schedule wrong delete",
			})
			return
		}
		query = fmt.Sprintf("DELETE FROM teaching_schedule WHERE id = %d and teacher_id = %d", deleteScheduleRequest.Id, id)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.GET("/get-schedule", func(c *gin.Context) {
		id, ok := c.Get("id")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Wrong id",
			})
		}
		teachingSchedules := make([]*TeachingSchedule, 0)
		rows, err := db.Query("SELECT *FROM teaching_schedule WHERE teacher_id = $1", id)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			teachingSchedule := new(TeachingSchedule)
			if err := rows.Scan(&teachingSchedule.Id, &teachingSchedule.TeacherId, &teachingSchedule.ClassId, &teachingSchedule.StartTime, &teachingSchedule.EndTime); err != nil {
				panic(err)
			}
			teachingSchedules = append(teachingSchedules, teachingSchedule)
		}
		c.JSON(http.StatusOK, teachingSchedules)
	})
	r.Run()
}
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/login" || c.Request.RequestURI == "/signup" {
			t := time.Now()
			c.Set("example", "schedule")
			c.Next()
			latency := time.Since(t)
			log.Print(latency)
		} else {
			tokenString := c.Request.Header["Token"][0]
			claims := &Claims{}
			tkn, _ := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			if tkn.Valid {
				log.Print("is valid token")
				t := time.Now()
				c.Set("id", claims.Id)
				c.Next()
				latency := time.Since(t)
				log.Print(latency)
				return
			} else {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"message": "invalid token",
				})
			}
		}
	}
}

func checkTimeCreate(db *sql.DB, start string, end string, teacherId int) bool {
	query := fmt.Sprintf("SELECT COUNT(*)FROM teaching_schedule WHERE "+
		"((start_time < '%s' and end_time >= '%s') or (start_time <= '%s' and  end_time >= '%s') or (end_time > '%s' and start_time <= '%s')) and teacher_id = %d ", end, end, start, end, start, start, teacherId)
	fmt.Println(query)
	row, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	var count int
	defer row.Close()
	row.Next()
	err = row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		return false
	}
	return true
}

func checkTimeUpdate(db *sql.DB, start string, end string, teacherId int) bool {
	query := fmt.Sprintf("SELECT COUNT(*)FROM teaching_schedule WHERE "+
		"((start_time < '%s' and end_time >= '%s') or (start_time <= '%s' and  end_time >= '%s') or (end_time > '%s' and start_time <= '%s')) and teacher_id = %d", end, end, start, end, start, start, teacherId)
	fmt.Println(query)
	row, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	var count int
	defer row.Close()
	row.Next()
	err = row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		return false
	}
	return true
}
