-- Sample battery data for existing sensors
INSERT INTO sensor_battery_data (sensor_id, date, battery_level, voltage) VALUES
('2CF7F1C0649007B3', NOW() - INTERVAL '1 day', 85.5, 3.7),
('2CF7F1C06490079D', NOW() - INTERVAL '2 days', 90.0, 3.8),
('2CF7F1C0649007B9', NOW() - INTERVAL '1 day', 75.2, 3.5),
('2CF7F1C064900792', NOW() - INTERVAL '2 days', 80.0, 3.6),
('2CF7F1C064900787', NOW() - INTERVAL '1 day', 95.0, 3.9),
('2CF7F1C0649007C6', NOW() - INTERVAL '2 days', 98.5, 4.0)
ON CONFLICT (sensor_id, date) DO UPDATE 
SET battery_level = EXCLUDED.battery_level,
    voltage = EXCLUDED.voltage;