-- Seed data for the sensors table

-- Add the essential sensor required by the default MQTT topic subscription
INSERT INTO sensors (id, name, location, installation_date)
VALUES ('444574498032128', 'Farm Sensor 1 (Default)', 'Main Farm Area', NOW())
ON CONFLICT (id) DO NOTHING; -- Avoid error if it somehow already exists

-- Add more sensor records below if needed

-- Add the individual sensor IDs from MQTT topics (using hex directly)
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C06270008D', 'Sensor 2CF7F1C06270008D', 'Device 444574498032128 - Sensor 1', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C0627000BC', 'Sensor 2CF7F1C0627000BC', 'Device 444574498032128 - Sensor 2', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C0627000C4', 'Sensor 2CF7F1C0627000C4', 'Device 444574498032128 - Sensor 3', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;
