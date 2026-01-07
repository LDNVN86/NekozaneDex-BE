// Â A A A A A Bắn Tung Tóe Bắng Máy Trăm Phát

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success 	bool 			`json:"success"`
	Message 	string 			`json:"message,omitempty"`
	Data 		interface{} 	`json:"data,omitempty"`
	Error 		string 			`json:"error,omitempty"`
}

type Pagination struct {
	Success 	bool 			`json:"success"`
	Data 		interface{} 	`json:"data"`
	Meta 		Meta 			`json:"meta"`
}

type Meta struct {
	Page		int				`json:"page"`
	Limit		int				`json:"limit"`
	Total		int64			`json:"total"`
	TotalPages	int				`json:"total_pages"`
}

//Phản hồi Thành Công (Response Success)
func Oke(c *gin.Context, data interface{}){
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: data,
	})
}

func Created(c *gin.Context, data interface{}){
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Tạo Thành Công Rồi Nhé",
		Data: data,
	})
}

//Phân Trang - Pagination
func PaginatedResponse(c *gin.Context, data interface{},page,limit int, total int64){
	totalPages := int(total)/limit
	if int(total)%limit>0{
		totalPages++
	}
	c.JSON(http.StatusOK, Pagination{
		Success: true,
		Data: data,
		Meta: Meta{
			Page: page,
			Limit: limit,
			Total: total,
			TotalPages: totalPages,
		},
	})
}

//400 - Phản hồi Thất Bại (Response Error)
func BadRequest(c *gin.Context, message string){
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: message,
	})
}

//401 - unauthorized - không có quyền
func Unauthorized(c *gin.Context, message string){
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: message,
	})
}

//403 - Forbidden - Không có quyền
func Forbidden(c *gin.Context, message string){
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: message,
	})
}

//404 - Not Found - Không tìm thấy
func NotFound(c *gin.Context, message string){
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: message,
	})
}

//429 - Too Many Requests - Quá nhiều yêu cầu
func TooManyRequests(c *gin.Context, message string){
	c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Error: message,
	})
}

//409 - Conflict - Xung đột
func Conflict(c *gin.Context, message string){
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error: message,
	})
}

//413 - Payload Too Large - Yêu cầu quá lớn
func PayloadTooLarge(c *gin.Context, message string){
	c.JSON(http.StatusRequestEntityTooLarge, Response{
		Success: false,
		Error: message,
	})
}

//500 - Internal Server Error - Lỗi Server
func InternalServerError(c *gin.Context, message string){
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: message,
	})
}

//502 - Bad Gateway - Lỗi Gateway
func BadGateway(c *gin.Context, message string){
	c.JSON(http.StatusBadGateway, Response{
		Success: false,
		Error: message,
	})
}

//503 - Service Unavailable - Dịch vụ không có sẵn
func ServiceUnavailable(c *gin.Context, message string){
	c.JSON(http.StatusServiceUnavailable, Response{
		Success: false,
		Error: message,
	})
}
