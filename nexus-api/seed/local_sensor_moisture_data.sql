-- Generate 6 months of hourly data for sensor_id 444574498032128
-- This uses generate_series to create timestamps and a random function for moisture values
WITH hourly_timestamps AS (
  SELECT 
    generate_series(
      now() - interval '6 months',  -- Start date: 6 months ago
      now(),                        -- End date: current time
      interval '1 hour'             -- Interval: hourly
    ) AS timestamp
)
INSERT INTO sensor_moisture_data (date, soil_moisture, sensor_id)
SELECT
  timestamp,
  -- Generate realistic soil moisture values that vary over time:
  -- Base moisture (20-60 range) + sine wave variation (±15) + random noise (±5)
  -- This creates values that smoothly oscillate while having some randomness
  (40 + 20 * random()) + 
  (15 * sin(extract(epoch from timestamp) / 86400)) + 
  (5 * random()) AS soil_moisture,
  444574498032128 AS sensor_id  -- The specific sensor ID
FROM hourly_timestamps
ON CONFLICT (date, sensor_id) DO NOTHING;

-- Add a note about the data volume
-- This should generate approximately 4320 rows (6 months × ~30 days × 24 hours)
