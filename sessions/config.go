package sessions

type SessionsConf struct {
	SessionIDKey       string `default:"sessionID" comment:"Key of session id in cookies map, which generates randomly."`
	SessionIDKeyLength int    `default:"24" comment:"Length of session id key for random generation."`
	SessionLifeTime    int    `default:"1800" comment:"Lifetime of a session. Seconds."`
	StrictIP           bool   `default:"true" comment:"Compare client IP with IP in session."`
}
