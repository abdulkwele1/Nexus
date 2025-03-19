CREATE TABLE sensor_temperature_data (
    id SERIAL PRIMARY KEY,
    sensor_id INTEGER NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    temperature DECIMAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);