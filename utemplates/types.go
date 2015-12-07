package utemplates

// UserTemplate Структура шаблона прав юзера для базы
type UserTemplate struct {
	Modules []string   // Список с названием доступных пользователю модулей панели (dns, mysql, www...)
	SSH     bool       // Есть ли доступ в ssh у пользователя (установлен ли bash или nologin, если доступа нет).
	Limits  UserLimits // Ограничения пользователя. Структура уже описана в bolt/users.go
}

// UserLimits Ограничения пользователя
type UserLimits struct {
	Diskspace  int64 // Размер диска в Мбайтах
	Traffic    int64 // Размер разрешенного трафика в ГБ
	DBs        int   // Количество баз данных
	DBUsers    int   // Количество пользователей баз данных
	FTPs       int   // Количество фтп аккаунтов
	WebDomains int   // Количество доменов
	DNSDomains int   // Количество доменов в DNS
	Emails     int   // Количество почтовых ящиков
}
