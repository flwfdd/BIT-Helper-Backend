/**
* @author:Ruiqin-Huang
* Package:
* @date:2024/11/11 14:30
* Description:
 */

package controller

import (
	"BIT-Helper/database"
	"BIT-Helper/util/config"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 所有话题的API格式（返回给前端的数据格式）
type TopicAPI struct {
	database.Topic
	User          UserAPI               `json:"user"`
	Images        []ImageAPI            `json:"images"`
	Tags          []string              `json:"tags"`
	Time          time.Time             `json:"time"`
	Own           bool                  `json:"own"`          // 标识是否为自己发布的话题
	Like          bool                  `json:"like"`         // 标识是否已经为该条话题点赞
	Vote          bool                  `json:"vote"`         // 标识是否已经为该话题投票
	VotedOptionID uint                  `json:"voted_option"` // 标识用户已经对哪个选项进行投票
	Options       []database.VoteOption `json:"options"`      // 投票选项及票数
}

// 用于分割字符串（处理空元素的情况）
func split(str string) []string {
	l := strings.Split(str, " ")
	out := make([]string, 0)
	for i := range l {
		if l[i] != "" {
			out = append(out, l[i])
		}
	}
	return out
}

// 获取话题请求接口
func GetTopicAPI(topic database.Topic, c *gin.Context) TopicAPI {
	var tags []database.TopicTag
	database.DB.Where("topic_id = ?", topic.ID).Find(&tags)
	// tagStrings是一个字符串数组，存储了所有的标签
	tagStrings := make([]string, len(tags))
	for i, tag := range tags {
		tagStrings[i] = tag.Tag
	}
	// 检查用户是否已点赞该话题
	Obj := "topic" + strconv.Itoa(int(topic.ID))

	// 获取投票选项及票数
	var voteOptions []database.VoteOption
	if topic.IsVote {
		database.DB.Where("topic_id = ?", topic.ID).Find(&voteOptions)
	}

	// 检查用户是否已投票
	var voteRecord database.VoteRecord
	voted := false
	votedOptionID := uint(0)
	if err := database.DB.Where("topic_id = ? AND user_id = ?", topic.ID, c.GetUint("uid_uint")).First(&voteRecord).Error; err == nil {
		voted = true
		votedOptionID = voteRecord.VoteOptionID
	}

	return TopicAPI{
		Topic:         topic,
		User:          GetUserAPI(int(topic.Uid)),
		Images:        GetImageAPIArr(split(topic.Image)),
		Tags:          tagStrings,
		Time:          topic.CreatedAt,
		Own:           (c.GetUint("uid_uint") == topic.Uid),
		Like:          CheckLike(Obj, c.GetUint("uid_uint")),
		Vote:          voted,
		VotedOptionID: votedOptionID,
		Options:       voteOptions,
	}
}

// 获取单个话题
func TopicGet(c *gin.Context) {
	var topic database.Topic
	// 获取话题类型和ID
	topicType := c.Param("type")
	id := c.Param("id")
	// TODO: 请看这里->gorm.ErrRecordNotFound错误处理问题
	// gorm.ErrRecordNotFound是 First、Last 和 Take 方法特有的错误返回
	// Find 方法不会返回 ErrRecordNotFound 错误，如果没有找到记录，它会返回一个空切片
	// 参考https://gorm.io/zh_CN/docs/query.html
	// 查询单条记录推荐使用First方法
	if err := database.DB.Where("type = ? AND id = ?", topicType, id).First(&topic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 若查询结果为空
			c.JSON(404, gin.H{"msg": "话题不存在Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		}
		return
	}
	// 查询成功，返回该条话题信息
	c.JSON(200, GetTopicAPI(topic, c))
}

// 使用多条件查询
// 前端推荐设计成如下形式：
// ------------------------------------------------------
// 1. 请输入话题关键字（标题中或内容中）：_____（可为空）
// 2. 请输入用户ID：_____（可为空）
// 3. 选择话题标签：标签1 | 标签2 | 标签3 | ...（预先展示部分已有标签（该type话题中用户定义的所有标签），点击以选择（一次只能选择一个标签），可不选）
// 4. 选择排序方式：点赞数升序 | 点赞数降序 | 评分升序 | 评分降序 | 评论数升序 | 评论数降序 | 时间升序 | 时间降序 （下拉框，点击以选择，可不选）
// 5. 显示查询按钮，点击后显示查询结果
// ------------------------------------------------------
type TopicListQuery struct {
	Mode    string `form:"mode"`
	Page    int    `form:"page"`
	Keyword string `form:"keyword"`
	Uid     int    `form:"uid"`
	Tag     string `form:"tag"` // 支持依据用户自定义tag查找话题
}

type TopicListResponse struct {
	Topics []TopicAPI `json:"topics"`
	Tags   []string   `json:"tags"`
}

// 获取话题列表（支持多条件查询）
func TopicList(c *gin.Context) {
	// 定义查询结构体
	var query TopicListQuery
	// 从context中获取查询类型及查询结构体
	topicType := c.Param("type")
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 排序方式
	var order string
	if query.Mode == "like_up" { // 按点赞数升序排序
		order = "LikeNum ASC"
	} else if query.Mode == "like_down" { // 按点赞数降序排序
		order = "LikeNum DESC"
	} else if query.Mode == "rate_up" { // 按评分升序排序
		order = "AvgRate ASC"
	} else if query.Mode == "rate_down" { // 按评分降序排序
		order = "AvgRate DESC"
	} else if query.Mode == "comment_up" { // 按评论数升序排序
		order = "CommentNum ASC"
	} else if query.Mode == "comment_down" { // 按评论数降序排序
		order = "CommentNum DESC"
	} else if query.Mode == "time_up" { // 按时间升序排序
		order = "updated_at ASC"
	} else { //	【默认】按时间降序排序(最新话题的在前)
		order = "updated_at DESC"
	}

	// topics存储数据库查询结果（topic对象列表）
	// 加上"%"表示模糊查询，表示匹配包含 query.Keyword 的任意字符串
	var topics []database.Topic

	if query.Keyword != "" {
		if query.Uid != 0 {
			if query.Tag != "" {
				// 根据标签筛选 TopicTag 表
				var topicTags []database.TopicTag
				database.DB.Where("tag = ?", query.Tag).Find(&topicTags)

				// 获取符合标签的 topic_id 列表
				var topicIDs []uint
				for _, topicTag := range topicTags {
					topicIDs = append(topicIDs, topicTag.TopicID)
				}

				// 根据其他查询条件查询 Topic 表
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(title LIKE ? OR content LIKE ?) AND (uid = ?) AND (type = ?) AND id IN (?)", "%"+query.Keyword+"%", "%"+query.Keyword+"%", query.Uid, topicType, topicIDs).
					Order(order).Find(&topics)
			} else {
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(title LIKE ? OR content LIKE ?) AND (uid = ?) AND (type = ?)", "%"+query.Keyword+"%", "%"+query.Keyword+"%", query.Uid, topicType).
					Order(order).Find(&topics)
			}
		} else {
			if query.Tag != "" {
				// 根据标签筛选 TopicTag 表
				var topicTags []database.TopicTag
				database.DB.Where("tag = ?", query.Tag).Find(&topicTags)

				// 获取符合标签的 topic_id 列表
				var topicIDs []uint
				for _, topicTag := range topicTags {
					topicIDs = append(topicIDs, topicTag.TopicID)
				}

				// 根据其他查询条件查询 Topic 表
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(title LIKE ? OR content LIKE ?) AND (type = ?) AND id IN (?)", "%"+query.Keyword+"%", "%"+query.Keyword+"%", topicType, topicIDs).
					Order(order).Find(&topics)
			} else {
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(title LIKE ? OR content LIKE ?) AND (type = ?)", "%"+query.Keyword+"%", "%"+query.Keyword+"%", topicType).
					Order(order).Find(&topics)
			}
		}
	} else {
		if query.Uid != 0 {
			if query.Tag != "" {
				// 根据标签筛选 TopicTag 表
				var topicTags []database.TopicTag
				database.DB.Where("tag = ?", query.Tag).Find(&topicTags)

				// 获取符合标签的 topic_id 列表
				var topicIDs []uint
				for _, topicTag := range topicTags {
					topicIDs = append(topicIDs, topicTag.TopicID)
				}

				// 根据其他查询条件查询 Topic 表
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(uid = ?) AND (type = ?) AND id IN (?)", query.Uid, topicType, topicIDs).
					Order(order).Find(&topics)
			} else {
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(uid = ?) AND (type = ?)", query.Uid, topicType).
					Order(order).Find(&topics)
			}
		} else {
			if query.Tag != "" {
				// 根据标签筛选 TopicTag 表
				var topicTags []database.TopicTag
				database.DB.Where("tag = ?", query.Tag).Find(&topicTags)

				// 获取符合标签的 topic_id 列表
				var topicIDs []uint
				for _, topicTag := range topicTags {
					topicIDs = append(topicIDs, topicTag.TopicID)
				}

				// 根据其他查询条件查询 Topic 表
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(type = ?) AND id IN (?)", topicType, topicIDs).
					Order(order).Find(&topics)
			} else {
				database.DB.Offset(config.Config.PageSize*query.Page).Limit(config.Config.PageSize).
					Where("(type = ?)", topicType).
					Order(order).Find(&topics)
			}
		}
	}

	// 返回查询结果
	c.JSON(200, topics)

	// 若查询结果为空，.find()返回空列表，不处理

	// 将数据库查询结果转换为API格式
	topicAPIList := make([]TopicAPI, 0)
	for _, v := range topics {
		topicAPIList = append(topicAPIList, GetTopicAPI(v, c))
	}

	// 获取该类型的所有标签
	var allTags []database.TopicTag
	database.DB.Where("type = ?", topicType).Find(&allTags)
	tagStrings := make([]string, len(allTags))
	for i, tag := range allTags {
		tagStrings[i] = tag.Tag
	}

	// 返回查询结果和所有标签
	response := TopicListResponse{
		Topics: topicAPIList,
		Tags:   tagStrings,
	}
	c.JSON(200, response)
}

// 发布话题请求接口
type TopicPostQuery struct {
	Type      int      `json:"type" binding:"required"`
	Title     string   `json:"title" binding:"required"`
	Content   string   `json:"content" binding:"required"`
	Tags      []string `json:"tags"`
	ImageMids []string `json:"image_mids"`
	IsVote    bool     `json:"is_vote"` // 是否为投票话题
	Options   []string `json:"options"` // 投票选项
}

// 发布一条话题
func TopicPost(c *gin.Context) {
	var query TopicPostQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var topic = database.Topic{
		Type:    query.Type,            // 指定话题类型（不从url中指定，而是从请求体中指定）
		Uid:     c.GetUint("uid_uint"), // 从上下文中获取 uid_uint 的值。这个值通常是在请求处理过程中【通过中间件】或其他方式设置到上下文中的，而【不是直接从请求体】中获取。
		Title:   query.Title,
		Content: query.Content,
		Image:   strings.Join(query.ImageMids, " "), // 在每张图片的mid之间加上空格，向数据库传递的是一个字符串
		IsVote:  query.IsVote,                       // 设置是否为投票话题
	}

	if err := database.DB.Create(&topic).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	for _, tag := range query.Tags {
		database.DB.Create(&database.TopicTag{
			TopicID: topic.ID, // 话题ID
			Tag:     tag,      // 话题标签(string)
		})
	}

	// 如果是投票话题，添加投票选项
	if query.IsVote {
		for _, option := range query.Options {
			database.DB.Create(&database.VoteOption{
				TopicID: topic.ID, // 所属主题ID
				Option:  option,   // 投票选项内容
			})
		}
	}

	c.JSON(200, GetTopicAPI(topic, c))
}

// 修改话题请求接口
type TopicPutQuery struct {
	Type      int      `json:"type"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags"`
	ImageMids []string `json:"image_mids"`
	IsVote    bool     `json:"is_vote"` // 是否为投票话题
	Options   []string `json:"options"` // 投票选项
}

// 修改话题
func TopicPut(c *gin.Context) {
	var query TopicPutQuery
	// 从url中获取话题id
	id := c.Param("id")
	// 从请求体中获取修改后的话题信息
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var topic database.Topic
	if err := database.DB.First(&topic, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"msg": "话题不存在Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		}
		return
	}

	if topic.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(401, gin.H{"msg": "没有修改权限Orz"})
		return
	}
	if !CheckImage(query.ImageMids) {
		c.JSON(400, gin.H{"msg": "存在未上传成功的图片Orz"})
		return
	}

	// 话题编号从1开始，1-校园生活，2-电影，3-音乐，4-读书
	// 若请求体中query.Type为0，则不修改话题类型
	if query.Type != 0 {
		topic.Type = query.Type
	}
	// 若请求体中query.Title为""，则不修改话题标题
	if query.Title != "" {
		topic.Title = query.Title
	}
	// 若请求体中query.Content为""，则不修改话题内容
	if query.Content != "" {
		topic.Content = query.Content
	}
	// 每次修改【必须】更新图片信息
	topic.Image = strings.Join(query.ImageMids, " ")
	// 更新是否为投票话题
	topic.IsVote = query.IsVote

	// 保存修改后的话题信息到数据库
	if err := database.DB.Save(&topic).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	// 每次修改【必须】更新话题标签信息
	// 先删除原有的标签信息
	database.DB.Where("topic_id = ?", topic.ID).Delete(&database.TopicTag{})
	// 再添加新的标签信息
	for _, tag := range query.Tags {
		database.DB.Create(&database.TopicTag{
			TopicID: topic.ID,
			Tag:     tag,
		})
	}

	// 如果是投票话题，处理投票选项
	if query.IsVote {
		// 获取现有的投票选项
		var existingOptions []database.VoteOption
		database.DB.Where("topic_id = ?", topic.ID).Find(&existingOptions)

		// 创建一个映射来跟踪现有选项
		existingOptionsMap := make(map[string]database.VoteOption)
		for _, option := range existingOptions {
			existingOptionsMap[option.Option] = option
		}

		// 处理新的投票选项
		for _, option := range query.Options {
			if _, exists := existingOptionsMap[option]; !exists {
				// 新增选项
				database.DB.Create(&database.VoteOption{
					TopicID: topic.ID,
					Option:  option,
				})
			} else {
				// 从映射中删除已存在的选项
				delete(existingOptionsMap, option)
			}
		}

		// 处理删除的选项
		for _, option := range existingOptionsMap {
			// 软删除选项
			database.DB.Delete(&option)
			// 删除相关的投票记录
			database.DB.Where("vote_option_id = ?", option.ID).Delete(&database.VoteRecord{})
		}
	}

	c.JSON(200, GetTopicAPI(topic, c))
}

// 删除话题
func TopicDelete(c *gin.Context) {
	var topic database.Topic
	id := c.Param("id")
	if err := database.DB.First(&topic, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"msg": "话题不存在Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		}
		return
	}

	if topic.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(401, gin.H{"msg": "没有删除权限Orz"})
		return
	}

	if err := database.DB.Delete(&topic).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	// 删除话题时，同时删除话题标签信息
	database.DB.Where("topic_id = ?", topic.ID).Delete(&database.TopicTag{})

	// 如果是投票话题，删除相关的投票选项和投票记录
	if topic.IsVote {
		var voteOptions []database.VoteOption
		database.DB.Where("topic_id = ?", topic.ID).Find(&voteOptions)
		for _, option := range voteOptions {
			// 删除投票选项
			database.DB.Delete(&option)
			// 删除相关的投票记录
			database.DB.Where("vote_option_id = ?", option.ID).Delete(&database.VoteRecord{})
		}
	}

	c.JSON(200, gin.H{"msg": "删除成功OvO"})
}

// 点赞话题（notice: 已转移到reaction.go中实现）
// func TopicLike(c *gin.Context) {
// 	var topic database.Topic
// 	id := c.Param("id")
// 	if err := database.DB.First(&topic, "id = ?", id).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			c.JSON(404, gin.H{"msg": "话题不存在Orz"})
// 		} else {
// 			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
// 		}
// 		return
// 	}

// 	topic.LikeNum++
// 	if err := database.DB.Save(&topic).Error; err != nil {
// 		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
// 		return
// 	}

// 	c.JSON(200, gin.H{"msg": "点赞成功OvO"})
// }

// 投票请求接口
type VoteTopicQuery struct {
	TopicID      uint `json:"topic_id" binding:"required"`       // 话题ID
	VoteOptionID uint `json:"vote_option_id" binding:"required"` // 投票选项ID
}

// 投票
func VoteTopic(c *gin.Context) {
	var query VoteTopicQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("uid_uint")

	// 检查话题是否存在
	var topic database.Topic
	if err := database.DB.First(&topic, "id = ?", query.TopicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"msg": "话题不存在Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		}
		return
	}

	// 检查是否为投票话题
	if !topic.IsVote {
		c.JSON(400, gin.H{"msg": "该话题不是投票话题Orz"})
		return
	}

	// 检查投票选项是否存在
	var voteOption database.VoteOption
	if err := database.DB.First(&voteOption, "id = ? AND topic_id = ?", query.VoteOptionID, query.TopicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"msg": "投票选项不存在Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		}
		return
	}

	// 检查用户是否已经投过票
	var voteRecord database.VoteRecord
	if err := database.DB.First(&voteRecord, "topic_id = ? AND user_id = ?", query.TopicID, userID).Error; err == nil {
		c.JSON(400, gin.H{"msg": "你已经投过票了Orz"})
		return
	}

	// 创建投票记录
	voteRecord = database.VoteRecord{
		TopicID:      query.TopicID,
		VoteOptionID: query.VoteOptionID,
		UserID:       userID,
	}
	if err := database.DB.Create(&voteRecord).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	// 更新投票选项的投票数
	voteOption.VoteNum++
	if err := database.DB.Save(&voteOption).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	c.JSON(200, gin.H{"msg": "投票成功OvO"})
}
