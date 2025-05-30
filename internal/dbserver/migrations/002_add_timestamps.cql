-- Add timestamp fields to job_data table
ALTER TABLE triggerx.job_data ADD created_at timestamp;
ALTER TABLE triggerx.job_data ADD updated_at timestamp;
ALTER TABLE triggerx.job_data ADD last_executed_at timestamp;
ALTER TABLE triggerx.job_data ADD timezone text;

-- Add timestamp fields to time_job_data table
ALTER TABLE triggerx.time_job_data ADD created_at timestamp;
ALTER TABLE triggerx.time_job_data ADD updated_at timestamp;
ALTER TABLE triggerx.time_job_data ADD last_executed_at timestamp;
ALTER TABLE triggerx.time_job_data ADD timezone text;

-- Add timestamp fields to event_job_data table
ALTER TABLE triggerx.event_job_data ADD created_at timestamp;
ALTER TABLE triggerx.event_job_data ADD updated_at timestamp;
ALTER TABLE triggerx.event_job_data ADD last_executed_at timestamp;
ALTER TABLE triggerx.event_job_data ADD timezone text;

-- Add timestamp fields to condition_job_data table
ALTER TABLE triggerx.condition_job_data ADD created_at timestamp;
ALTER TABLE triggerx.condition_job_data ADD updated_at timestamp;
ALTER TABLE triggerx.condition_job_data ADD last_executed_at timestamp;
ALTER TABLE triggerx.condition_job_data ADD timezone text;

-- Indexes for time-based queries
CREATE INDEX ON triggerx.job_data (created_at);
CREATE INDEX ON triggerx.job_data (updated_at);
CREATE INDEX ON triggerx.job_data (last_executed_at);
CREATE INDEX ON triggerx.job_data (timezone); 