package item

import (
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client ItemClient
}

func NewHandler(client ItemClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) CreateWeaponHandler(c *gin.Context) {
	var req pb.CreateWeaponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Invalid request body",
			"error":      err.Error(),
		})
		return
	}

	weapon, err := h.client.CreateWeapon(c.Request.Context(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch st.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.AlreadyExists:
			httpStatus = http.StatusConflict
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    st.Message(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Weapon created successfully",
		"result":     weapon,
	})
}

func (h *Handler) ListWeaponsWithTemplateHandler(c *gin.Context) {
	response, err := h.client.ListWeaponsWithTemplate(c.Request.Context())
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    st.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Weapons retrieved successfully",
		"weapons":    response.Weapons,
	})
}

// CreateItemTemplateHandler 創建物品模板
// 重要：這個方法會觸發 RabbitMQ 事件，發送通知給管理員
func (h *Handler) CreateItemTemplateHandler(c *gin.Context) {
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	var httpReq struct {
		ItemName      string `json:"item_name" binding:"required"`
		ItemCode      string `json:"item_code" binding:"required"`
		TypeID        string `json:"type_id" binding:"required"`
		RarityID      string `json:"rarity_id" binding:"required"`
		ItemType      string `json:"item_type" binding:"required"`
		ItemID        string `json:"item_id" binding:"required"`
		IconURL       string `json:"icon_url"`
		IsTradeable   *bool  `json:"is_tradeable"`
		IsDroppable   *bool  `json:"is_droppable"`
		RequiredLevel *int32 `json:"required_level"`
		BaseSellPrice *int32 `json:"base_sell_price"`
		BaseBuyPrice  *int32 `json:"base_buy_price"`
	}

	if err := c.ShouldBindJSON(&httpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Invalid request body",
			"error":      err.Error(),
		})
		return
	}

	grpcReq := &pb.CreateItemTemplateRequest{
		UserId:   userIdStr.(string),
		ItemName: httpReq.ItemName,
		ItemCode: httpReq.ItemCode,
		TypeId:   httpReq.TypeID,
		RarityId: httpReq.RarityID,
		ItemType: httpReq.ItemType,
		ItemId:   httpReq.ItemID,
	}

	if httpReq.IconURL != "" {
		grpcReq.IconUrl = &httpReq.IconURL
	}
	if httpReq.IsTradeable != nil {
		grpcReq.IsTradeable = httpReq.IsTradeable
	}
	if httpReq.IsDroppable != nil {
		grpcReq.IsDroppable = httpReq.IsDroppable
	}
	if httpReq.RequiredLevel != nil {
		grpcReq.RequiredLevel = httpReq.RequiredLevel
	}
	if httpReq.BaseSellPrice != nil {
		grpcReq.BaseSellPrice = httpReq.BaseSellPrice
	}
	if httpReq.BaseBuyPrice != nil {
		grpcReq.BaseBuyPrice = httpReq.BaseBuyPrice
	}

	template, err := h.client.CreateItemTemplate(c.Request.Context(), grpcReq)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch st.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.AlreadyExists:
			httpStatus = http.StatusConflict
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    st.Message(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Item template created successfully. Notification sent to admins.",
		"result":     template,
	})
}
