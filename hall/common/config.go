package common

var (
	HallConf = &HallServiceConf{}
)

type HallServiceConf struct {
	AccountHost string
	AccountPort int
	HallHost    string
	HallPort    int
	HallSecret  string
	LogPath     string
	LogLevel    string
	Version     string

	AppId       int		//qq登录
	AppKey      string
	RedirectUri string
}
