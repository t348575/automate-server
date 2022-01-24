package main

// import (
// 	"github.com/go-playground/validator/v10"
// 	"github.com/gofiber/fiber/v2"
// )

// type Job struct{
//     Type          string `validate:"required,min=3,max=32"`
//     Salary        int    `validate:"required,number"`
// }

// type User struct{
//     Name          string  `validate:"required,min=3,max=32"`
//     Password    string `validate:"required,min=8,max=32,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#\$%\^&\*])(?=.{8,})"`
// }

// type ErrorResponse struct {
//     FailedField string
//     Tag         string
//     Value       string
// }

// type Organization struct{
// 	DisplayName string `validate:"required,alphanum,min=1,max=32"`
// 	Password    string `validate:"required,min=8,max=32,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#\$%\^&\*])(?=.{8,})"`
// 	Name string `validate:"required,alphanum,min=1,max=64"`
// 	Owner User `validate:"required,dive"`
// }

// func ValidateStruct(user User) []*ErrorResponse {
//     var errors []*ErrorResponse
//     validate := validator.New()
//     err := validate.Struct(user)
//     if err != nil {
//         for _, err := range err.(validator.ValidationErrors) {
//             var element ErrorResponse
//             element.FailedField = err.StructNamespace()
//             element.Tag = err.Tag()
//             element.Value = err.Param()
//             errors = append(errors, &element)
//         }
//     }
//     return errors
// }

// func InitOrg(router fiber.Router) {
// 	casdoor.InitConfig()
// 	router.Post("/", createOrg)
// }

// func createOrg(c *fiber.Ctx) error {
// 	user := new(User)

//     if err := c.BodyParser(user); err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "message": err.Error(),
//         })
//     }

//     errors := ValidateStruct(*user)
//     if errors != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(errors)
//     }
// 	return c.JSON(user)
// }