package menu

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MenuItem struct {
	Name        string
	Description string
	Price       float32
	Category    string
}

type Menu struct {
	ID          string
	Name        string
	Description string
	Items       []MenuItem
}

func GetMenu(c *gin.Context) {
	menu := Menu{
		Name:        "Winter Menu",
		Description: "Fresh new menu for winter",
		Items: []MenuItem{
			{
				Name:        "Espresso",
				Description: "A strong, black coffee made by forcing steam through finely ground coffee beans.",
				Price:       2.50,
				Category:    "Hot Drinks",
			},
			{
				Name:        "Cappuccino",
				Description: "Espresso topped with steamed milk and foam.",
				Price:       3.50,
				Category:    "Hot Drinks",
			},
			{
				Name:        "Fredo Espresso",
				Description: "Cold Espresso with lots of ice",
				Price:       3,
				Category:    "Cold Drinks",
			},
		},
	}
	c.JSON(http.StatusOK, gin.H{
		"menu": menu,
	})
}
