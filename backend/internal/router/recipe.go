package router

import (
	"novablog/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterRecipeRoutes 注册美食菜谱管理路由。
func RegisterRecipeRoutes(r *gin.RouterGroup, recipeController *controller.RecipeController, authMiddleware gin.HandlerFunc) {
	recipes := r.Group("/recipes")
	recipes.Use(authMiddleware)
	{
		recipes.POST("", recipeController.Create)
		recipes.GET("", recipeController.GetList)
		recipes.GET("/:id", recipeController.GetByID)
		recipes.PUT("/:id", recipeController.Update)
		recipes.DELETE("/:id", recipeController.Delete)
	}
}
