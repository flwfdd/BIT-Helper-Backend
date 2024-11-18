/*
 * @Author: flwfdd
 * @Date: 2024-11-17 23:39:12
 * @LastEditTime: 2024-11-18 00:39:16
 * @Description:
 * _(:з」∠)_
 */
package ai

import (
	"BIT-Helper/util/config"
	"BIT-Helper/util/request"
	"encoding/json"
)

func QueryAI(prompt string) (string, error) {
	query_json := map[string]interface{}{
		"model": "glm-4-flash",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	query, err := json.Marshal(query_json)
	if err != nil {
		return "", err
	}
	res, err := request.PostJSON("https://open.bigmodel.cn/api/paas/v4/chat/completions", string(query), map[string]string{
		"Authorization": "Bearer " + config.Config.ZhipuKey,
	})
	if err != nil {
		return "", err
	}
	type ResponseBody struct {
		Choices []struct {
			Index        int    `json:"index"`
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	var res_body ResponseBody
	err = json.Unmarshal(res.Content, &res_body)
	if err != nil {
		return "", err
	}
	return res_body.Choices[0].Message.Content, nil
}
