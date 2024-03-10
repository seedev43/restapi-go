package main

import (
	"ephoto/modules"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func main() {

	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Get("/ephoto/", func(c fiber.Ctx) error {
		text1 := c.Query("text1")
		text2 := c.Query("text2", "SeeDev")
		url := c.Query("url")
		if text1 == "" || url == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": "required parameter text1 and url",
			})
		}
		res, err := modules.Ephoto360(url, text1, text2)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"code":    http.StatusOK,
			"message": "success create image",
			"data":    res,
		})
	})

	log.Fatal(app.Listen(":3000"))

	// fmt.Println(ok)
}
