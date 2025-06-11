# TODO List for the Keeper Backend Production Scaling

## Redis Service

* Question to PS: If streams are handling all the caching with Upstash / local redis, then why was `internal/cache` package created? How is it used?

* [x] `NewRedisClient`: Create a new redis client, uses Upstash if available, otherwise uses local redis, panics if both are not available.
* [x] `JobStream`: 2 kinds of streams, running and completed. TTL of 120 hrs. COntains JobID, TaskDefinitionID, TaskIDs.
* [x] `TaskStream`: 5 kinds of streams with 1 minute TTL for each task:
  * [x] Ready: Queue of tasks that are ready to be executed, i.e., to be sent to performer. There will be not tasks here until the network is busy will no performers available.
  * [x] Processing: Tasks sent to performer. Waiting for Performer ack.
  * [x] Retry: Tasks that failed to be executed. 3 retries.
  * [ ] Failed: Tasks that failed to be executed 3 times.
    * TODO: What can we do here?
  * [x] Completed: Tasks that were successfully executed.
* [ ] `PerformerLocks`: Health service `/operators` endpoint would have the current active performers. Implement round robin for them. Lock the performer when sending tasks to them. Release them if sending tasks fails or if the task is completed ack is received.
  * We need to keep the next performer ready as soon as the current performer is locked.
  * **TODO**:
    * ~~Round robin logic~~ (skip for now, we have fixed performers)
    * ~~Fetch the list of performers from health service~~ (skip for now, we have fixed performers)
    * `AcquirePerformerLock` and `ReleasePerformerLock`: Defined, but we need to call them when we send tasks to performers from `redis` package.

### API Server

I was thinking of implementing the routes in an API server, which shall be accessed by Othentic Attester node, to get the updates regarding task execution requests and validation requests. It can be used to get the status of the tasks, and to get the results of the tasks, along with keeper's status: busy or idle.

* [ ] `/p2p/meassage`: receives the execution data
* [ ] `/task/validate`: receives the validation data

**TODO:** Basic handlers deifned, need to add the data parsing logic, along with the logic to update the redis streams.

## Schedulers

### Time based Tasks

No workers, only one scheduler, which polls every 30 seconds and sends the tasks to the performer with `next_execution_timestamp`. Peroformer will wait till `next_execution_timestamp - time_drift` to execute the task. The time drift will be 0 initially, and will be adjusted as we get the results in testing.

* [x] `pollAndScheduleJobs`: It polls every 30 seconds for jobs with `next_execution_timestamp` < 40 seconds. that 10 second is a security margin.
* [x] `processBatch`: The jobs are processed in batches of 15. This is to ensure resource usage is not too high, and can be changed as per benchmarking. Only next 30 seconds of jobs are processed. In case the next poll fails, we use the 10 secs oj jobs for processing, will retrying to poll again. (Current assumption is 1 task per 2 sec, so 15 tasks per 30 seconds and hence the batch size is 15)
* [x] `executeJob`: update redis service regarding these tasks are to be executed for these jobs. Update the JobStream and TaskStream accordingly. Also, get performerLock for this batch of jobs.
* [x] `performJobExecution`: Send the list of tasks to the performer. It will be busy for next 30 seconds, executing the tasks at exact time possible. Handle the nonce and gas fees for this.
* [ ] `check.go`: Will receive the list of tasks from redis that were successfully executed from api server endpoints. Update the JobStream and TaskStream accordingly.
  * **TODO**: It is a blank file.

### Event and Condition based Tasks

Pool of workers, each monitoring the "condition". When it happens, it notifies the redis to get the keeper, and execute the task.

* [x] `api/server.go`: API server with `/jobs/schedule` and `/jobs/delete` endpoints. DB server calls these endpoints to schedule and delete jobs as it gets the requests from frontend / sdk.
* [x] worker = go routine. It will be monitoring the condition. When it happens, it will fetch keeper from redis, lock it and execute the task.
* [ ] `check.go`: Will receive the list of tasks from redis that were successfully executed from api server endpoints. Update the JobStream and TaskStream accordingly.
  * **TODO**: It is a blank file.

### Task Execution

* [x] `checkIfPerformer`: Check is the performer is self. If not, it will not perform the task.
  * [x] `executeTask`:
    * [x] `validateSchedulerSignature`: Validate the scheduler signature.
    * [x] `validateTrigger`: Validate the trigger.
    * [x] `takeActionWithStaticArgs`: Take action with static arguments.
    * [x] `takeActionWithDynamicArgs`: Take action with dynamic arguments.
  * [x] `generateProof`: TLS certificate proof generation.
  * [x] `signIPFSData`: Sign the ipfs data using Consensus private key.
  * [x] `uploadToIPFS`: Upload the data + proof to IPFS.

### Task Validation

* [x] `validateTask`: Validate the task.
  * [x] `validateSchedulerSignature`: Validate the scheduler signature.
  * [x] `validateTrigger`: Validate the trigger.
  * [x] `validateAction`: Validate the action.
  * [x] `validateProof`: Validate the proof.
  * [x] `validateSignature`: Validate the performer signature.

More details in [Keeper](keeper.md).
