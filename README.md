# Hotel Booking System (Microservices Architecture)

## Repository Name

**hotel-booking-microservices**

---

# 📄 Project Overview

This project is a **Hotel Booking System** built using a **microservices architecture**, designed as a technical test to demonstrate backend engineering capability, clean architecture, SOLID principles, database design, payment gateway integration, and containerized deployments.

The system provides core hotel operations such as:

* Room management
* Availability search
* Booking (order)
* Payment via 3rd-party provider (Midtrans / Xendit)
* Webhook handling
* Check-in & Check-out
* Refund processing
* Authentication & Authorization

All services are containerized with Docker, orchestrated using Docker Compose, and use PostgreSQL as the primary database.

---

# 🧱 Architecture Summary

The system consists of **4–5 microservices**:

### 1. **Auth Service**

* User registration & login
* JWT authentication
* User roles (admin, customer)
* Middleware for service protection

### 2. **Room Service**

* CRUD Room Types
* CRUD Rooms
* Room availability per date
* Price per night & optional overrides

### 3. **Booking Service**

* Create booking (PENDING)
* Calculate price based on date range
* Manage booking lifecycle: Pending → Paid → Confirmed → Checked-in → Checked-out
* Cancel booking
* Initiate refund requests
* Communicate with Payment Service

### 4. **Payment Service**

* Integrates with **Midtrans or Xendit**
* Creates payment invoice/charge
* Handles webhook callbacks
* Verifies HMAC/Signature
* Trigger booking status changes
* Refund processing

### 5. **Notification Service** *(optional)*

* Sends email/SMS notifications for:

  * Booking confirmation
  * Payment success
  * Refund processing

---

# 🗄 Database

The project uses **PostgreSQL** with migration tooling (Prisma/TypeORM).

## Core Entities

* Users
* Room Types
* Rooms
* Room Inventory
* Bookings
* Room Bookings
* Payments
* Checkins

An ERD diagram is included in the documentation.

---

# 🔌 Payment Integration

The system integrates with:
**Midtrans** (Snap API + Notifications) or **Xendit** (Invoices)

### Payment Flow

1. Booking created with `PENDING` status
2. Booking Service requests Payment Service to create payment
3. Payment Service calls provider API → returns payment URL
4. User pays via hosted payment page
5. Provider fires webhook callback
6. Payment Service verifies signature & updates local DB
7. Payment Service notifies Booking Service → Booking becomes `CONFIRMED`
8. Refund (if requested) follows provider's refund API

---

# ▶️ Running the System

Make sure you have:

* Docker
* Docker Compose
* Node.js (optional for local runs)

### Start Everything

```bash
docker-compose up --build
```

### Services Included in `docker-compose.yml`

* auth-service
* room-service
* booking-service
* payment-service
* notification-service (optional)
* postgres

### Environment Variables

Each service contains an `.env.example` file.

Copy and configure:

```bash
cp services/auth/.env.example services/auth/.env
cp services/booking/.env.example services/booking/.env
cp services/payment/.env.example services/payment/.env
```

---

# 🧪 Testing

This project includes **unit tests & integration tests** focusing on:

* Business logic (price calculation, booking rules)
* Payment webhook verification
* API route testing (via supertest/Jest)
* Repository & service layer tests

Run tests:

```bash
npm test
```

---

# 📜 API Documentation

Included formats:

* **OpenAPI/Swagger (YAML)**
* **Postman Collection (JSON)**

Describes all endpoints, sample requests, responses, and error structures.

---

# 🧩 SOLID Principles

The project applies:

* **S**ingle Responsibility: Each microservice handles one domain
* **O**pen/Closed: Payment Provider uses interfaces + adapter pattern
* **L**iskov Substitution: Any payment provider can replace another
* **I**nterface Segregation: Small interfaces between modules
* **D**ependency Inversion: Service layer depends on abstractions

---

# 📁 Repository Structure

```
hotel-booking-microservices/
│
├── services/
│   ├── auth/
│   ├── room/
│   ├── booking/
│   ├── payment/
│   └── notification/
│
├── docker-compose.yml
├── docs/
│   ├── ERD.png
│   ├── openapi.yml
│   └── postman_collection.json
│
└── README.md
```

---

# 📝 Project Scope Summary

This project demonstrates:

* Scalable microservices
* Database design & migrations
* Payment integration with third-party providers
* Consistent API contracts
* Secure authentication
* Proper containerization
* Clean, testable backend architecture

---

# 👨‍💻 Author

Fitry Yuliani
