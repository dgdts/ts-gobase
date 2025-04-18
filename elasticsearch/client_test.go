package elasticsearch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestDoc struct {
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Tags     []string  `json:"tags"`
	CreateAt time.Time `json:"create_at"`
}

func TestElasticsearch(t *testing.T) {
	// 初始化客户端
	cfg := Config{
		Addresses: []string{"https://localhost:9200"}, // 修改为 https
		Username:  "elastic",
		Password:  "wVpOH1cQtCBUVAi1PE7r",
	}

	err := InitEsClient(cfg)
	assert.NoError(t, err)

	ctx := context.Background()
	index := "test_articles"
	docID := "test_001"

	// 准备测试文档
	doc := TestDoc{
		Title:    "测试文章",
		Content:  "这是一篇测试文章的内容",
		Tags:     []string{"测试", "示例"},
		CreateAt: time.Now(),
	}

	// 测试索引文档
	t.Run("Index Document", func(t *testing.T) {
		err := Index(ctx, index, docID, doc)
		assert.NoError(t, err)
	})

	// 测试获取文档
	t.Run("Get Document", func(t *testing.T) {
		var result TestDoc
		err := Get(ctx, index, docID, &result)
		assert.NoError(t, err)
		assert.Equal(t, doc.Title, result.Title)
		assert.Equal(t, doc.Content, result.Content)
		assert.ElementsMatch(t, doc.Tags, result.Tags)
	})

	// 测试搜索文档
	t.Run("Search Document", func(t *testing.T) {
		query := map[string]any{
			"query": map[string]any{
				"match": map[string]any{
					"title": "测试",
				},
			},
		}

		var result map[string]any
		err := Search(ctx, index, query, &result)
		assert.NoError(t, err)
		assert.NotNil(t, result["hits"])
	})

	// 测试删除文档
	t.Run("Delete Document", func(t *testing.T) {
		err := Delete(ctx, index, docID)
		assert.NoError(t, err)

		// 验证文档已被删除
		var result TestDoc
		err = Get(ctx, index, docID, &result)
		assert.Error(t, err)
	})
}
