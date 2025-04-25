package i18n

import (
	"errors"
	"fmt"

	"github.com/dgdts/ts-gobase/atomic_buffer"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	buffer = atomic_buffer.NewAtomicBuffer(&ki18n{})
)

type ki18n struct {
	localizerMap map[string]*i18n.Localizer
	bundle       *i18n.Bundle
}

func (i *ki18n) createLocalizer(lang string, multiLangMap map[string]map[string]string) (*i18n.Localizer, error) {
	langMessage := make(map[string]map[string]string)
	for key, value := range multiLangMap {
		langMessage[key] = map[string]string{
			"other": value[lang],
		}
	}

	yamlData, err := yaml.Marshal(langMessage)
	if err != nil {
		return nil, err
	}
	i.bundle.MustParseMessageFileBytes(yamlData, fmt.Sprintf("%s.yaml", lang))

	localizer := i18n.NewLocalizer(i.bundle, lang)

	return localizer, nil
}

func InitAndUpdateI18nWithYaml(rawYamlData []byte) error {
	multiLangMap := make(map[string]map[string]string)
	err := yaml.Unmarshal(rawYamlData, &multiLangMap)
	if err != nil {
		return err
	}

	return InitAndUpdateI18n(multiLangMap)
}

func InitAndUpdateI18n(multiLangMap map[string]map[string]string) error {
	initI18n := &ki18n{}

	initI18n.localizerMap = make(map[string]*i18n.Localizer)
	initI18n.bundle = i18n.NewBundle(language.English)
	initI18n.bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	languages := make(map[string]struct{})
	if len(multiLangMap) == 0 {
		return errors.New("no languages found")
	}

	for _, langs := range multiLangMap {
		for lang := range langs {
			languages[lang] = struct{}{}
		}
	}

	for lang := range languages {
		localizer, err := initI18n.createLocalizer(lang, multiLangMap)
		if err != nil {
			return err
		}
		initI18n.localizerMap[lang] = localizer
	}

	buffer.Store(initI18n)

	return nil
}

func GetLocalizeMessage(lang string, key string, params ...any) (string, error) {
	i := buffer.Load()
	localizerMap := i.localizerMap

	if localizer, ok := localizerMap[lang]; ok {
		templateData := make(map[string]any)
		for _, param := range params {
			if paramMap, ok := param.(map[string]any); ok {
				for k, v := range paramMap {
					templateData[k] = v
				}
			}
		}

		return localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: templateData,
		})
	}
	return "", errors.New("lang not found")
}

func MustGetLocalizeMessage(lang string, key string, params ...any) string {
	message, err := GetLocalizeMessage(lang, key, params...)
	if err != nil {
		panic(err)
	}
	return message
}
