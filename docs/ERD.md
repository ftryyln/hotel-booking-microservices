# Entity Relationship Diagram

| Entity | Primary Key | Key Fields | Relations |
|--------|-------------|------------|-----------|
| users | id UUID | email, role | 1 -* bookings |
| hotels | id UUID | name, address | 1 -* room_types |
| room_types | id UUID | hotel_id, base_price | 1 -* rooms, * -1 hotels |
| rooms | id UUID | room_type_id, number | * -1 room_types |
| bookings | id UUID | user_id, room_type_id, status, total_price | * -1 users, * -1 room_types, 1-1 payments, 1-1 checkins |
| payments | id UUID | booking_id, amount, provider | 1-1 bookings, 1 -* refunds |
| refunds | id UUID | payment_id, amount | * -1 payments |
| checkins | id UUID | booking_id, timestamps | 1-1 bookings |
