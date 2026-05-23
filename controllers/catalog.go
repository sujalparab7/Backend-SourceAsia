package controllers

import (
	"sync"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Product struct{
	ID string `json:"id"`
	Name string `json:"name"`
	SKU string `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs[]string `json:"video_urls"`
}

type ProductListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SKU        string `json:"sku"`
	ImageCount int    `json:"image_count"`
	VideoCount int    `json:"video_count"`
}

var(
	productStore=make(map[string]*Product)
	existingSKUs=make(map[string]bool)
	catalogMu sync.Mutex
)

type CreateProductRequest struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"` // Optional [cite: 69]
	VideoURLs []string `json:"video_urls"` // Optional [cite: 70]
}

type AddMediaRequest struct {
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

func CreateProduct(c *gin.Context) {
	var input CreateProductRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body format"})
		return
	}

	if input.Name == "" || input.SKU == "" {
		c.JSON(400, gin.H{"error": "Name and SKU are required and cannot be empty"})
		return
	}

	catalogMu.Lock()
	defer catalogMu.Unlock()

	if existingSKUs[input.SKU] {
		c.JSON(409, gin.H{"error": "A product with this SKU already exists"})
		return
	}

	newID := fmt.Sprintf("prod_%d", len(productStore)+1)
	newProduct := &Product{
		ID:        newID,
		Name:      input.Name,
		SKU:       input.SKU,
		ImageURLs: input.ImageURLs,
		VideoURLs: input.VideoURLs,
	}

	if newProduct.ImageURLs == nil {
		newProduct.ImageURLs = []string{}
	}
	if newProduct.VideoURLs == nil {
		newProduct.VideoURLs = []string{}
	}

	productStore[newID] = newProduct
	existingSKUs[input.SKU] = true // Mark this SKU as taken!

	c.JSON(201, newProduct)
}

func GetProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	catalogMu.Lock()
	defer catalogMu.Unlock()

	var productList []ProductListItem

	for _, product := range productStore {
		summary := ProductListItem{
			ID:         product.ID,
			Name:       product.Name,
			SKU:        product.SKU,
			ImageCount: len(product.ImageURLs),
			VideoCount: len(product.VideoURLs),
		}
		productList = append(productList, summary)
	}

	totalItems := len(productList)
	if offset >= totalItems {
		c.JSON(200, []ProductListItem{})
		return
	}

	end := offset + limit
	if end > totalItems {
		end = totalItems // Prevent "index out of bounds" crash!
	}

	paginatedList := productList[offset:end]

	c.JSON(200, paginatedList)
}

func GetProductByID(c *gin.Context) {
	id := c.Param("id")

	catalogMu.Lock()
	product, exists := productStore[id]
	catalogMu.Unlock() // We can unlock immediately after reading

	if !exists {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(200, product)
}

func AddMedia(c *gin.Context) {
	id := c.Param("id")

	var input AddMediaRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	if len(input.ImageURLs) == 0 && len(input.VideoURLs) == 0 {
		c.JSON(400, gin.H{"error": "At least one of image_urls or video_urls is required"})
		return
	}

	if len(input.ImageURLs) > 20 || len(input.VideoURLs) > 20 {
		c.JSON(400, gin.H{"error": "Maximum 20 URLs allowed per array per request"})
		return
	}

	catalogMu.Lock()
	defer catalogMu.Unlock()

	product, exists := productStore[id]
	if !exists {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	product.ImageURLs = append(product.ImageURLs, input.ImageURLs...)
	product.VideoURLs = append(product.VideoURLs, input.VideoURLs...)

	c.JSON(200, gin.H{
		"message": "Media added successfully",
		"product": product,
	})
}