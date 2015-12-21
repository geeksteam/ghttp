package bruteforce

// BruteIP is a info struct, which holds information about given IP's activity history.
type BruteIP struct {
	Timestamp int64 // Unixtime последней попытки
	Attempts  int   // Количество попыток
}
