-- First, remove any duplicate entries keeping the most recent one
DELETE FROM sensor_temperature_data a
USING (
  SELECT sensor_id, date, MAX(id) as max_id
  FROM sensor_temperature_data
  GROUP BY sensor_id, date
  HAVING COUNT(*) > 1
) b
WHERE a.sensor_id = b.sensor_id 
  AND a.date = b.date 
  AND a.id < b.max_id;

-- Now add the unique constraint
ALTER TABLE sensor_temperature_data 
ADD CONSTRAINT unique_sensor_temperature_data 
UNIQUE (sensor_id, date);

-- Generate 6 months of hourly data for sensor_id 444574498032128
-- This uses generate_series to create timestamps and a function for temperature values
WITH hourly_timestamps AS (
  SELECT 
    generate_series(
      now() - interval '6 months',  -- Start date: 6 months ago
      now(),                        -- End date: current time
      interval '1 hour'             -- Interval: hourly
    ) AS timestamp
)
INSERT INTO sensor_temperature_data (sensor_id, date, soil_temperature)
SELECT
  444574498032128 AS sensor_id,  -- The specific sensor ID
  timestamp,
  -- Generate realistic soil temperature values that vary over time:
  -- Base temperature (15-25°C range) + daily cycle (±5°C) + seasonal variation (±10°C) + random noise (±2°C)
  -- This creates values that follow daily and seasonal patterns while having some randomness
  (20 + 5 * random()) + 
  (5 * sin(extract(epoch from timestamp) / 43200)) + -- Daily cycle (12 hours = 43200 seconds)
  (10 * sin(extract(epoch from timestamp) / 15778476)) + -- Seasonal cycle (6 months = ~15778476 seconds)
  (2 * random()) AS soil_temperature
FROM hourly_timestamps
ON CONFLICT (sensor_id, date) DO NOTHING;

-- Add a note about the data volume
-- This should generate approximately 4320 rows (6 months × ~30 days × 24 hours)
