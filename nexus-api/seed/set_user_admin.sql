-- Set a user as admin
-- Replace 'username' with the actual username you want to make admin
UPDATE login_authentications 
SET role = 'admin' 
WHERE user_name = 'username';

-- Verify the update
SELECT user_name, role 
FROM login_authentications 
WHERE user_name = 'username';

-- To set multiple users as admin at once:
-- UPDATE login_authentications 
-- SET role = 'admin' 
-- WHERE user_name IN ('user1', 'user2', 'user3');

