-- Create the table
CREATE TABLE solar_panel_yield_data (
    id SERIAL PRIMARY KEY,          -- Auto-incrementing primary key
    date TIMESTAMP WITH TIME ZONE NOT NULL,        -- Date for the energy production data
    kwh_yield DECIMAL NOT NULL,     -- Amount of energy produced in kilowatts
    panel_id INTEGER NOT NULL
);
