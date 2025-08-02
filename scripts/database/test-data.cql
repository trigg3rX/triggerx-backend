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
    (998, '0xbae997efc6e496d8a34181c2443d97b491a3f057', 'aggregator', 120, 
    0.0, 0.0, true, false);

INSERT INTO keeper_data 
    (keeper_id, keeper_address, keeper_name, operator_id, 
    rewards_booster, keeper_points, registered, on_imua) 
    VALUES 
    (989, '0xebe4f49161eb9002ed19e06f943dce73d24b596e', 'nexus', 121, 
    0.0, 0.0, true, false);

INSERT INTO keeper_data 
    (keeper_id, keeper_address, keeper_name, operator_id, 
    rewards_booster, keeper_points, whitelisted, registered, on_imua) 
    VALUES 
    (899, '0x011fcbae5f306cd793456ab7d4c0cc86756c693d', 'luffy', 124,
    1.6, 0, true, true, false);
