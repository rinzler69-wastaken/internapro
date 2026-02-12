-- Admin Seeder
-- Email: office@klikdsi.com
-- Password: internapro2025

INSERT INTO users (email, password_hash, role, name, is_2fa_enabled, created_at, updated_at)
VALUES 
(
    'office@klikdsi.com', 
    '$2a$10$GQdxZAxHt1TyWBJs7jgh3Of5BiXQRN7jnCFKhOnDMQhYvreK3MFW', -- Hash for 'internapro2025'
    'admin', 
    'DSI-Admin', 
    0, 
    NOW(), 
    NOW()
)
ON DUPLICATE KEY UPDATE 
    password_hash = VALUES(password_hash),
    name = VALUES(name),
    role = 'admin',
    updated_at = NOW();
