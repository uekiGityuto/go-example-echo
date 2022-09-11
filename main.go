package main

import (
	validation "github.com/go-ozzo/ozzo-validation"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

var db *sqlx.DB

func main() {
	e := echo.New()
	var err error
	db, err = sqlx.Open("mysql", "member:@(127.0.0.1:3306)/practice")
	if err != nil {
		e.Logger.Fatal(err)
	}

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

const (
	Required      = "Required"
	Invalid       = "Invalid Data"
	Inconsistency = "Data Inconsistency"
)

type User struct {
	Id         uuid.UUID `json:"id" db:"id"`
	FamilyName string    `json:"family_name" db:"family_name"`
	GivenName  string    `json:"given_name" db:"given_name"`
	Age        int       `json:"age" db:"age"`
	Sex        string    `json:"sex" db:"sex"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.FamilyName, validation.Required.Error(Required)),
		validation.Field(&u.GivenName, validation.Required.Error(Required)),
		validation.Field(&u.Age, validation.Required.Error(Required), validation.Min(0).Error(Invalid)),
		validation.Field(&u.Sex, validation.Required.Error(Required), validation.In("男", "女").Error(Invalid)),
	)
}

type OK struct {
	Msg string `json:"message"`
}

func NewOK(msg string) OK {
	return OK{
		Msg: msg,
	}
}

type Error struct {
	Msg string `json:"error_message"`
}

func NewError(msg string) Error {
	return Error{
		Msg: msg,
	}
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
	ctx := c.Request().Context()
	id := c.QueryParam("id")
	// TODO: idのvalidation
	u := new(User)
	if err := db.GetContext(ctx, u, "SELECT * FROM user WHERE id = ?", id); err != nil {
		//c.Logger().Error(err)
		// TODO: システムエラーと見つからなかったエラーを分ける
		e := NewError("ユーザ情報取得が見つかりません")
		return c.JSON(http.StatusBadRequest, e)
	}

	return c.JSON(http.StatusOK, u)
}

func postUser(c echo.Context) error {
	ctx := c.Request().Context()
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

	id, err := uuid.NewRandom()
	if err != nil {
		c.Logger().Error(err)
		e := NewError("システムエラー")
		return c.JSON(http.StatusInternalServerError, e)
	}
	u.Id = id

	_, err = db.NamedExecContext(ctx, `INSERT INTO user (id, family_name, given_name, age, sex) VALUES (:id, :family_name, :given_name, :age, :sex)`, u)
	if err != nil {
		c.Logger().Error(err)
		e := NewError("システムエラー")
		return c.JSON(http.StatusInternalServerError, e)
	}

	ok := NewOK("完了")
	return c.JSON(http.StatusOK, ok)
}
