INSERT INTO sensors (id, name, location, installation_date, latitude, longitude) VALUES
(1, 'Sensor 1', 'Field A', '2024-01-01', 37.7749, -122.4194)
ON CONFLICT DO NOTHING;
