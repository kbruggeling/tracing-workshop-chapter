CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (name, email) VALUES 
    ('John Doe', 'john.doe@example.com'),
    ('Jane Smith', 'jane.smith@example.com'),
    ('Bob Wilson', 'bob.wilson@example.com'),
    ('Alice Brown', 'alice.brown@example.com'),
    ('Charlie Davis', 'charlie.davis@example.com'),
    ('Diana Miller', 'diana.miller@example.com'),
    ('Frank Johnson', 'frank.johnson@example.com'),
    ('Grace Lee', 'grace.lee@example.com'),
    ('Henry Taylor', 'henry.taylor@example.com'),
    ('Isabel Garcia', 'isabel.garcia@example.com');