-- Seed data for the sensors table

-- Add the essential sensor required by the default MQTT topic subscription
INSERT INTO sensors (id, name, location, installation_date)
VALUES (444574498032128, 'Farm Sensor 1 (Default)', 'Main Farm Area', NOW())
ON CONFLICT (id) DO NOTHING; -- Avoid error if it somehow already exists

-- Add more sensor records below if needed
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
(1, 'Sensor 1', 'Field A', '2024-01-01', 37.7749, -122.4194)
ON CONFLICT DO NOTHING;
