package common

var (
	GameConf = &GameServerConf{}
)

type GameServerConf struct {
	AccountHost string
	AccountPort int

	HallHost   string
	HallPort   int
	HallSecret string

	GameHost string
	GamePort int
	LogPath  string
	LogLevel string
}
