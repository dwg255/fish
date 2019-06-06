package router

import (
	"fish/game/controllers"
	"fish/game/service"
	"net/http"
)

func init()  {
	http.HandleFunc("/get_server_info", controllers.GetServerInfo)

	http.HandleFunc("/create_room", controllers.CreateRoom)
	http.HandleFunc("/create_public_room", controllers.CreatePublicRoom)
	http.HandleFunc("/enter_room", controllers.EnterRoom)
	http.HandleFunc("/ping", controllers.Ping)
	http.HandleFunc("/is_room_running", controllers.IsRoomRunning)
	http.HandleFunc("/enter_public_room", controllers.EnterPublicRoom)
	http.HandleFunc("/socket.io/",service.ServeWs)
}
