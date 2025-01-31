INSERT INTO solar_panel_consumption_data (date, capacity_kwh,consumed_kwh, panel_id) VALUES
('2024-12-20', 100,90,1),
('2024-12-21', 150,80,1),
('2024-12-22', 125,100,1),
('2024-12-23', 200,120,1),
('2024-12-24', 175,150,1)
ON CONFLICT DO NOTHING;
