USE iot_db;

CREATE TABLE device_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    device_id VARCHAR(50),
    value DOUBLE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);