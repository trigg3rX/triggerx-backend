-- Switch to keyspace
USE triggerx;

-- API key
INSERT INTO apikeys 
    (key, created_at, is_active, owner, rate_limit) 
    VALUES 
    ('ADMIN', '2025-08-01 00:00:00', true, 'admin', 1000);

-- Keepers
INSERT INTO keeper_data 
    (keeper_id, keeper_address, keeper_name, operator_id, 
    rewards_booster, keeper_points, registered, on_imua) 
    VALUES 
    (100, '0xbae997efc6e496d8a34181c2443d97b491a3f057', 'aggregator', 1, 
    0.0, 0.0, true, false);

INSERT INTO keeper_data 
    (keeper_id, keeper_address, keeper_name, operator_id, 
    rewards_booster, keeper_points, registered, whitelisted, on_imua) 
    VALUES 
    (101, '0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e', 'zoro', 3, 
    1.5, 0.0, true, true, false);

INSERT INTO keeper_data 
    (keeper_id, keeper_address, keeper_name, operator_id, 
    rewards_booster, keeper_points, whitelisted, registered, on_imua) 
    VALUES 
    (102, '0x011fcbae5f306cd793456ab7d4c0cc86756c693d', 'luffy', 2,
    1.0, 0, true, true, false);
