# Testing

NR Diag utilizes two types of testing: [Unit testing](Unit-Testing.md) and [Integration testing](Integration-Testing.md).

We recommend **unit testing** to make up a majority of your code's test coverage because unit tests:
 * Are quicker to run than integration tests
 * Are more reliable
 * Result in more readable, testable, and maintainable code
 * Faciliate Test-driven Development

 If you believe you are unable to use unit tests to adequately test your code, please add your reasoning for not using unit tests.
 
Here is our documentation on both unit and integration testing your NR Diag contributions:

* [**Unit testing**](Unit-Testing.md)
* [**Integration testing**](Integration-Testing.md)

It's also important to run your task a few times in a realistic environment. For instance, try setting up a virtual machine or vagrant box, installing a full copy of everything you're testing and then running your task. This will help expose any mistaken assumptions that your tests might not have uncovered(since they were assuming the same things as your code). Performing a smoke test like this before releasing has helped the team find several bugs that we missed when running in more controlled environments.
