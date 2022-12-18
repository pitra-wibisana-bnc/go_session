package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"

	_ "github.com/go-sql-driver/mysql"

	"go_session/models"
)

type M map[string]interface{}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Failed load .env : ", err.Error())
	} else {

		// Init Engine DB
		var dbConnection string = os.Getenv("DB_CONNECTION")
		var dbHost string = os.Getenv("DB_HOST")
		var dbPort string = os.Getenv("DB_PORT")
		var dbName string = os.Getenv("DB_DATABASE")
		var dbUser string = os.Getenv("DB_USERNAME")
		var dbPass string = os.Getenv("DB_PASSWORD")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)
		dbEngine, err := xorm.NewEngine(dbConnection, dsn)
		if err != nil {
			fmt.Println("dsn: ", dsn)
			fmt.Println("Failed Connect DB : ", err.Error())
			return
		} else {
			fmt.Println("DB Connected")
			loc, err := time.LoadLocation("Asia/Jakarta")
			if err == nil {
				dbEngine.SetTZDatabase(loc)
				dbEngine.SetTZLocation(loc)
			}
		}

		// Init Engine Session
		var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
		_ = store

		// Init Echo Web Server
		e := echo.New()

		// Handle memory leak
		e.Use(echo.WrapMiddleware(context.ClearHandler))

		api := e.Group("/api")
		auth := api.Group("/auth")

		auth.POST("/do_login", func(c echo.Context) error {

			username := c.FormValue("username")
			password := c.FormValue("password")

			if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Please fill username and password",
				})
			}

			// Get By username
			var allusers []models.Users
			err := dbEngine.Where("username = ?", username).Limit(1, 0).Find(&allusers)

			if err != nil {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Username not found",
				})
			} else if len(allusers) == 0 {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Username not found",
				})
			} else if allusers[0].Password != password {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Password not match",
				})
			}

			// get session by os.Getenv("SESSION_KEY")
			session, _ := store.Get(c.Request(), os.Getenv("SESSION_KEY"))

			session.Values["username"] = allusers[0].Username
			session.Values["first_name"] = allusers[0].FirstName
			session.Values["last_name"] = allusers[0].LastName

			// store session
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return c.JSON(http.StatusOK, M{
				"status":  "OK",
				"message": "Login success",
			})
		})

		auth.POST("/do_logout", func(c echo.Context) error {
			// process get session
			session, _ := store.Get(c.Request(), os.Getenv("SESSION_KEY"))

			// process to expired session
			session.Options.MaxAge = -1
			session.Save(c.Request(), c.Response())

			return c.JSON(http.StatusOK, M{
				"status":  "OK",
				"message": "Logout success",
			})
		})

		auth.GET("/current_session", func(c echo.Context) error {
			// process get session
			session, _ := store.Get(c.Request(), os.Getenv("SESSION_KEY"))

			// validate session is no data
			if len(session.Values) == 0 {
				return c.JSON(http.StatusOK, M{
					"status": "fail",
					"error":  "no sessions",
				})
			}

			return c.JSON(http.StatusOK, M{
				"status": "OK",
				"data": M{
					"username":   session.Values["username"],
					"first_name": session.Values["first_name"],
					"last_name":  session.Values["last_name"],
				},
			})
		})

		api.POST("/register", func(c echo.Context) error {

			username := c.FormValue("username")
			password := c.FormValue("password")
			first_name := c.FormValue("first_name")
			last_name := c.FormValue("last_name")

			if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" || strings.TrimSpace(first_name) == "" || strings.TrimSpace(last_name) == "" {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Please fill all form input",
				})
			} else {
				_, err := mail.ParseAddress(username)
				if err != nil {
					return c.JSON(http.StatusBadRequest, M{
						"status":  "fail",
						"message": "Please fill username with email",
					})
				}
			}

			// Get By username
			var allusers []models.Users
			err := dbEngine.Where("username = ?", username).Limit(1, 0).Find(&allusers)

			if err != nil {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Username not found",
				})
			} else if len(allusers) > 0 {
				return c.JSON(http.StatusBadRequest, M{
					"status":  "fail",
					"message": "Username " + username + " already exists",
				})
			}

			row_user := new(models.Users)
			row_user.Username = username
			row_user.Password = password
			row_user.FirstName = first_name
			row_user.LastName = last_name

			affected, err := dbEngine.Insert(row_user)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, M{
					"status":  "fail",
					"message": "Error " + err.Error(),
				})
			} else {
				fmt.Println("affected:", affected)

				return c.JSON(http.StatusOK, M{
					"status":  "OK",
					"message": "Register success",
				})
			}
		})

		// Static Asset
		e.Static("/assets", "static/assets")

		// Static HTML
		e.File("/", "static/view/index.html")
		e.File("/dashboard", "static/view/dashboard.html")
		e.File("/register", "static/view/register.html")

		e.Start(":4444")
	}
}
