package main

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func main() {
	e := echo.New()

	e.Validator = new(CustomValidator)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/user", getUser)
	e.POST("/user", postUser)

	e.Logger.Fatal(e.Start(":1323"))
}

type CustomValidator struct{}

func (cv *CustomValidator) Validate(i interface{}) error {
	if v, ok := i.(validation.Validatable); ok {
		return v.Validate()
	}
	return nil
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Name, validation.Required),
		validation.Field(&u.Email, validation.Required),
	)
}

type ValidationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func NewValidationError(field string, reason string) ValidationError {
	return ValidationError{
		Field:  field,
		Reason: reason,
	}
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
	if err := c.Validate(u); err != nil {
		errs := err.(validation.Errors)
		var validationErrors []ValidationError
		for k, err := range errs {
			validationErrors = append(validationErrors, NewValidationError(k, err.Error()))
		}
		return c.JSON(http.StatusBadRequest, validationErrors)
	}

	return c.JSON(http.StatusOK, u)
}
