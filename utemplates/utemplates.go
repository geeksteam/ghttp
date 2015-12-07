package utemplates

import (
	"log"

	"github.com/geekbros/Tools/boltdb"
)

var (
	cfg                       Utemplates
	BoltDBMain                string
	BotlDBUserTemplatesBacket string
)

func SetConfig(c Utemplates) {
	cfg = c
	BoltDBMain = cfg.BoltDBMain
	BotlDBUserTemplatesBacket = cfg.BotlDBUserTemplatesBacket
}

// Set Creates new or update entry for template
func Set(templateName string, templ UserTemplate) {
	boltdb.DB(BoltDBMain, cfg.DataEncoding).Bucket(BotlDBUserTemplatesBacket).Set(templateName, templ)
}

//Delete Deletes template from db
func Delete(templateName string) {
	boltdb.DB(BoltDBMain, cfg.DataEncoding).Bucket(BotlDBUserTemplatesBacket).Delete(templateName)
}

// Get Returns template with gien name
func Get(templateName string) *UserTemplate {
	templ := &UserTemplate{}
	err := boltdb.DB(BoltDBMain, cfg.DataEncoding).Bucket(BotlDBUserTemplatesBacket).Get(templateName, templ)
	if err != nil {
		log.Println("Can't load user tamplate:", err)
		return nil
	}
	return templ
}

// GetAll Get all tamplates with their names
func GetAll() (map[string]UserTemplate, error) {
	templs := map[string]UserTemplate{}

	ut, err := boltdb.DB(BoltDBMain, cfg.DataEncoding).Bucket(BotlDBUserTemplatesBacket).GetAll(&UserTemplate{})
	if err != nil {
		return templs, err
	}

	for k, v := range ut {
		templs[k] = *v.(*UserTemplate)
	}
	return templs, nil
}
