INSERT INTO sensor_temperature_data (date, temperature, sensor_id) VALUES
('2024-12-20', 100, 1),
('2024-12-21', 150, 1),
('2024-12-22', 125, 1),
('2024-12-23', 200, 1),
('2024-12-24', 175, 1)
ON CONFLICT DO NOTHING;
