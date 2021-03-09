package main

import (
	"github.com/xiaobing94/dorm"
	"strings"
)

type AddProductsReq struct {
	RaisedAmount int64          `json:"raised_amount" dorm:"name:目标;shift:8;float:true"` // 募集额 募集目标
	I18N         []*ProductI18n `json:"i18n" dorm:"name:国际化;group:keys=en zh-CN ru-RU"`  // 多语言
}

type ProductI18n struct {
	//ID
	ID int64 `gorm:"primary_key;comment:'ID'" json:"id"`
	//语言代码：en zh-CN ru-RU
	Language string `gorm:"comment:'语言代码：en zh-CN ru-RU';index:product_id_language;" json:"language" validate:"oneof=en zh-CN ru-RU" dorm:"name:语言"`
	//产品ID
	ProductID int64 `gorm:"comment:'产品ID';index:product_id_language;" json:"product_id"`
	//产品名称
	Name string `gorm:"comment:'产品名称';size:255" json:"name" validate:"required" dorm:"name:名称[\\S]+;reg:true"`
	//产品描述
	Introduction string `gorm:"comment:'产品描述';size:1000;default:''" json:"introduction" dorm:"name:产品介绍[\\S]+;reg:true"`
}

func (p *ProductI18n) UnmarshalDocument(tagName string, data map[string]interface{}, metaInfo interface{}, opt ...interface{}) error {
	languages := []string{"en", "zh-CN", "ru-RU"}
	for _, language := range languages {
		isBreak := false
		for k, _ := range data {
			if strings.Contains(k, language) {
				p.Language = language
				isBreak = true
				break
			}
		}
		if isBreak {
			break
		}
	}
	err := dorm.Unmarshal(p, data, metaInfo)
	return err
}


func main() {
	mapper, err := dorm.OpenXlsFile("/Users/yanjianguo/product.xlsx")
	if err != nil {
		println(err.Error())
		return
	}

	var p []*AddProductsReq
	err = mapper.GetObjectsFromParser(&p)
	if err != nil {
		println(err.Error())
		return
	}
}
