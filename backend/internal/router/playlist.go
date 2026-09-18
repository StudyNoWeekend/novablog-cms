package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterPlaylistRoutes 注册第三方歌单管理路由。
func RegisterPlaylistRoutes(r *gin.RouterGroup, playlistController *controller.PlaylistController, authMiddleware gin.HandlerFunc) {
	playlist := r.Group("/playlists")
	playlist.Use(authMiddleware)
	{
		playlist.POST("", playlistController.Create)
		playlist.GET("", playlistController.GetList)
		playlist.GET("/:id", playlistController.GetByID)
		playlist.PUT("/:id", playlistController.Update)
		playlist.DELETE("/:id", playlistController.Delete)
	}
}
