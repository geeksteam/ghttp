package bruteforce

type BruteForce struct {
	BlockAttempts int    `default:"10" comment:"How much attempts before ban"`
	BanTime       int64  `default:"600" comment:"How much seconds will be banned after failed attempts"`
	DataEncoding  string `default:"mspack" comment:"Encoding of values for boltdb storage. Values:[mspack, json]"`
}
