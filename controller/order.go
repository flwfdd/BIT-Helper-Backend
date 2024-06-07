/**
* @author:YHCnb
* Package:
* @date:2024/6/7 14:28
* Description:
 */
package controller

import (
	"BIT-Helper/database"
	"BIT-Helper/util/config"
	"github.com/gin-gonic/gin"
	"time"
)

type OrderAPI struct {
	database.Order
	Goods    GoodsAPI      `json:"goods"`
	Receiver database.User `json:"receiver"`
	Time     time.Time     `json:"time"`
}

// 获取订单API
func GetOrderAPI(order database.Order) OrderAPI {
	goods := database.Goods{}
	database.DB.Limit(1).Find(&goods, "id = ?", order.Goods)
	return OrderAPI{
		Order:    order,
		Goods:    GetGoodsAPI(goods),
		Receiver: GetUserAPI(int(order.Receiver)),
		Time:     order.CreatedAt,
	}
}

// 获取订单列表请求接口
type OrderListQuery struct {
	State int `form:"state"`
	Page  int `form:"page"`
}

// 获取订单列表
func OrderList(c *gin.Context) {
	var query OrderListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	if query.State <= 0 || query.State > 3 {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var orders []database.Order
	q := database.DB
	q = q.Where("state = ?", query.State)
	if err := q.Offset(query.Page * config.Config.PageSize).Limit(config.Config.PageSize).Order("updated_at DESC").Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	orderAPIs := make([]OrderAPI, 0)
	for _, order := range orders {
		orderAPIs = append(orderAPIs, GetOrderAPI(order))
	}

	c.JSON(200, orderAPIs)
}

// 获取订单
func OrderGet(c *gin.Context) {
	var order database.Order
	if err := database.DB.Limit(1).Find(&order, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if order.ID == 0 {
		c.JSON(500, gin.H{"msg": "订单不存在Orz"})
		return
	}

	c.JSON(200, GetOrderAPI(order))
}

// 发布订单请求接口
type OrderPostQuery struct {
	Goods uint `json:"goods" binding:"required"`
}

// 新建订单
func OrderPost(c *gin.Context) {
	var query OrderPostQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var goods database.Goods
	if err := database.DB.Limit(1).Find(&goods, "id = ?", query.Goods).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if goods.ID == 0 {
		c.JSON(500, gin.H{"msg": "商品不存在Orz"})
		return
	}
	if goods.Num <= 0 {
		c.JSON(500, gin.H{"msg": "商品已售完Orz"})
		return
	}
	if goods.Uid == c.GetUint("uid_uint") {
		c.JSON(500, gin.H{"msg": "不能购买自己的商品Orz"})
		return
	}

	// 创建订单
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			tx.Rollback()
		}
	}()

	order := database.Order{
		State:    1,
		Goods:    query.Goods,
		Receiver: c.GetUint("uid_uint"),
	}
	goods.Num--
	if err := tx.Create(&order).Error; err != nil {
		panic(err)
	}
	if err := tx.Save(&goods).Error; err != nil {
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"msg": "下单成功OvO"})
}

// 修改订单状态请求接口
type OrderPutQuery struct {
	State int `json:"state" binding:"required"`
}

// 修改订单状态
func OrderPut(c *gin.Context) {
	var query OrderPutQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var order database.Order
	if err := database.DB.Limit(1).Find(&order, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if order.ID == 0 {
		c.JSON(500, gin.H{"msg": "订单不存在Orz"})
		return
	}
	orderAPI := GetOrderAPI(order)
	uid := c.GetUint("uid_uint")
	if orderAPI.Goods.Uid != uid && order.Receiver != uid {
		c.JSON(500, gin.H{"msg": "无关订单，无法修改Orz"})
		return
	}
	if query.State == 2 {
		//二手：接收方确认
		//求助、活动：发布方确认
		if orderAPI.Goods.Type == 2 && order.Receiver != uid {
			c.JSON(500, gin.H{"msg": "此订单由接收方确认Orz"})
			return
		}
		if orderAPI.Goods.Type != 2 && orderAPI.Goods.Uid != uid {
			c.JSON(500, gin.H{"msg": "此订单由发布方确认Orz"})
			return
		}
		order.State = query.State
	} else if query.State == 3 {
		//双方都可撤销
		order.State = query.State
	} else {
		c.JSON(500, gin.H{"msg": "不支持的状态Orz"})
		return
	}

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	c.JSON(200, gin.H{"msg": "修改成功OvO"})
}

// 评价订单请求接口
type OrderReviewQuery struct {
	Star    int    `json:"star"`
	Content string `json:"content"`
}

// 评价订单
func OrderReview(c *gin.Context) {
	var query OrderReviewQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var order database.Order
	if err := database.DB.Limit(1).Find(&order, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if order.ID == 0 {
		c.JSON(500, gin.H{"msg": "订单不存在Orz"})
		return
	}
	orderAPI := GetOrderAPI(order)
	uid := c.GetUint("uid_uint")
	if orderAPI.Goods.Uid != uid && order.Receiver != uid {
		c.JSON(500, gin.H{"msg": "无关订单，无法评价Orz"})
		return
	}
	if order.State != 2 && order.State != 3 {
		c.JSON(500, gin.H{"msg": "订单未完成Orz"})
		return
	}

	// TODO 创建评价/修改信誉分

	c.JSON(200, gin.H{"msg": "评价成功OvO"})
}
