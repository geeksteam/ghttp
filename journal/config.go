package journal

type Journal struct {
	BoltDB              string `default:"./db/journal.db" comment:"Path to bolt db"`
	BucketForOperations string `default:"Operations" comment:"name of bucket which holds operations"`
	Capacity            int    `default:"60" comment:"How much days store in journal"`
	DataEncoding        string `default:"mspack" comment:"Encoding of values for boltdb storage. Values:[mspack, json]"`
}
