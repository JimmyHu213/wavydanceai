package controller

import "github.com/gin-gonic/gin"

// SendError writes the standard error envelope with a stable code.
// code comes from common/errcode; message is the human-readable default the
// frontend falls back to when it has no translation for the code. httpStatus
// keeps whatever the original site used — this helper does not change status
// semantics.
func SendError(c *gin.Context, httpStatus int, code, message string) {
	c.JSON(httpStatus, gin.H{
		"success": false,
		"code":    code,
		"message": message,
	})
}
