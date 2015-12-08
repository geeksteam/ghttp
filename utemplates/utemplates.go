package utemplates

import (
	"log"

	"github.com/geekbros/Tools/boltdb"
)

var (
	cfg Utemplates
)

func SetConfig(c Utemplates) {
	cfg = c
}

// Set Creates new or update entry for template
func Set(templateName string, templ UserTemplate) {
	boltdb.DB(cfg.BoltDBMain, cfg.DataEncoding).Bucket(cfg.BotlDBUserTemplatesBacket).Set(templateName, templ)
}

//Delete Deletes template from db
func Delete(templateName string) {
	boltdb.DB(cfg.BoltDBMain, cfg.DataEncoding).Bucket(cfg.BotlDBUserTemplatesBacket).Delete(templateName)
}

// Get Returns template with gien name
func Get(templateName string) *UserTemplate {
	log.Printf("GETTING TEMPLATE %v\n", templateName)
	templ := &UserTemplate{}
	log.Printf("db main: %v, encoding: %v\n", cfg.BoltDBMain, cfg.DataEncoding)
	err := boltdb.DB(cfg.BoltDBMain, cfg.DataEncoding).Bucket(cfg.BotlDBUserTemplatesBacket).Get(templateName, templ)
	if err != nil {
		log.Println("Can't load user tamplate:", err)
		return nil
	}
	return templ
}

// GetAll Get all tamplates with their names
func GetAll() (map[string]UserTemplate, error) {
	templs := map[string]UserTemplate{}

	ut, err := boltdb.DB(cfg.BoltDBMain, cfg.DataEncoding).Bucket(cfg.BotlDBUserTemplatesBacket).GetAll(&UserTemplate{})
	if err != nil {
		return templs, err
	}

	for k, v := range ut {
		templs[k] = *v.(*UserTemplate)
	}
	return templs, nil
}
