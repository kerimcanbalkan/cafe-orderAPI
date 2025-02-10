// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Kerimcan Balkan",
            "url": "https://github.com/kerimcanbalkan",
            "email": "kerimcanbalkan@gmail.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/events": {
            "get": {
                "description": "Establishes an SSE connection to receive real-time updates.",
                "produces": [
                    "text/event-stream"
                ],
                "tags": [
                    "SSE"
                ],
                "summary": "Handle SSE connection",
                "responses": {
                    "200": {
                        "description": "SSE stream opened",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/menu": {
            "get": {
                "description": "Fetches the entire menu from the database",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "menu"
                ],
                "summary": "Get all menu items",
                "responses": {
                    "200": {
                        "description": "List of menu items",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/menu.MenuItem"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Adds a new item to the menu with an image upload. Only accessible by users with the \"admin\" role.",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "menu"
                ],
                "summary": "Create a new menu item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name of the item",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Description of the item",
                        "name": "description",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "number",
                        "description": "Price of the item",
                        "name": "price",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Category of the item",
                        "name": "category",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "file",
                        "description": "Image file",
                        "name": "image",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Item added successfully",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/menu/images/{filename}": {
            "get": {
                "description": "Retrieves the image of a menu item by filename. This route is publicly accessible.",
                "tags": [
                    "menu"
                ],
                "summary": "Get the image of a menu item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Filename of the image",
                        "name": "filename",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Image file",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "404": {
                        "description": "Image not found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/menu/{id}": {
            "delete": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Deletes a menu item and its related image. Only accessible by users with the \"admin\" role.",
                "tags": [
                    "menu"
                ],
                "summary": "Delete a menu item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the menu item to delete",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Menu item not found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/order": {
            "get": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Retrieves all orders for admin, cashier, and waiter roles",
                "tags": [
                    "order"
                ],
                "summary": "Get all orders",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "Filter by closed status (true/false)",
                        "name": "is_closed",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Filter by served status (true/false)",
                        "name": "served",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Filter by table number",
                        "name": "table",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by order date (YYYY-MM-DD)",
                        "name": "date",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of orders with their IDs",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/order.Order"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/order/complete/{id}": {
            "patch": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admin and cashier roles to mark an order as complete",
                "tags": [
                    "order"
                ],
                "summary": "Mark an order as complete",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Order ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Order completed successfully",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "404": {
                        "description": "Order not found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/order/serve/{id}": {
            "patch": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admin and waiter roles to mark an order as served",
                "tags": [
                    "order"
                ],
                "summary": "Mark an order as served",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Order ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Order served successfully",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "404": {
                        "description": "Order not found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/order/stats": {
            "get": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Fetches statistics for a specific date range.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Statistics"
                ],
                "summary": "Get statistics for a given date range.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The date for which to fetch the statistics (format: yyyy-mm-dd). Defaults to today's date if not provided.",
                        "name": "from",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The date for which to fetch the statistics (format: yyyy-mm-dd). Defaults to today's date if not provided.",
                        "name": "to",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Order statistics data",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Invalid date format",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Failed to fetch statistics",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/order/{id}": {
            "patch": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admin, cashier, and waiter roles to update an order",
                "tags": [
                    "order"
                ],
                "summary": "Update an existing order",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Order ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Order update details",
                        "name": "order",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/order.orderRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Order updated successfully",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Invalid request"
                    },
                    "404": {
                        "description": "Order not found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/order/{table}": {
            "post": {
                "description": "Creates a new order for a specific table and saves it in the database",
                "tags": [
                    "order"
                ],
                "summary": "Create a new order",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Table number",
                        "name": "table",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Order details",
                        "name": "order",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/order.orderRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Order created successfully",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Invalid request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user": {
            "get": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admin role to retrieve a list of all users",
                "tags": [
                    "user"
                ],
                "summary": "Retrieve all users",
                "responses": {
                    "200": {
                        "description": "List of users",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admin role to create a new user",
                "tags": [
                    "user"
                ],
                "summary": "Create a new user",
                "parameters": [
                    {
                        "description": "User details",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "User created successfully",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Invalid request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/login": {
            "post": {
                "description": "Allows users to log in by providing username and password",
                "tags": [
                    "user"
                ],
                "summary": "User login",
                "parameters": [
                    {
                        "description": "Login details",
                        "name": "loginBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.LoginBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "JWT token and expiration time",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Invalid request"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/{id}": {
            "delete": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admin role to delete a user by their ID",
                "tags": [
                    "user"
                ],
                "summary": "Delete a user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "User deleted successfully"
                    },
                    "400": {
                        "description": "Invalid ID"
                    },
                    "404": {
                        "description": "User not found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/users/{id}/stats": {
            "get": {
                "security": [
                    {
                        "bearerToken": []
                    }
                ],
                "description": "Allows admins to retrieve all user statistics. Regular users can only retrieve their own statistics.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get user statistics for a given date range",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Start date (YYYY-MM-DD). Defaults to today.",
                        "name": "from",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "End date (YYYY-MM-DD). Defaults to today.",
                        "name": "to",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Statistics data",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Invalid input",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "menu.MenuItem": {
            "type": "object",
            "required": [
                "category",
                "description",
                "image",
                "name",
                "price"
            ],
            "properties": {
                "category": {
                    "type": "string",
                    "maxLength": 60,
                    "minLength": 2
                },
                "description": {
                    "type": "string",
                    "maxLength": 150,
                    "minLength": 5
                },
                "id": {
                    "type": "string"
                },
                "image": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "maxLength": 60,
                    "minLength": 2
                },
                "price": {
                    "type": "number"
                }
            }
        },
        "order.Order": {
            "type": "object",
            "required": [
                "items"
            ],
            "properties": {
                "closedAt": {
                    "type": "string"
                },
                "closedBy": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "handledBy": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/menu.MenuItem"
                    }
                },
                "servedAt": {
                    "type": "string"
                },
                "tableNumber": {
                    "type": "integer"
                },
                "totalPrice": {
                    "type": "number"
                }
            }
        },
        "order.orderRequest": {
            "type": "object",
            "required": [
                "items"
            ],
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/menu.MenuItem"
                    }
                }
            }
        },
        "user.LoginBody": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.User": {
            "type": "object",
            "required": [
                "email",
                "gender",
                "name",
                "password",
                "role",
                "surname",
                "username"
            ],
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "gender": {
                    "type": "string",
                    "enum": [
                        "male",
                        "female"
                    ]
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "maxLength": 20,
                    "minLength": 3
                },
                "password": {
                    "type": "string",
                    "maxLength": 20,
                    "minLength": 8
                },
                "role": {
                    "type": "string",
                    "enum": [
                        "admin",
                        "cashier",
                        "waiter"
                    ]
                },
                "surname": {
                    "type": "string",
                    "maxLength": 20,
                    "minLength": 3
                },
                "username": {
                    "type": "string",
                    "maxLength": 20,
                    "minLength": 3
                }
            }
        }
    },
    "securityDefinitions": {
        "bearerToken": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1",
	Host:             "localhost:8000",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Cafe Order API",
	Description:      "An API for cafe owners to manage menus, process orders, and oversee fulfillment.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
