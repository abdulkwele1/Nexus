-- Seed data for the sensors table

-- Add the essential sensor required by the default MQTT topic subscription
-- Removed sensor 444574498032128

-- Add more sensor records below if needed

-- Add the individual sensor IDs from MQTT topics (using hex directly)
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C0649007B3', 'Sensor 2CF7F1C0649007B3', 'Device 444574498032128 - Sensor 1', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C064900792', 'Sensor 2CF7F1C064900792', 'Device 444574498032128 - Sensor 2', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C064900787', 'Sensor 2CF7F1C064900787', 'Device 444574498032128 - Sensor 3', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;
