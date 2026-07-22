package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // trên VPS, systemd nạp env qua EnvironmentFile

	if err := connectDB(); err != nil {
		log.Fatal("❌ Không kết nối được PostgreSQL: ", err)
	}
	log.Println("✅ Đã kết nối PostgreSQL")

	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("❌ Không tạo được thư mục uploads: ", err)
	}
	// MkdirAll áp umask của user hệ thống → có thể ra 750 thay vì 755.
	// Ép lại tường minh để Nginx (www-data) luôn đọc được ảnh.
	if err := os.Chmod("uploads", 0755); err != nil {
		log.Fatal("❌ Không chỉnh được quyền uploads: ", err)
	}

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"}) // chỉ tin Nginx chạy cùng máy
	r.MaxMultipartMemory = 8 << 20

	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		api.GET("/health", healthCheck)
		api.GET("/products", listProducts)
		api.GET("/products/:id", getProduct)
		api.POST("/products", createProduct)
		api.PUT("/products/:id", updateProduct)
		api.DELETE("/products/:id", deleteProduct)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("🚀 API chạy tại http://localhost:" + port)

	// Bọc log.Fatal — nếu không, cổng bận sẽ khiến app thoát IM LẶNG
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Không khởi động được server: ", err)
	}
}
