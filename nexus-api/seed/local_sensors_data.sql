INSERT INTO sensor_data (sensor_id, data, timestamp) VALUES
(1, 100, '2024-12-20 00:00:00'),
(1, 150, '2024-12-21 00:00:00'),
(1, 125, '2024-12-22 00:00:00'),
(1, 200, '2024-12-23 00:00:00'),
(1, 175, '2024-12-24 00:00:00')
ON CONFLICT DO NOTHING;