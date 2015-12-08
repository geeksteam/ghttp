package utemplates

import (
	"log"

	"github.com/geekbros/Tools/boltdb"
)

type Utemplates struct {
	BotlDBUserTemplatesBacket string `default:"UserTemplates" comment:"Name of usertamplates bucket in main db"`
	BoltDBMain                string `default:"./db/main.db" comment:"Path to main db"`
	DataEncoding              string `default:"mspack" comment:"Encoding of values for boltdb storage. Values:[mspack, json]"`
}

// Set Creates new or update entry for template
func (u Utemplates) Set(templateName string, templ UserTemplate) {
	boltdb.DB(u.BoltDBMain, u.DataEncoding).Bucket(u.BotlDBUserTemplatesBacket).Set(templateName, templ)
}

//Delete Deletes template from db
func (u Utemplates) Delete(templateName string) {
	boltdb.DB(u.BoltDBMain, u.DataEncoding).Bucket(u.BotlDBUserTemplatesBacket).Delete(templateName)
}

// Get Returns template with gien name
func (u Utemplates) Get(templateName string) *UserTemplate {
	log.Printf("GETTING TEMPLATE %v\n", templateName)
	templ := &UserTemplate{}
	log.Printf("db main: %v, encoding: %v\n", u.BoltDBMain, u.DataEncoding)
	err := boltdb.DB(u.BoltDBMain, u.DataEncoding).Bucket(u.BotlDBUserTemplatesBacket).Get(templateName, templ)
	if err != nil {
		log.Println("Can't load user tamplate:", err)
		return nil
	}
	return templ
}

// GetAll Get all tamplates with their names
func (u Utemplates) GetAll() (map[string]UserTemplate, error) {
	templs := map[string]UserTemplate{}

	ut, err := boltdb.DB(u.BoltDBMain, u.DataEncoding).Bucket(u.BotlDBUserTemplatesBacket).GetAll(&UserTemplate{})
	if err != nil {
		return templs, err
	}

	for k, v := range ut {
		templs[k] = *v.(*UserTemplate)
	}
	return templs, nil
}
