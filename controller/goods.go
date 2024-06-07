package controller

import (
	"BIT-Helper/database"
	"BIT-Helper/util/config"
	"time"

	"github.com/gin-gonic/gin"
)

type GoodsAPI struct {
	database.Goods
	User database.User `json:"user"`
	Time time.Time     `json:"time"`
}

// 获取商品
func GetGoodsAPI(goods database.Goods) GoodsAPI {
	return GoodsAPI{
		Goods: goods,
		User:  GetUserAPI(int(goods.Uid)),
		Time:  goods.CreatedAt,
	}
}

// 获取商品请求接口
func GoodsGet(c *gin.Context) {
	var goods database.Goods
	if err := database.DB.Limit(1).Find(&goods, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if goods.ID == 0 {
		c.JSON(500, gin.H{"msg": "商品不存在Orz"})
		return
	}

	c.JSON(200, GetGoodsAPI(goods))
}

// 发布商品请求接口
type GoodsPostQuery struct {
	Type  int     `json:"type" binding:"required"`
	Title string  `json:"title" binding:"required"`
	Intro string  `json:"intro" binding:"required"`
	Num   int     `json:"num" binding:"required"`
	Price float32 `json:"price" binding:"required"`
}

// 新建商品
func GoodsPost(c *gin.Context) {
	var query GoodsPostQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var goods = database.Goods{
		Type:  query.Type,
		Uid:   c.GetUint("uid_uint"),
		Title: query.Title,
		Intro: query.Intro,
		Num:   query.Num,
		Price: query.Price,
	}
	if err := database.DB.Create(&goods).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "发布成功OvO"})
}

// 修改商品请求接口
type GoodsPutQeury struct {
	Type  int     `json:"type"`
	Title string  `json:"title"`
	Intro string  `json:"intro"`
	Num   int     `json:"num"`
	Price float32 `json:"price"`
}

// 修改商品
func GoodsPut(c *gin.Context) {
	var query GoodsPutQeury
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var goods database.Goods
	if err := database.DB.Limit(1).Find(&goods, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if goods.ID == 0 {
		c.JSON(500, gin.H{"msg": "商品不存在Orz"})
		return
	}
	if goods.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有修改权限Orz"})
		return
	}

	if query.Type != 0 {
		goods.Type = query.Type
	}
	if query.Title != "" {
		goods.Title = query.Title
	}
	if query.Intro != "" {
		goods.Intro = query.Intro
	}
	if query.Num != 0 {
		goods.Num = query.Num
	}
	if query.Price != 0 {
		goods.Price = query.Price
	}
	if err := database.DB.Save(&goods).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	c.JSON(200, gin.H{"msg": "修改成功OvO"})
}

// 删除商品
func GoodsDelete(c *gin.Context) {
	var goods database.Goods
	if err := database.DB.Limit(1).Find(&goods, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if goods.ID == 0 {
		c.JSON(500, gin.H{"msg": "商品不存在Orz"})
		return
	}
	if goods.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有删除权限Orz"})
		return
	}

	if err := database.DB.Delete(&goods).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "删除成功OvO"})
}

// 获取商品列表请求接口
type GoodsListQuery struct {
	Mode    string `form:"mode"`
	Page    int    `form:"page"`
	Keyword string `form:"keyword"`
	Uid     int    `form:"uid"`
}

// 获取商品列表
func GoodsList(c *gin.Context) {
	var query GoodsListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 排序
	var order string
	if query.Mode == "time_up" {
		order = "updated_at ASC"
	} else if query.Mode == "price_up" {
		order = "price ASC"
	} else if query.Mode == "price_down" {
		order = "price DESC"
	} else {
		order = "updated_at DESC"
	}

	var goods []database.Goods
	if query.Keyword != "" {
		if query.Uid != 0 {
			database.DB.Limit(config.Config.PageSize).Where("(title LIKE ? OR intro LIKE ?) AND uid = ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%", query.Uid).Order(order).Find(&goods)
		} else {
			database.DB.Limit(config.Config.PageSize).Where("title LIKE ? OR intro LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%").Order(order).Find(&goods)
		}
	} else {
		if query.Uid != 0 {
			database.DB.Limit(config.Config.PageSize).Where("uid = ?", query.Uid).Order(order).Find(&goods)
		} else {
			database.DB.Limit(config.Config.PageSize).Order(order).Find(&goods)
		}
	}

	goodsAPI := make([]GoodsAPI, 0)
	for _, v := range goods {
		goodsAPI = append(goodsAPI, GetGoodsAPI(v))
	}

	c.JSON(200, goodsAPI)
}
