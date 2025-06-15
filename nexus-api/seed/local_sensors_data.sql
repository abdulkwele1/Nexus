-- Seed data for the sensors table

-- Add the essential sensor required by the default MQTT topic subscription
INSERT INTO sensors (id, name, location, installation_date)
VALUES (444574498032128, 'Farm Sensor 1 (Default)', 'Main Farm Area', NOW())
ON CONFLICT (id) DO NOTHING; -- Avoid error if it somehow already exists

-- Add more sensor records below if needed
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
(1, 'Sensor 1', 'Field A', '2024-01-01', 37.7749, -122.4194)
ON CONFLICT DO NOTHING;

-- Add the individual sensor IDs from MQTT topics (converted from hex to decimal)
-- 2CF7F1C06270008D (hex) = 3273093830958964877 (decimal)
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
(3273093830958964877, 'Sensor 2CF7F1C06270008D', 'Device 444574498032128 - Sensor 1', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

-- 2CF7F1C0627000BC (hex) = 3273093830958965948 (decimal)
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
(3273093830958965948, 'Sensor 2CF7F1C0627000BC', 'Device 444574498032128 - Sensor 2', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

-- 2CF7F1C0627000C4 (hex) = 3273093830958965956 (decimal)
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
(3273093830958965956, 'Sensor 2CF7F1C0627000C4', 'Device 444574498032128 - Sensor 3', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;
