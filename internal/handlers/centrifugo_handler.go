package handlers

import (
	"nekozanedex/internal/centrifugo"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CentrifugoHandler struct {
	client *centrifugo.Client
}

func NewCentrifugoHandler(client *centrifugo.Client) *CentrifugoHandler {
	return &CentrifugoHandler{
		client: client,
	}
}

func (h *CentrifugoHandler) GenerateConnectionToken(c *gin.Context){
	userID, exist := c.Get("user_id")
	if !exist {
		response.Unauthorized(c, "Chưa Đăng Nhập")
		return
	}

	token, err := h.client.GenerateConnectionToken(userID.(uuid.UUID).String())
	if err != nil {
		response.InternalServerError(c, "Không Thể Tạo Token")
		return
	}
	response.Oke(c, gin.H{"token": token})
}