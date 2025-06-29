CREATE TABLE sensor_moisture_data (
    id SERIAL PRIMARY KEY,
    sensor_id VARCHAR(32) NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    soil_moisture DECIMAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id),
    UNIQUE (sensor_id, date)
);