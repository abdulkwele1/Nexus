-- Add role column to login_authentications table
ALTER TABLE login_authentications 
ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

-- Add check constraint to ensure valid roles
ALTER TABLE login_authentications 
ADD CONSTRAINT check_valid_role 
CHECK (role IN ('user', 'admin', 'root_admin'));

-- Create index on role for faster queries
CREATE INDEX idx_login_authentications_role ON login_authentications(role);

-- Update existing users to have 'user' role (already set by default)
-- Set abdul as root_admin
UPDATE login_authentications 
SET role = 'root_admin' 
WHERE user_name = 'abdul';
