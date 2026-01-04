## DevTest environment
This repo's test is ran from VSCode test runner
1. In order to run individual test from VSCode inside the test file directly, you need to put the .env.test within the /integration_test folder
- It's unclear how to point the env file to the root of the project in a clean manner so this is a compromise for now

2. In order to run and have an overview of all the test case from the Testing tab in VSCode you need to have a .env.test at the root level

These are the potential areas to be reconfigured when we have more resources

### Environment Setup
Run `docker compose up` to bring up the database