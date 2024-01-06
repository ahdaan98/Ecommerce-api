package main

import (
	"ecommerce-mobilestore/controller"
	"ecommerce-mobilestore/initializer"
	"ecommerce-mobilestore/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	initializer.LoadEnvVariables()
	initializer.ConnectToDb()
	initializer.SyncDatabase()
}

func main() {
	r := gin.Default()

	user := r.Group("/user")
	user.POST("/signup", controller.SignUp)
	user.POST("/checkotp", controller.VerifyOtpAndSignupUser)
	user.POST("/login", controller.LoginUser)
	user.GET("/logout", controller.LogoutUser)
	user.GET("/list/products", middleware.RequireAuth, controller.UserLogged, controller.ListProducts)
	user.PUT("/add/cart/:id", middleware.RequireAuth, controller.AddToCart)
	user.GET("/cart/items", middleware.RequireAuth, controller.ListCart)
	user.DELETE("/delete/cart/item/:id", middleware.RequireAuth, controller.RemoveFromCart)
	user.POST("/checkout", middleware.RequireAuth, controller.ProceedToCheckout)
	user.POST("/place/order", middleware.RequireAuth, controller.PlaceOrderCOD)
	user.GET("/view/profile", middleware.RequireAuth, controller.Profile)
	user.GET("/view/orders", middleware.RequireAuth, controller.ViewOrders)
	user.GET("/view/address", middleware.RequireAuth, controller.ViewAddress)
	user.PUT("/edit/address/:id", middleware.RequireAuth, controller.EditAddress)
	user.PUT("/edit/profile", middleware.RequireAuth, controller.EditProfile)
	user.PUT("/cancel/order/:id", middleware.RequireAuth, controller.CancelOrder)

	admin := r.Group("/admin")
	admin.POST("/login", controller.Adminlogin)
	admin.GET("/logout", controller.Adminlogout)
	admin.GET("/users", middleware.RequireAuthAdmin, controller.AdminLogged, controller.ListofUsers)
	admin.PUT("/users/block/:id", middleware.RequireAuthAdmin, controller.BlockUser)
	admin.PUT("/users/unblock/:id", middleware.RequireAuthAdmin, controller.UnblockUser)
	admin.POST("/product/add", middleware.RequireAuthAdmin, controller.AdminLogged, controller.AddProduct)
	admin.PUT("/product/edit/:id", middleware.RequireAuthAdmin, controller.AdminLogged, controller.EditProduct)
	admin.DELETE("/product/delete/:id", middleware.RequireAuthAdmin, controller.AdminLogged, controller.DeleteProduct)
	admin.POST("/category/add", middleware.RequireAuthAdmin, controller.AddCategory)
	admin.PUT("/category/edit/:id", middleware.RequireAuthAdmin, controller.UpdateCategory)
	admin.DELETE("/category/delete/:id", middleware.RequireAuthAdmin, controller.DeleteCategory)
	admin.GET("/category/get/all", middleware.RequireAuthAdmin, controller.GetAllCategories)
	admin.GET("/list/orders", middleware.RequireAuthAdmin, controller.ListOrders)
	admin.PUT("/users/orders/cancel/:id", middleware.RequireAuthAdmin, controller.UserOrdersCancel)
	admin.PUT("/users/orders/status/edit/:id", middleware.RequireAuthAdmin, controller.EditOrderStatus)
	admin.POST("/add/product/img/:id",middleware.RequireAuthAdmin,controller.UploadProductImage)

	r.Run()
}
