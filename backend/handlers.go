package main

import (
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Price       int64     `json:"price"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
}

var sanitizeRe = regexp.MustCompile(`[^a-z0-9.]`)

// sanitizeFilename: chặn path traversal (filepath.Base) + chuẩn hoá tên file
func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	lower := strings.ToLower(base)
	return sanitizeRe.ReplaceAllString(lower, "-")
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// saveImage lưu file upload vào uploads/, trả về đường dẫn public (vd: /uploads/xxx.jpg)
func saveImage(c *gin.Context, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > maxUploadSize {
		return "", fmt.Errorf("ảnh vượt quá 5MB")
	}

	name := fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeFilename(fileHeader.Filename))
	dst := filepath.Join("uploads", name)

	// Gin tự chmod thư mục đích mỗi lần upload (mặc định 0750 nếu không truyền perm)
	// → phải truyền 0755 tường minh, nếu không Nginx (www-data) mất quyền đọc ảnh sau upload kế tiếp
	if err := c.SaveUploadedFile(fileHeader, dst, 0755); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// removeImage xoá file ảnh cũ trên đĩa, bỏ qua nếu không tồn tại
func removeImage(imagePath string) {
	if imagePath == "" {
		return
	}
	name := filepath.Base(imagePath)
	_ = os.Remove(filepath.Join("uploads", name))
}

func listProducts(c *gin.Context) {
	rows, err := db.Query(`SELECT id, name, price, description, image, created_at FROM products ORDER BY id DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.Image, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		products = append(products, p)
	}

	// Thiếu bước này: kết nối đứt giữa chừng → trả về danh sách thiếu mà không báo lỗi
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func getProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID không hợp lệ"})
		return
	}

	var p Product
	err = db.QueryRow(`SELECT id, name, price, description, image, created_at FROM products WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.Image, &p.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy sản phẩm"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func createProduct(c *gin.Context) {
	name := c.PostForm("name")
	priceStr := c.PostForm("price")
	description := c.PostForm("description")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Thiếu tên sản phẩm"})
		return
	}
	price, err := strconv.ParseInt(priceStr, 10, 64)
	if err != nil || price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Giá không hợp lệ"})
		return
	}

	image := ""
	if fileHeader, err := c.FormFile("image"); err == nil {
		image, err = saveImage(c, fileHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	var p Product
	err = db.QueryRow(
		`INSERT INTO products (name, price, description, image) VALUES ($1, $2, $3, $4)
		 RETURNING id, name, price, description, image, created_at`,
		name, price, description, image,
	).Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.Image, &p.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func updateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID không hợp lệ"})
		return
	}

	var p Product
	err = db.QueryRow(`SELECT id, name, price, description, image, created_at FROM products WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.Image, &p.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy sản phẩm"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if v, ok := c.GetPostForm("name"); ok && v != "" {
		p.Name = v
	}
	if v, ok := c.GetPostForm("price"); ok && v != "" {
		price, err := strconv.ParseInt(v, 10, 64)
		if err != nil || price < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Giá không hợp lệ"})
			return
		}
		p.Price = price
	}
	if v, ok := c.GetPostForm("description"); ok {
		p.Description = v
	}
	if fileHeader, err := c.FormFile("image"); err == nil {
		newImage, err := saveImage(c, fileHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		removeImage(p.Image)
		p.Image = newImage
	}

	_, err = db.Exec(
		`UPDATE products SET name = $1, price = $2, description = $3, image = $4 WHERE id = $5`,
		p.Name, p.Price, p.Description, p.Image, p.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func deleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID không hợp lệ"})
		return
	}

	var image string
	err = db.QueryRow(`SELECT image FROM products WHERE id = $1`, id).Scan(&image)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"message": "Không tìm thấy sản phẩm"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if _, err := db.Exec(`DELETE FROM products WHERE id = $1`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	removeImage(image)

	c.JSON(http.StatusOK, gin.H{"message": "Đã xoá sản phẩm"})
}
