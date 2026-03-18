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

// CreateCompleteWeaponHandler creates a complete weapon (weapon + template) in one request
func (h *Handler) CreateCompleteWeaponHandler(c *gin.Context) {
	// 1. Extract userId from JWT
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	// 2. Parse JSON body
	var httpReq struct {
		// Template fields
		ItemName      string  `json:"item_name" binding:"required"`
		ItemCode      string  `json:"item_code" binding:"required"`
		IconURL       *string `json:"icon_url"`
		RequiredLevel *int32  `json:"required_level"`
		BaseSellPrice *int32  `json:"base_sell_price"`
		BaseBuyPrice  *int32  `json:"base_buy_price"`

		// Weapon fields
		TypeID       string  `json:"type_id" binding:"required"`
		RarityID     string  `json:"rarity_id" binding:"required"`
		AttackPower  int32   `json:"attack_power" binding:"required"`
		Durability   int32   `json:"durability" binding:"required"`
		CriticalRate float32 `json:"critical_rate"`
		WeaponType   string  `json:"weapon_type"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&httpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Invalid request body",
			"error":      err.Error(),
		})
		return
	}

	// 3. Build gRPC request
	grpcReq := &pb.CreateCompleteWeaponRequest{
		UserId:       userIdStr.(string),
		ItemName:     httpReq.ItemName,
		ItemCode:     httpReq.ItemCode,
		TypeId:       httpReq.TypeID,
		RarityId:     httpReq.RarityID,
		AttackPower:  httpReq.AttackPower,
		Durability:   httpReq.Durability,
		CriticalRate: httpReq.CriticalRate,
		WeaponType:   httpReq.WeaponType,
		Description:  httpReq.Description,
	}

	// Handle optional fields
	if httpReq.IconURL != nil {
		grpcReq.IconUrl = httpReq.IconURL
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

	// 4. Call items-service (will create weapon, create template, send RabbitMQ)
	weaponDetail, err := h.client.CreateCompleteWeapon(c.Request.Context(), grpcReq)
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

	// 5. Success response
	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Complete weapon created successfully. Notification sent to admins.",
		"result":     weaponDetail,
	})
}

// CreateCompleteArmorHandler creates a complete armor (armor + template) in one request
func (h *Handler) CreateCompleteArmorHandler(c *gin.Context) {
	// 1. Extract userId from JWT
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	// 2. Parse JSON body
	var httpReq struct {
		// Template fields
		ItemName      string  `json:"item_name" binding:"required"`
		ItemCode      string  `json:"item_code" binding:"required"`
		IconURL       *string `json:"icon_url"`
		RequiredLevel *int32  `json:"required_level"`
		BaseSellPrice *int32  `json:"base_sell_price"`
		BaseBuyPrice  *int32  `json:"base_buy_price"`

		// Armor fields
		TypeID          string `json:"type_id" binding:"required"`
		RarityID        string `json:"rarity_id" binding:"required"`
		DefenseRating   int32  `json:"defense_rating" binding:"required"`
		Durability      int32  `json:"durability" binding:"required"`
		MagicResistance int32  `json:"magic_resistance"`
		ArmorSlot       string `json:"armor_slot"`
		Description     string `json:"description"`
	}

	if err := c.ShouldBindJSON(&httpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Invalid request body",
			"error":      err.Error(),
		})
		return
	}

	// 3. Build gRPC request
	grpcReq := &pb.CreateCompleteArmorRequest{
		UserId:          userIdStr.(string),
		ItemName:        httpReq.ItemName,
		ItemCode:        httpReq.ItemCode,
		TypeId:          httpReq.TypeID,
		RarityId:        httpReq.RarityID,
		DefenseRating:   httpReq.DefenseRating,
		Durability:      httpReq.Durability,
		MagicResistance: httpReq.MagicResistance,
		ArmorSlot:       httpReq.ArmorSlot,
		Description:     httpReq.Description,
	}

	// Handle optional fields
	if httpReq.IconURL != nil {
		grpcReq.IconUrl = httpReq.IconURL
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

	// 4. Call items-service
	armorDetail, err := h.client.CreateCompleteArmor(c.Request.Context(), grpcReq)
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

	// 5. Success response
	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Complete armor created successfully. Notification sent to admins.",
		"result":     armorDetail,
	})
}

// CreateCompleteConsumableHandler creates a complete consumable (consumable + template) in one request
func (h *Handler) CreateCompleteConsumableHandler(c *gin.Context) {
	// 1. Extract userId from JWT
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	// 2. Parse JSON body
	var httpReq struct {
		// Template fields
		ItemName      string  `json:"item_name" binding:"required"`
		ItemCode      string  `json:"item_code" binding:"required"`
		IconURL       *string `json:"icon_url"`
		RequiredLevel *int32  `json:"required_level"`
		BaseSellPrice *int32  `json:"base_sell_price"`
		BaseBuyPrice  *int32  `json:"base_buy_price"`

		// Consumable fields
		TypeID        string `json:"type_id" binding:"required"`
		RarityID      string `json:"rarity_id" binding:"required"`
		HealingAmount int32  `json:"healing_amount"`
		ManaAmount    int32  `json:"mana_amount"`
		BuffDuration  int32  `json:"buff_duration"`
		MaxStackSize  int32  `json:"max_stack_size" binding:"required"`
		Description   string `json:"description"`
	}

	if err := c.ShouldBindJSON(&httpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Invalid request body",
			"error":      err.Error(),
		})
		return
	}

	// 3. Build gRPC request
	grpcReq := &pb.CreateCompleteConsumableRequest{
		UserId:        userIdStr.(string),
		ItemName:      httpReq.ItemName,
		ItemCode:      httpReq.ItemCode,
		TypeId:        httpReq.TypeID,
		RarityId:      httpReq.RarityID,
		HealingAmount: httpReq.HealingAmount,
		ManaAmount:    httpReq.ManaAmount,
		BuffDuration:  httpReq.BuffDuration,
		MaxStackSize:  httpReq.MaxStackSize,
		Description:   httpReq.Description,
	}

	// Handle optional fields
	if httpReq.IconURL != nil {
		grpcReq.IconUrl = httpReq.IconURL
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

	// 4. Call items-service
	consumableDetail, err := h.client.CreateCompleteConsumable(c.Request.Context(), grpcReq)
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

	// 5. Success response
	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Complete consumable created successfully. Notification sent to admins.",
		"result":     consumableDetail,
	})
}

// ListItemTypesHandler returns all item types for dropdown options
func (h *Handler) ListItemTypesHandler(c *gin.Context) {
	// Call items-service to get item types
	itemTypes, err := h.client.ListItemTypes(c.Request.Context())
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Failed to list item types",
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
		"message":    "Item types retrieved successfully",
		"result":     itemTypes.ItemTypes,
	})
}

// ListItemRaritiesHandler returns all item rarities for dropdown options
func (h *Handler) ListItemRaritiesHandler(c *gin.Context) {
	// Call items-service to get item rarities
	rarities, err := h.client.ListItemRarities(c.Request.Context())
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Failed to list item rarities",
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
		"message":    "Item rarities retrieved successfully",
		"result":     rarities.ItemRarities,
	})
}

