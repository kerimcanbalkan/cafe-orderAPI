package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

func TestGetMenu(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		first := mtest.CreateCursorResponse(1, "testDB.menu", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Burger"},
			{Key: "description", Value: "A delicious beef burger"},
			{Key: "price", Value: 5.99},
			{Key: "category", Value: "Main"},
			{Key: "image", Value: "burger_image_url"},
		})
		second := mtest.CreateCursorResponse(1, "testDB.menu", mtest.NextBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Pizza"},
			{Key: "description", Value: "Cheese and tomato pizza"},
			{Key: "price", Value: 8.99},
			{Key: "category", Value: "Main"},
			{Key: "image", Value: "pizza_image_url"},
		})

		// Simulate cursor close
		killCursors := mtest.CreateCursorResponse(0, "testDB.menu", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// test route
		r := gin.Default()
		r.GET("/test/menu", menu.GetMenu(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test/menu", nil)
		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var menuResponse MenuResponse
		err := json.Unmarshal(w.Body.Bytes(), &menuResponse)
		assert.Nil(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Len(t, menuResponse.Data, 2) // Two items expected
		assert.Equal(t, "Burger", menuResponse.Data[0].Name)
		assert.Equal(t, "Pizza", menuResponse.Data[1].Name)
	})
}

func TestCreateMenu(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		mockClient := db.NewMockMongoClient(mt.Coll)
		menuItem := menu.MenuItem{
			Name:        "Coffee",
			Description: "Enjoy a freshly brewed cup of coffee, perfect for starting your day or taking a relaxing break.",
			Price:       5.99,
			Category:    "drink",
			Img:         "path/to/image.jpg",
		}

		body := new(bytes.Buffer)
		writer, err := generateMultipartForm(menuItem, body)
		if err != nil {
			fmt.Printf("Error generating multipart form: %v", err)
			return
		} // Test route
		r := gin.Default()
		r.POST("/test/menu", menu.CreateMenuItem(mockClient))

		// Create the request
		req := httptest.NewRequest(http.MethodPost, "/test/menu", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Perform the request
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var createResponse CreateResponse
		err = json.Unmarshal(w.Body.Bytes(), &createResponse)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Item added successfully", createResponse.Message)
	})

	mt.Run("custom duplicate key error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))

		mockClient := db.NewMockMongoClient(mt.Coll)
		menuItem := menu.MenuItem{
			Name:        "Coffee",
			Description: "Enjoy a freshly brewed cup of coffee, perfect for starting your day or taking a relaxing break.",
			Price:       5.99,
			Category:    "drink",
			Img:         "path/to/image.jpg",
		}

		body := new(bytes.Buffer)
		writer, err := generateMultipartForm(menuItem, body)
		if err != nil {
			fmt.Printf("Error generating multipart form: %v", err)
			return
		} // Test route
		r := gin.Default()
		r.POST("/test/menu", menu.CreateMenuItem(mockClient))

		// Create the request
		req := httptest.NewRequest(http.MethodPost, "/test/menu", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Perform the request
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var errorResponse ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.Nil(t, err)
		assert.Equal(t,
			fmt.Sprintf("Menu item named %s already exists", menuItem.Name),
			errorResponse.Error,
		)
	})
}

func TestMenuValidation(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("custom error validation", func(mt *mtest.T) {
		mockClient := db.NewMockMongoClient(mt.Coll)

		r := gin.Default()
		r.POST("/test/menu", menu.CreateMenuItem(mockClient))
		cases := []struct {
			name          string
			input         menu.MenuItem
			expectedError string
		}{
			{
				"Missing Name",
				menu.MenuItem{
					Description: "A refreshing beverage",
					Price:       5.99,
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Name is required",
			},
			{
				"Name Too Short",
				menu.MenuItem{
					Name:        "C",
					Description: "A refreshing beverage",
					Price:       5.99,
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Name must be at least 2 characters",
			},
			{
				"Name Too Long",
				menu.MenuItem{
					Name:        "A very long menu item name that exceeds the max limit A very long menu item name that exceeds the max limit",
					Description: "A refreshing beverage",
					Price:       5.99,
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Name must be at most 60 characters",
			},
			{
				"Missing Description",
				menu.MenuItem{
					Name:     "Coffee",
					Price:    5.99,
					Category: "drink",
					Img:      "path/to/image.jpg",
				},
				"Description is required",
			},
			{
				"Description Too Short",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "sh",
					Price:       5.99,
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Description must be at least 5 characters",
			},
			{
				"Description Too Long",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "very very very very very very very very very very very very very very very very very veryvery very long menu item description that should fail the validation.",
					Price:       5.99,
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Description must be at most 150 characters",
			},
			{
				"Invalid Price (Too Low)",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "A refreshing beverage",
					Price:       -1.0,
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Price must be greater than 0",
			},
			{
				"Missing Price",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "A refreshing beverage",
					Category:    "drink",
					Img:         "path/to/image.jpg",
				},
				"Price is required",
			},
			{
				"Missing Category",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "A refreshing beverage",
					Price:       5.99,
					Img:         "path/to/image.jpg",
				},
				"Category is required",
			},
			{
				"Category Too Short",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "A refreshing beverage",
					Price:       5.99,
					Category:    "a",
					Img:         "path/to/image.jpg",
				},
				"Category must be at least 2 characters",
			},
			{
				"Category Too Long",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "A refreshing beverage",
					Price:       5.99,
					Category:    "A very long category name that exceeds the max length A very long category name that exceeds the max length",
					Img:         "path/to/image.jpg",
				},
				"Category must be at most 60 characters",
			},
			{
				"Missing Image",
				menu.MenuItem{
					Name:        "Coffee",
					Description: "A refreshing beverage",
					Price:       5.99,
					Category:    "drink",
				},
				"Image file is required",
			},
		}

		for _, tc := range cases {
			mt.Log("TEST CASE", tc.name)
			body := new(bytes.Buffer)
			writer, err := generateMultipartForm(tc.input, body)
			if err != nil {
				fmt.Printf("Error generating multipart form: %v", err)
				return
			}
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test/menu", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			r.ServeHTTP(w, req)

			var response ErrorResponse
			json.Unmarshal(w.Body.Bytes(), &response)
			mt.Log("this is what has been returned", w.Body.String())

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, tc.expectedError, response.Error)

		}
	})
}

// Generates a 1x1 JPEG image in memory
func generatePlaceholderImage() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255}) // Red pixel

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generateMultipartForm(item menu.MenuItem, buffer *bytes.Buffer) (*multipart.Writer, error) {
	writer := multipart.NewWriter(buffer)

	// Create form fields
	if err := writer.WriteField("name", item.Name); err != nil {
		return nil, fmt.Errorf("failed to write field 'name': %v", err)
	}
	if err := writer.WriteField("description", item.Description); err != nil {
		return nil, fmt.Errorf("failed to write field 'description': %v", err)
	}
	if err := writer.WriteField("price", fmt.Sprintf("%.2f", item.Price)); err != nil {
		return nil, fmt.Errorf("failed to write field 'price': %v", err)
	}
	if err := writer.WriteField("category", item.Category); err != nil {
		return nil, fmt.Errorf("failed to write field 'category': %v", err)
	}

	if item.Img != "" {
		// Create a part for the image with custom headers
		h := make(textproto.MIMEHeader)
		h.Set(
			"Content-Disposition",
			fmt.Sprintf(`form-data; name="image"; filename="%s"`, item.Img),
		)
		h.Set("Content-Type", "image/jpeg")

		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, fmt.Errorf("could not create MIME part for file: %v", err)
		}
		// Generate a placeholder image
		imgBytes, err := generatePlaceholderImage()
		if err != nil {
			return nil, fmt.Errorf("Failed to generate image: %v", err)
		}

		// Write the image bytes into the part
		if _, err := part.Write(imgBytes); err != nil {
			return nil, fmt.Errorf("could not write image data: %v", err)
		}
	}

	// Close the writer to finalize the multipart data
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	return writer, nil
}

func TestDeleteMenuItem(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// Generate id
		id := primitive.NewObjectID().Hex()

		// Mock a deleted menu item response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: id},
				{Key: "name", Value: "Coffee"},
				{Key: "description", Value: "Enjoy a freshly brewed cup of coffee..."},
				{Key: "price", Value: 1.5},
				{Key: "category", Value: "drink"},
				{
					Key:   "image",
					Value: "uploads/77sF65eeRcVpCDpLjr5Wad",
				},
			}},
		})
		// Test route
		r := gin.Default()
		r.DELETE("/test/menu/:id", menu.DeleteMenuItem(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/test/menu/"+id, nil)

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	mt.Run("custom not found error", func(mt *mtest.T) {
		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		id := primitive.NewObjectID().Hex()
		// Mock a deleted user response
		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 0}},
		)

		// Test route
		r := gin.Default()
		r.DELETE("/test/menu/:id", menu.DeleteMenuItem(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/test/menu/"+id, nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetMenuItemImage(t *testing.T) {
	// Create a test router
	r := gin.Default()

	// Serve the route
	r.GET("/menu-item/image/:filename", menu.GetMenuItemImage)

	// Ensure the uploads directory exists
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Generate placeholder image
	imageData, err := generatePlaceholderImage()
	if err != nil {
		t.Fatalf("Failed to generate placeholder image: %v", err)
	}

	// Define the test image path
	testImage := uploadsDir + "/test-image.jpg"

	// Write the generated image to the file
	err = os.WriteFile(testImage, imageData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer os.Remove(testImage) // Cleanup after test

	// Create a test request to the route with a valid filename
	req, err := http.NewRequest("GET", "/menu-item/image/test-image.jpg", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder to capture the response
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assert that the status code is OK (200)
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert that the Content-Type header is correctly set (should detect as image/jpeg)
	assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
}

func TestGetMenuItemImage_NotFound(t *testing.T) {
	// Create a test router
	r := gin.Default()

	// Serve the route
	r.GET("/menu-item/image/:filename", menu.GetMenuItemImage)

	// Create a request for a non-existent file
	req, err := http.NewRequest("GET", "/menu-item/image/nonexistent.jpg", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder to capture the response
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	t.Log(w.Body.String())

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Image not found", response.Error)
}
