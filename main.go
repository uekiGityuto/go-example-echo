package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/user", getUser)
	e.POST("/user", postUser)

	e.Logger.Fatal(e.Start(":1323"))
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func getUser(c echo.Context) error {
	u := User{
		Name:  "John",
		Email: "john@example.com",
	}
	return c.JSON(http.StatusOK, u)
}

func postUser(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}
