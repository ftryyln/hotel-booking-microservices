CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS hotels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    address TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS room_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hotel_id UUID NOT NULL REFERENCES hotels(id),
    name TEXT NOT NULL,
    capacity INT NOT NULL,
    base_price NUMERIC NOT NULL,
    amenities TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_room_types_hotel ON room_types(hotel_id);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    number TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'available'
);
CREATE INDEX IF NOT EXISTS idx_rooms_room_type ON rooms(room_type_id);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    status TEXT NOT NULL,
    total_price NUMERIC NOT NULL,
    total_nights INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES bookings(id) UNIQUE,
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    provider TEXT NOT NULL,
    payment_url TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID NOT NULL REFERENCES payments(id),
    amount NUMERIC NOT NULL,
    reason TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_refunds_payment ON refunds(payment_id);

CREATE TABLE IF NOT EXISTS checkins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES bookings(id) UNIQUE,
    check_in_at TIMESTAMPTZ,
    check_out_at TIMESTAMPTZ
);
