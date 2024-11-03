package controller

import (
	"BIT-Helper/database"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatlistAPI struct {
	User database.User `json:"user"`
}

type ChatAPI struct {
	Id       int       `json:"id"`
	Time     time.Time `json:"time"`
	Sender   UserAPI   `json:"sender"`
	Receiver UserAPI   `json:"receiver"`
	Content  string    `json:"content"`
}

// 获取消息列表
func GetChatlistAPI(id int) UserAPI {
	return GetUserAPI(id)
}
func GetChatsAPI(now_chat database.Chats) ChatAPI {
	return ChatAPI{
		Id:       int(now_chat.ID),
		Time:     now_chat.Time,
		Sender:   GetUserAPI(int(now_chat.Uid_from)),
		Receiver: GetUserAPI(int(now_chat.Uid_to)),
		Content:  now_chat.Content,
	}
}

// 获取消息列表接口
func GetAllChats(c *gin.Context) {
	var chats_from []int
	var chats_to []int

	// 从数据库获取所有记录并按时间字段升序排序
	if err := database.DB.Model(&database.Chats{}).Where("uid_from = ?", c.GetUint("uid_uint")).Distinct("uid_to").Pluck("uid_to", &chats_to).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误", "error": err.Error()})
		return
	}
	if err := database.DB.Model(&database.Chats{}).Where("uid_to = ?", c.GetUint("uid_uint")).Distinct("uid_from").Pluck("uid_from", &chats_from).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误", "error": err.Error()})
		return
	}
	//去重
	uniqueValues := make(map[int]bool)
	for _, value := range chats_from {
		uniqueValues[value] = true
	}
	for _, value := range chats_to {
		uniqueValues[value] = true
	}

	// 将map的键转换为切片
	chats := make([]int, 0, len(uniqueValues))
	for value := range uniqueValues {
		chats = append(chats, value)
	}
	// // 检查是否找到任何记录
	// if len(chats) == 0 {
	// 	c.JSON(404, gin.H{"msg": "没有找到记录"})
	// 	// 返回空列表？or ？
	// 	return
	// }

	// 成功找到记录，返回JSON响应
	chatsAPI := make([]UserAPI, 0)
	for _, v := range chats {
		chatsAPI = append(chatsAPI, GetChatlistAPI(v))
	}

	c.JSON(200, chatsAPI)
}

// 发送一个对话
func ChatsPost(c *gin.Context) {
	var query struct {
		Type    string `json:"type" binding:"required"`
		Content string `form:"content" binding:"required"` // 消息内容，必填
	}
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa", "error": err.Error()})
		return
	}
	i, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa", "error": err.Error()})
		return
	} else {
		//fmt.Println("转换结果:", i)
	}
	// 创建聊天记录的模型实例
	var chat = database.Chats{
		Type:     query.Type,
		Uid_from: c.GetUint("uid_uint"), // 发送者ID从上下文中获取
		Uid_to:   uint(i),               // 接收者ID
		Content:  query.Content,         // 消息内容
	}

	// 在数据库中创建新的聊天记录
	if err := database.DB.Create(&chat).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz", "error": err.Error()})
		return
	}

	// 返回成功消息
	c.JSON(200, gin.H{"msg": "消息发送成功OvO"})
}

// 通过id获取消息
func GetChatsById(c *gin.Context) {
	var chats []database.Chats
	i, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa", "error": err.Error()})
		return
	}

	// 从数据库获取所有记录并按时间字段升序排序
	if err := database.DB.Order("created_at asc").Find(&chats, "(uid_from = ? and uid_to = ?) or (uid_to = ? and uid_from = ?)", c.GetUint("uid_uint"), i, c.GetUint("uid_uint"), i).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误", "error": err.Error()})
		return
	}
	// 检查是否找到任何记录
	if len(chats) == 0 {
		c.JSON(404, gin.H{"msg": "没有找到记录"})
		return
	}
	chatsAPI := make([]ChatAPI, 0)
	for _, v := range chats {
		chatsAPI = append(chatsAPI, GetChatsAPI(v))
	}

	c.JSON(200, chatsAPI)
}
