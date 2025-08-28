-- Create sensor battery data table
CREATE TABLE IF NOT EXISTS sensor_battery_data (
    id SERIAL PRIMARY KEY,
    sensor_id VARCHAR(255) NOT NULL REFERENCES sensors(id),
    date TIMESTAMP NOT NULL,
    battery_level FLOAT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sensor_id, date)
);