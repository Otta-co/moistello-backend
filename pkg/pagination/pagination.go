package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func Parse(c *gin.Context) (page int, limit int, offset int) {
	page = 1
	limit = 20
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	offset = (page - 1) * limit
	return
}

func Defaults(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
