CREATE TABLE sensor_moisture_data (
    sensor_id INTEGER NOT NULL,
    id SERIAL PRIMARY KEY,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    soil_moisture DECIMAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);

ALTER TABLE sensors
ADD COLUMN latitude DOUBLE PRECISION,
ADD COLUMN longitude DOUBLE PRECISION;