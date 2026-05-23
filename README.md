# Go Backend API - Rate Limiter & Product Catalog

This repository contains the backend engineering assignment featuring a concurrency-safe Rate Limiter and an optimized Product Catalog API built in Go using the Gin framework.

##  Setup & Execution

This application runs two separate APIs concurrently on different ports to simulate a microservices environment.

1. Ensure you have Go installed on your machine.
2. Clone this repository and navigate to the project root.
3. Run the following command to download the dependencies and boot the servers:

\`\`\`bash
go run .
\`\`\`

**Ports:**
* **Port 8080:** Rate Limiter API (`/request`, `/stats`)
* **Port 8081:** Product Catalog API (`/products`)

---

##  Part 1: Rate Limiter Architecture

The Rate Limiter tracks incoming requests by `user_id` and enforces a strict limit of 5 accepted requests per 60 seconds.

### Technical Design Decisions
* **Fixed Window Algorithm:** I implemented a fixed 1-minute window (`time.Since(WindowStart)`) rather than a sliding window. This is highly performant, requires less memory per user, and perfectly satisfies the requirement without over-engineering the time checks.
* **Concurrency Safety:** Since Go maps panic upon simultaneous reads/writes, all access to the `rateLimiter` map is strictly protected by a `sync.Mutex`. 

### Production Limitations (In-Memory vs. Distributed)
Currently, the rate limit data is stored in server RAM. This presents two limitations in a real-world production environment:
1. **Data Loss:** If the Go process restarts or crashes, all user history is wiped.
2. **Horizontal Scaling:** If this API is deployed behind a load balancer with multiple instances (e.g., 3 separate Go servers), they cannot share memory. A user could technically get 15 requests (5 per server).
* **The Production Fix:** In a real deployment, the in-memory Go map would be replaced with a distributed **Redis cache**, allowing all server instances to read and update a centralized, ultra-fast counter.

---

##  Part 2: Product Catalog & Production Data Modeling

The Product Catalog API manages products with associated images and videos, enforcing unique SKUs and handling large datasets efficiently.

### Current Implementation (In-Memory)
To meet the strict performance requirement for `GET /products`, the application uses two distinct data models:
1. `Product`: Stores the full entity including the massive slices of image and video URLs.
2. `ProductListItem`: A lightweight struct used specifically for the list view. Instead of passing arrays over the network, it calculates and returns `image_count` and `video_count`, ensuring the list endpoint remains lightning-fast even with thousands of products.
* *Note:* A secondary `existingSKUs` map is used as a fast-lookup table to achieve `O(1)` time complexity when validating duplicate SKUs during product creation.

### PostgreSQL Production Translation
If this service were moved to a relational database like PostgreSQL for production, the massive arrays of URLs would not be stored in a single column. The schema would be normalized into two tables:

**1. `products` Table:**
* `id` (Primary Key, UUID)
* `name` (VARCHAR)
* `sku` (VARCHAR, UNIQUE constraint to prevent duplicates at the DB level)

**2. `product_media` Table:**
* `id` (Primary Key, UUID)
* `product_id` (Foreign Key referencing `products.id`)
* `media_type` (ENUM: 'image' or 'video')
* `url` (TEXT)

For the `GET /products` list view, a `COUNT()` aggregation with a `GROUP BY` clause would be utilized in the SQL query to return the counts without ever pulling the raw URLs into the backend server's memory.

### Media & CDN Integration
This API handles the URLs as strings, not the binary media files themselves. In a production environment, the client (frontend or mobile app) would upload the raw image directly to a Cloud Delivery Network (CDN) like AWS S3 or Cloudflare. The CDN would return a fast, distributed URL, and the client would then send that URL to this API to be securely saved in the database.

---

##  Testing Commands

Here are some example `curl` commands to easily test the endpoints.

**Rate Limiter (Port 8080)**
\`\`\`bash
# Hit the endpoint to increment the counter
curl -i -X POST http://localhost:8080/request -H "Content-Type: application/json" -d "{\"user_id\": \"user_123\", \"payload\": {\"some_data\": 42}}"

# Check the global statistics
curl -i http://localhost:8080/stats
\`\`\`

**Product Catalog (Port 8081)**
\`\`\`bash
# Create a Product
curl -i -X POST http://localhost:8081/products -H "Content-Type: application/json" -d "{\"name\": \"Gaming Mouse\", \"sku\": \"MOUSE-001\", \"image_urls\": [\"http://example.com/img1.jpg\"]}"

# Get highly optimized list of all products (with pagination)
curl -i "http://localhost:8081/products?limit=10&offset=0"

# Add new media to an existing product
curl -i -X POST http://localhost:8081/products/prod_1/media -H "Content-Type: application/json" -d "{\"image_urls\": [\"http://example.com/img2.jpg\"], \"video_urls\": []}"

# Get a specific product by ID
curl -i http://localhost:8081/products/prod_1
\`\`\`
