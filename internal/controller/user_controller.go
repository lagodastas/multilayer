package controller

import (
	"multilayer/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	userService service.UserServiceInterface
}

func NewUserController(userService service.UserServiceInterface) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	// Получаем ID из URL
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	// Парсим входные данные
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Вызываем сервис
	user, err := c.userService.UpdateUser(uint(id), input.Username, input.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(user)
}

func (c *UserController) Register(ctx *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user, err := c.userService.RegisterUser(input.Username, input.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(user)
}

func (c *UserController) GetUser(ctx *fiber.Ctx) error {
	id, _ := strconv.Atoi(ctx.Params("id"))
	user, err := c.userService.GetUser(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	return ctx.JSON(user)
}
