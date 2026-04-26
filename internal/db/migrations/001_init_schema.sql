--liquibase formatted sql

--changeset dolmitya:01-create-guest
CREATE TABLE IF NOT EXISTS guest
(
    id           UUID PRIMARY KEY,
    last_name    VARCHAR(100) NOT NULL,
    first_name   VARCHAR(100) NOT NULL,
    middle_name  VARCHAR(100),
    birth_date   DATE NOT NULL,
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    created_at   TIMESTAMP NOT NULL DEFAULT now()
);

--changeset dolmitya:02-create-room
CREATE TABLE IF NOT EXISTS room
(
    id         UUID PRIMARY KEY,
    floor      INT NOT NULL CHECK (floor >= 0),
    number     VARCHAR(20) NOT NULL,
    capacity   INT NOT NULL CHECK (capacity > 0),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uk_room_number UNIQUE (number)
);

CREATE INDEX IF NOT EXISTS idx_room_floor ON room(floor);

--changeset dolmitya:03-create-booking
CREATE TABLE IF NOT EXISTS booking
(
    id         UUID PRIMARY KEY,
    room_id    UUID NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time   TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_booking_room
        FOREIGN KEY (room_id)
            REFERENCES room (id)
            ON DELETE RESTRICT,
    CONSTRAINT chk_booking_time CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_booking_room_id ON booking(room_id);
CREATE INDEX IF NOT EXISTS idx_booking_time_range ON booking(room_id, start_time, end_time);

--changeset dolmitya:04-create-booking-guest
CREATE TABLE IF NOT EXISTS booking_guest
(
    booking_id UUID NOT NULL,
    guest_id   UUID NOT NULL,
    CONSTRAINT pk_booking_guest PRIMARY KEY (booking_id, guest_id),
    CONSTRAINT fk_bg_booking
        FOREIGN KEY (booking_id)
            REFERENCES booking (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_bg_guest
        FOREIGN KEY (guest_id)
            REFERENCES guest (id)
            ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_bg_guest_id ON booking_guest(guest_id);
