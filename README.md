# triggerx-backend

Backend Utilities and APIs for TriggerX

## Dependencies

* Node v20
* Docker (If running a keeper Node)

## Commands

1. Environment variables:
   Create a .env file from example given and change the values.

2. Install packages:

    ```bash
    yarn
    ```

3. Open terminal #1 and run the command:

   ```bash
   make start-taskmanager
   ```

4. Open terminal #2 and run:

    ```bash
    make start-aggregator
    ```

5. Open terminal #3, and setup keeper using:

    ```bash
    make start-keeper
    ```

    Alternatively, you can run the keeper in Docker:

    ```bash
    make start-docker
    ```
