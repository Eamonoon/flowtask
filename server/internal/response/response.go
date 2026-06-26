package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Data:    data,
		Message: "created",
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Data:    nil,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 40001, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, 40101, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, 40301, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, 40401, message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, 40901, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, 50001, message)
}
