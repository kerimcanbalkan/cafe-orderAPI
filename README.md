# Cafe Order API

Cafe Order API is a RESTful API built using Go and the Gin framework to manage orders, users, and menus for a cafe or restaurant. The API provides authentication, order tracking, and statistics functionality.

## Features
- Menu management (create, retrieve, delete menu items)
- Order management (create, update, serve, close orders)
- User authentication and management
- Real-time order notifications via Server-Sent Events (SSE)
- Statistics for orders and employee performance
- Swagger documentation

## Technologies Used
- **Gin** - Web framework for Go
- **MongoDB** - NoSQL database for data storage
- **JWT Authentication** - Secure user authentication
- **Swagger** - API documentation
- **Server-Sent Events (SSE)** - Real-time notifications

## Installation

### Prerequisites
- Go 1.19+
- MongoDB

For development can use mongodb docker image
```sh
sudo docker run -d -p 2717:27017 -v ~/mymongo:/data/db --name mymongo mongo:latest
```

### Clone Repository
```sh
git clone https://github.com/kerimcanbalkan/cafe-orderAPI.git
cd cafe-orderAPI
```

### Install Dependencies
```sh
go mod tidy
```

### Configure Environment Variables
```
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=mongodb
SERVER_PORT=8000
DEFAULT_ADMIN_USERNAME=admin
DEFAULT_ADMIN_PASSWORD=password
SECRET=reallysecuresecret
```

## Running the API
```sh
go run main.go
```

## API Endpoints

### Swagger Documentation
Swagger documentation is available at:
```
GET /api/v1/swagger/index.html
```

### Menu Routes
| Method | Endpoint               | Description                          | Auth Required |
|--------|------------------------|--------------------------------------|--------------|
| GET    | `/api/v1/menu`          | Retrieve menu items                 | No           |
| POST   | `/api/v1/menu`          | Create a new menu item              | Admin        |
| DELETE | `/api/v1/menu/:id`      | Delete a menu item                  | Admin        |
| GET    | `/api/v1/menu/images/:filename` | Get menu item image         | No           |

### Order Routes
| Method | Endpoint                | Description                          | Auth Required |
|--------|-------------------------|--------------------------------------|--------------|
| POST   | `/api/v1/order/:table`   | Create a new order                  | No           |
| GET    | `/api/v1/order`          | Get all orders                      | Admin, Cashier, Waiter |
| PATCH  | `/api/v1/order/:id`      | Update an order                     | Admin, Cashier, Waiter |
| PATCH  | `/api/v1/order/serve/:id`| Mark an order as served             | Admin, Waiter |
| PATCH  | `/api/v1/order/close/:id`| Close an order                      | Admin, Cashier |
| GET    | `/api/v1/order/stats`    | Get order statistics                | Admin        |

### User Routes
| Method | Endpoint                  | Description                          | Auth Required |
|--------|---------------------------|--------------------------------------|--------------|
| POST   | `/api/v1/user`            | Create a new user                   | Admin        |
| GET    | `/api/v1/user`            | Get all users                       | Admin        |
| GET    | `/api/v1/user/:id`        | Get user details                    | Admin, Cashier, Waiter |
| GET    | `/api/v1/user/:id/stats`  | Get user statistics                 | Admin, Cashier, Waiter |
| GET    | `/api/v1/user/me`         | Get current user details            | Admin, Cashier, Waiter |
| POST   | `/api/v1/user/login`      | Authenticate user and get token     | No           |
| DELETE | `/api/v1/user/:id`        | Delete a user                       | Admin        |

### Real-time Updates
| Method | Endpoint         | Description                          | Auth Required |
|--------|-----------------|--------------------------------------|--------------|
| GET    | `/api/v1/events`| Server-Sent Events for live updates | Admin, Cashier, Waiter|

## Authentication
The API uses JWT for authentication. After logging in via `/api/v1/user/login`, include the token in the `Authorization` header:
```sh
Authorization: Bearer <token>
```
