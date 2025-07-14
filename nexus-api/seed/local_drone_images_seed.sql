-- Seed data for drone_images table
INSERT INTO drone_images (
    id,
    file_name,
    file_path,
    upload_date,
    file_size,
    mime_type,
    description,
    metadata
) VALUES 
(
    'e47c6e8b-1234-5678-90ab-cdef01234567',
    'drone_image_1.jpg',
    '/drone_images/e47c6e8b-1234-5678-90ab-cdef01234567',
    CURRENT_TIMESTAMP - INTERVAL '1 day',
    1024576,  -- 1MB
    'image/jpeg',
    'North field aerial view',
    '{"location": "North Field", "altitude": "50m", "camera": "DJI Phantom 4 Pro"}'::jsonb
),
(
    'f58d7f9c-2345-6789-01cd-ef0123456789',
    'drone_image_2.jpg',
    '/drone_images/f58d7f9c-2345-6789-01cd-ef0123456789',
    CURRENT_TIMESTAMP - INTERVAL '2 days',
    2048576,  -- 2MB
    'image/jpeg',
    'South field irrigation inspection',
    '{"location": "South Field", "altitude": "30m", "camera": "DJI Phantom 4 Pro"}'::jsonb
),
(
    '69e8a0bd-3456-789a-12ef-ab0123456789',
    'drone_image_3.jpg',
    '/drone_images/69e8a0bd-3456-789a-12ef-ab0123456789',
    CURRENT_TIMESTAMP - INTERVAL '3 days',
    1536576,  -- 1.5MB
    'image/jpeg',
    'East field crop monitoring',
    '{"location": "East Field", "altitude": "40m", "camera": "DJI Phantom 4 Pro"}'::jsonb
)
ON CONFLICT (id) DO NOTHING; 