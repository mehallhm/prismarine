package router

import (
	"prismarine/shard/manager"

	"github.com/gofiber/fiber/v2"
)

func Create(m *manager.Manager) *fiber.App {
	router := fiber.New()

	router.Use(func(c *fiber.Ctx) error {
		c.Locals("manager", m)
		return c.Next()
	})

	instance := router.Group("/instance")
	_ = instance

	return router
}
