package i18n

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestI18n_Basic(t *testing.T) {
	// Read test resource file
	yamlData, err := os.ReadFile("resource.yaml")
	assert.NoError(t, err)

	err = InitAndUpdateI18nWithYaml(yamlData)
	assert.NoError(t, err)

	t.Run("Chinese Welcome", func(t *testing.T) {
		msg, err := GetLocalizeMessage("zh_CN", "welcome")
		assert.NoError(t, err)
		assert.Equal(t, "欢迎！", msg)
	})

	t.Run("English Welcome", func(t *testing.T) {
		msg, err := GetLocalizeMessage("en_US", "welcome")
		assert.NoError(t, err)
		assert.Equal(t, "Welcome!", msg)
	})

	t.Run("Japanese Goodbye", func(t *testing.T) {
		msg, err := GetLocalizeMessage("ja", "goodbye")
		assert.NoError(t, err)
		assert.Equal(t, "さようなら！", msg)
	})
}

func TestI18n_WithParams(t *testing.T) {
	// Read test resource file
	yamlData, err := os.ReadFile("resource.yaml")
	assert.NoError(t, err)

	err = InitAndUpdateI18nWithYaml(yamlData)
	assert.NoError(t, err)

	t.Run("Chinese Cats", func(t *testing.T) {
		msg, err := GetLocalizeMessage("zh_CN", "cats", map[string]interface{}{"PluralCount": 2})
		assert.NoError(t, err)
		assert.Equal(t, "我有2只猫。", msg)
	})

	t.Run("English Cats", func(t *testing.T) {
		msg, err := GetLocalizeMessage("en_US", "cats", map[string]interface{}{"PluralCount": 1})
		assert.NoError(t, err)
		assert.Equal(t, "I have 1 cat(s).", msg)
	})
}

func TestI18n_ComplexParams(t *testing.T) {
	// Read test resource file
	yamlData, err := os.ReadFile("resource.yaml")
	assert.NoError(t, err)

	err = InitAndUpdateI18nWithYaml(yamlData)
	assert.NoError(t, err)

	t.Run("Chinese Pet Owner", func(t *testing.T) {
		msg, err := GetLocalizeMessage("zh_CN", "pet_owner", map[string]interface{}{
			"Name":        "小明",
			"Age":         25,
			"PluralCount": 2,
			"PetType":     "猫",
		})
		assert.NoError(t, err)
		assert.Equal(t, "小明今年25岁，有2只猫。", msg)
	})

	t.Run("English Pet Owner", func(t *testing.T) {
		msg, err := GetLocalizeMessage("en_US", "pet_owner", map[string]interface{}{
			"Name":        "Alice",
			"Age":         30,
			"PluralCount": 1,
			"PetType":     "dog",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Alice is 30 years old and has 1 dog(s).", msg)
	})
}

func TestI18n_GreetingTime(t *testing.T) {
	// Read test resource file
	yamlData, err := os.ReadFile("resource.yaml")
	assert.NoError(t, err)

	err = InitAndUpdateI18nWithYaml(yamlData)
	assert.NoError(t, err)

	t.Run("Japanese Greeting", func(t *testing.T) {
		msg, err := GetLocalizeMessage("ja", "greeting_time", map[string]interface{}{
			"Name":    "田中",
			"Time":    "9:00",
			"Message": "おはようございます",
		})
		assert.NoError(t, err)
		assert.Equal(t, "田中さん、9:00です。おはようございます", msg)
	})

	t.Run("Korean Greeting", func(t *testing.T) {
		msg, err := GetLocalizeMessage("ko", "greeting_time", map[string]interface{}{
			"Name":    "김철수",
			"Time":    "15:00",
			"Message": "안녕하세요",
		})
		assert.NoError(t, err)
		assert.Equal(t, "김철수님, 15:00입니다. 안녕하세요", msg)
	})
}

func TestI18n_ErrorCases(t *testing.T) {
	// Read test resource file
	yamlData, err := os.ReadFile("resource.yaml")
	assert.NoError(t, err)

	err = InitAndUpdateI18nWithYaml(yamlData)
	assert.NoError(t, err)

	t.Run("Missing Parameters", func(t *testing.T) {
		msg, err := GetLocalizeMessage("zh_CN", "pet_owner", map[string]interface{}{
			"Name":        "小明",
			"PluralCount": 2,
		})
		assert.NoError(t, err)
		assert.Equal(t, "小明今年<no value>岁，有2只<no value>。", msg)
	})

	t.Run("Multiple Separate Parameters", func(t *testing.T) {
		msg, err := GetLocalizeMessage("en_US", "greeting_time",
			map[string]interface{}{"Name": "Bob"},
			map[string]interface{}{"Time": "10:00"},
			map[string]interface{}{"Message": "good morning"},
		)
		assert.NoError(t, err)
		assert.Equal(t, "Bob, it's 10:00, good morning", msg)
	})

	t.Run("Invalid Language", func(t *testing.T) {
		_, err := GetLocalizeMessage("invalid", "welcome")
		assert.Error(t, err)
	})

	t.Run("Invalid Key", func(t *testing.T) {
		_, err := GetLocalizeMessage("zh_CN", "nonexistent")
		assert.Error(t, err)
	})
}

func TestI18n_MustGet(t *testing.T) {
	// Read test resource file
	yamlData, err := os.ReadFile("resource.yaml")
	assert.NoError(t, err)

	err = InitAndUpdateI18nWithYaml(yamlData)
	assert.NoError(t, err)

	t.Run("Must Get Success", func(t *testing.T) {
		msg := MustGetLocalizeMessage("zh_CN", "welcome")
		assert.Equal(t, "欢迎！", msg)
	})

	t.Run("Must Get Panic", func(t *testing.T) {
		assert.Panics(t, func() {
			MustGetLocalizeMessage("invalid", "welcome")
		})
	})
}

func TestI18n_InvalidInit(t *testing.T) {
	t.Run("Invalid YAML", func(t *testing.T) {
		err := InitAndUpdateI18nWithYaml([]byte("invalid yaml"))
		assert.Error(t, err)
	})

	t.Run("Empty YAML", func(t *testing.T) {
		err := InitAndUpdateI18nWithYaml([]byte("{}"))
		assert.Error(t, err)
		assert.Equal(t, "no languages found", err.Error())
	})
}
