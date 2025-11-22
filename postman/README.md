# Postman Collection - Hotel Booking Microservices

## 📋 Overview
This Postman collection contains all 25 API endpoints for the Hotel Booking Microservices system, organized into 8 logical folders for easy testing.

## 🚀 Quick Start

### 1. Import Collection
1. Open Postman
2. Click **Import** button
3. Select `Hotel-Booking-Microservices.postman_collection.json`
4. Collection will be imported with all endpoints and variables

### 2. Setup Environment
The collection uses **Collection Variables** for dynamic data:
- `base_url`: http://localhost:8088/api/v1
- `access_token`: Auto-populated after customer login
- `admin_token`: Auto-populated after admin login
- `user_id`, `hotel_id`, `room_id`, `booking_id`, `payment_id`: Auto-populated from responses

### 3. Run the Flow

#### **Recommended Testing Flow:**

1. **Authentication** (Folder 1)
   - Run "Register Admin" → Creates admin account
   - Run "Login Admin" → Gets admin token (auto-saved)
   - Run "Register Customer" → Creates customer account
   - Run "Login Customer" → Gets customer token (auto-saved)

2. **Setup Hotel Data** (Folders 2-4, use Admin token)
   - Run "Create Hotel" → hotel_id auto-saved
   - Run "Create Room Type" → room_type_id auto-saved
   - Run "Create Room" → room_id auto-saved
   - Test "Update Hotel" ⭐ NEW
   - Test "Update Room" ⭐ NEW

3. **Booking Flow** (Folder 5, use Customer token)
   - Run "Create Booking" → booking_id auto-saved
   - Run "Get Payment by Booking ID" → payment_id auto-saved
   - Run "Payment Webhook" → Simulates payment confirmation
   - Run "Check-in Booking" → Changes status to checked_in

4. **Additional Features**
   - Test "Refund Payment" (Admin)
   - Test "List Notifications"
   - Test "Get Booking Aggregate"

## 📁 Collection Structure

### 1. Authentication (3 endpoints)
- Register Customer
- Register Admin
- Login Customer
- Login Admin
- Get User Profile

### 2. Hotel Management (5 endpoints)
- List Hotels (Public)
- Get Hotel by ID (Public)
- Create Hotel (Admin) 🔒
- **Update Hotel (Admin) 🔒 ⭐ NEW**
- **Delete Hotel (Admin) 🔒 ⭐ NEW**

### 3. Room Type Management (2 endpoints)
- List Room Types (Public)
- Create Room Type (Admin) 🔒

### 4. Room Management (5 endpoints)
- List Rooms (Public)
- **Get Room by ID (Public) ⭐ NEW**
- Create Room (Admin) 🔒
- **Update Room (Admin) 🔒 ⭐ NEW**
- **Delete Room (Admin) 🔒 ⭐ NEW**

### 5. Booking Management (5 endpoints)
- Create Booking 🔒
- List Bookings 🔒
- Get Booking by ID 🔒
- Check-in Booking 🔒
- Cancel Booking 🔒

### 6. Payment Management (3 endpoints)
- Get Payment by Booking ID 🔒
- Payment Webhook (Xendit Mock)
- Refund Payment (Admin) 🔒

### 7. Notifications (1 endpoint)
- List Notifications 🔒

### 8. Gateway Aggregation (1 endpoint)
- Get Booking Aggregate 🔒

**Total: 25 Endpoints**

## 🔐 Authentication

### Public Endpoints (No Auth Required)
- List Hotels
- Get Hotel by ID
- List Room Types
- List Rooms
- Get Room by ID

### Customer Endpoints (Requires `access_token`)
- All Booking operations
- Get Payment
- List Notifications

### Admin Endpoints (Requires `admin_token`)
- Create/Update/Delete Hotel
- Create Room Type
- Create/Update/Delete Room
- Refund Payment

## 🎯 Auto-Variable Extraction

The collection includes **Test Scripts** that automatically extract and save IDs:

```javascript
// Example: After "Login Customer"
if (pm.response.code === 200) {
    const response = pm.response.json();
    pm.collectionVariables.set('access_token', response.data.access_token);
}
```

This means you don't need to manually copy-paste IDs between requests!

## ⭐ New Features Highlighted

All newly implemented CRUD endpoints are marked with **⭐ NEW**:
- Update Hotel
- Delete Hotel (Soft Delete)
- Get Room by ID
- Update Room (Supports partial updates)
- Delete Room (Soft Delete)

## 🔄 Auto-Checkout Feature

The system includes a **CronJob** that runs daily at 10:00 AM to automatically:
- Find bookings with `checkout_date = today` AND `status = checked_in`
- Transition them to `completed` status
- Send notifications

**Note**: This is automatic and doesn't require API calls.

## 🧪 Testing Tips

1. **Run in Order**: Follow the recommended flow for best results
2. **Check Variables**: View Collection Variables to see auto-saved IDs
3. **Admin vs Customer**: Switch between tokens by changing the Authorization header
4. **Soft Delete**: Deleted hotels/rooms are not permanently removed (soft delete)
5. **Payment Webhook**: In production, this would be called by Xendit automatically

## 📝 Sample Test Scenario

```
1. Register Admin → Login Admin
2. Create Hotel "Grand Luxury Hotel"
3. Create Room Type "Deluxe Suite" (Rp 1,500,000/night)
4. Create Room "101" (available)
5. Register Customer → Login Customer
6. Create Booking (Dec 15-20, 2025)
7. Simulate Payment via Webhook
8. Check-in Booking
9. (Wait for auto-checkout at 10:00 AM on Dec 20)
```

## 🐛 Troubleshooting

**401 Unauthorized**
- Make sure you've logged in and token is saved
- Check if using correct token (admin vs customer)

**404 Not Found**
- Verify the ID variables are set (check Collection Variables)
- Ensure you've created the resource first

**400 Bad Request**
- Check request body format
- Verify required fields are present

## 📚 Additional Resources

- **Swagger UI**: http://localhost:8087
- **API Gateway**: http://localhost:8088
- **Adminer (Database)**: http://localhost:8089
- **README**: See main project README for detailed API documentation

---

**Created for**: Hotel Booking Microservices Tech Test  
**Total Endpoints**: 25  
**New Features**: 5 CRUD endpoints + Auto-Checkout CronJob
