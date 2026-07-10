package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
   c.JSON(http.StatusOK, gin.H{})
}

func Logout(c *gin.Context) {
   c.JSON(http.StatusOK, gin.H{})
}

func CreateUser(c *gin.Context) {
   c.JSON(http.StatusCreated, gin.H{})
}

func UpdateUser(c *gin.Context) {
   c.JSON(http.StatusOK, gin.H{})
}

func DeleteUser(c *gin.Context) {
   c.JSON(http.StatusNoContent, gin.H{})
}
