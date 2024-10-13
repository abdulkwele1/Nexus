INSERT INTO "login_authentications"
 ("id", "user_name", "password_hash")
 VALUES
 (DEFAULT, 'levi', '$2a$10$HqQx4jxUzfQm1fZYUZRLbOBaMNWHmhSmweH03rl0EykgE4BNfDciO'),
 (DEFAULT, 'abdul', '$2a$14$KXCe7VMOjZdf/BwSKIFLxu2FRHcr.DAQntjq8OfdqQI69EOQz4gHW'),
 (DEFAULT, 'demo', '$2a$10$5oGdW6FWwPrEJTyBx0IVfeFG4PQ4Ar3if0tYuKPOaiYuwtyprdV.W')
 ON CONFLICT ("user_name") DO NOTHING;
