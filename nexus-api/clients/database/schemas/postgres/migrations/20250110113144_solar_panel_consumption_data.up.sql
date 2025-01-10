-- Create the table
CREATE TABLE solar_panel_consumption_data (
    id SERIAL PRIMARY KEY,          -- Auto-incrementing primary key
    date TIMESTAMP WITH TIME ZONE NOT NULL,        -- Date for the energy production data
    capacity_kWh DECIMAL NOT NULL,     -- Amount of energy stored
    consumed_kWh DECIMAL NOT NULL, --Amount of energy used not recycled back into the site
    panel_id INTEGER NOT NULL
);