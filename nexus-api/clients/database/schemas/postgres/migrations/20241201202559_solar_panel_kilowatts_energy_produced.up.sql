-- Create the table
CREATE TABLE solar_panel_kilowatts_energy_produced_data (
    id SERIAL PRIMARY KEY,          -- Auto-incrementing primary key
    date TIMESTAMP NOT NULL,        -- Date for the energy production data
    production INTEGER NOT NULL,     -- Amount of energy produced in kilowatts
    panel_id TEXT
);
