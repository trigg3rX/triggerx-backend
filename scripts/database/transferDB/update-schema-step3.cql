-- Switch to keyspace
USE triggerx;

-- Drop the temporary new tables after data transfer is complete
DROP TABLE IF EXISTS user_data_new;
DROP TABLE IF EXISTS job_data_new;
DROP TABLE IF EXISTS task_data_new;
DROP TABLE IF EXISTS keeper_data_new; 