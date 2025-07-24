-- Real sensor definitions - no mock data
INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C0649007B3', 'Sensor B3', 'Farm Sensor Location 1', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C064900792', 'Sensor 92', 'Farm Sensor Location 2', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;

INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
('2CF7F1C064900787', 'Sensor 87', 'Farm Sensor Location 3', NOW(), 37.7749, -122.4194)
ON CONFLICT (id) DO NOTHING;
