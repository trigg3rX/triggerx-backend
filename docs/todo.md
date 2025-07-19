# TODO List for mainnet-ready TriggerX (on Holesky)

## 1. Schedulers

- [ ] Time based Tasks: time drift for Dynamic Job Scheduling (more details here)
- [ ] Benchmarking and Performance Optimization
- [ ] Logging
- [ ] Metrics and Monitoring

## 2. Task Manager (Redis)

- [ ] Streams management and optimization
- [ ] ~~Dynamic Keeper Selection~~
- [ ] Logging
- [ ] Metrics and Monitoring

## 3. Keepers

- [ ] Benchmarking and Performance Optimization
- [ ] Logging
- [ ] Metrics and Monitoring

## 4. DB Server, Registrar, Health

- [ ] Logging
- [ ] Metrics and Monitoring

### Time Based Tasks

- Current implementation does not account for the docker execution time in the the calculation of the `next_execution_timestamp`.
- We poll the tasks every 30 seconds, and the tasks are executed in batches of 15.
- An example: Current template job has a script with total docker execution time of 40 seconds. This makes all our validations for that job to fail (current execution time and before expiration time checks).
- We need to account for the docker execution time in the the calculation of the `next_execution_timestamp`, and for that we need to pass the docker execution time to the `next_execution_timestamp` calculation. This needs to be saved in the DB.
- Similar delay would happen for rest 2 types of jobs too, maybe put a warning when task fee is calculated that the tasks will be delayed by the docker execution time.

#### Future Improvements

- Same docker for all the tasks in a keeper (execution docker can be DinD or a separate docker, or a central service we run which keepers can call and use when executing dynamic tasks).
- 
