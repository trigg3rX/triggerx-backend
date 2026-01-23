-- Switch to keyspace
USE triggerx;

-- API key
INSERT INTO apikeys 
    (key, created_at, is_active, owner, rate_limit, success_count, failed_count, last_used) 
    VALUES 
    ('ADMIN', '2025-08-01 00:00:00', true, 'admin', 1000, 0, 0, '2025-08-01 00:00:00');

-- Keepers
INSERT INTO keeper_data 
    (keeper_address, keeper_name, rewards_address, operator_id, 
    registered, registered_at, network, rewards_booster) 
    VALUES 
    ('0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e', 'zoro', '0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e', 3, 
    true, ['2025-08-01 00:00:00'], 'sepolia', 1.2);

INSERT INTO keeper_data 
    (keeper_address, keeper_name, rewards_address, operator_id, 
    registered, registered_at, network, rewards_booster) 
    VALUES 
    ('0x011fcbae5f306cd793456ab7d4c0cc86756c693d', 'luffy', '0x011fcbae5f306cd793456ab7d4c0cc86756c693d', 2, 
    true, ['2025-08-01 00:00:00'], 'sepolia', 1.6);
