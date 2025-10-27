CREATE TABLE
    users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(50) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- CREATE TABLE
--     tables (
--         id SERIAL PRIMARY KEY,
--         name VARCHAR(100) NOT NULL,
--         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
--     );
-- CREATE TABLE products (
--         id SERIAL PRIMARY KEY,
--         name VARCHAR(100) NOT NULL,
--         price NUMERIC(10, 2) NOT NULL,
--         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
--     );

CREATE TABLE
    events (
        id SERIAL PRIMARY KEY,
        user_id INT REFERENCES users (id),
        event_type VARCHAR(50) NOT NULL,
        event_subject VARCHAR(50) NOT NULL,
        event_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        payload JSON
    );