package handlers

import "github.com/gofiber/fiber/v2"

// Welcome handles the HTTP request to render the welcome page.
func Welcome(c *fiber.Ctx) error {
	return c.Render("welcome", nil, "layouts/main")
}
