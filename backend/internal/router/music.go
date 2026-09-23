package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterMusicRoutes 注册音乐播放器管理路由。
func RegisterMusicRoutes(r *gin.RouterGroup, musicController *controller.MusicController, authMiddleware gin.HandlerFunc) {
	music := r.Group("/music")
	music.Use(authMiddleware)
	{
		// 歌曲
		music.POST("/songs", musicController.CreateSong)
		music.POST("/songs/batch", musicController.BatchCreateSongs)
		music.GET("/songs", musicController.GetSongList)
		music.GET("/songs/:id", musicController.GetSongByID)
		music.PUT("/songs/:id", musicController.UpdateSong)
		music.DELETE("/songs/:id", musicController.DeleteSong)

		// 解析
		music.POST("/parse", musicController.ParseMusic)
		music.GET("/parse/:task_id", musicController.GetParseStatus)

		// 获取播放地址（返回B站官方外链播放器地址，前端用 iframe 内嵌播放）
		music.GET("/audio-url/:song_id", musicController.GetAudioURL)
	}
}
