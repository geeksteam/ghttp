package utemplates

type Utemplates struct {
	BotlDBUserTemplatesBacket string `default:"UserTemplates" comment:"Name of usertamplates bucket in main db"`
	BoltDBMain                string `default:"./db/main.db" comment:"Path to main db"`
	DataEncoding              string `default:"mspack" comment:"Encoding of values for boltdb storage. Values:[mspack, json]"`
}
