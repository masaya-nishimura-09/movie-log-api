package main

import (
	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler"
)

func main() {
    router := gin.Default()

    users := router.Group("/users")
    {
        users.POST("/login", handler.Login)
        users.POST("/register", handler.CreateUser)
        users.POST("/logout", handler.Logout)
        users.PUT("/:id", handler.UpdateUser)
        users.DELETE("/:id", handler.DeleteUser)
    }

    router.Run("localhost:8080")
}
