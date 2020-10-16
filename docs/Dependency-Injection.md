# Dependency Injection

Dependency injection allows us decouple our code's logic from external dependencies like network calls or filesystem I/O. By passing in the functions that make these external calls, we're able to swap the functions out for mock versions of themselves in our tests. This allows us to create a consistent context for our tests, ensuring that our code's logic is the only thing being tested. 

Using dependency injection is especially useful when writing unit tests for your task's `Execute()` function. 

## Real world example

We have a simple task called `BaseCollectorConnectUS` that performs an HTTP request to the US collector endpoint. The task returns a `Success` result status if the response code from the collector is 200.

In this task, we have a helper function that reaches out to a New Relic collector endpoint and returns an http response and error.

We _could_ hardcode the method we want to use for making HTTP calls in our function. In this case, every time our function is called, it will always use the same HTTP request function (`httpHelper.MakeHTTPRequest()`) under the hood.

This might look something like this:


```go
func makeCollectorRequest(collectorUrl string) (*http.Response, error) {
    wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            collectorUrl,
		TimeoutSeconds: 30,
	}

    //Hardcoded method used here
    return httpHelper.MakeHTTPRequest(wrapper)
}
```

Since `httpHelper.MakeHTTPRequest` is hardcoded, each time our unit tests run, they will make an actual HTTP call.


### Why is this a problem?

By making an actual network call in our test, we wind up testing our function's logic **as well as** the ability to make network connections. 
We really want to isolate and test only the code/logic we're implementing: how does our task handle particular response codes, errors, etc.

>**Flaky Tests:** 
>A good suite of tests should let you decide whether the code is ready to be released. If your tests have external dependencies, a failure might not indicate a problem with your code, but with other external conditions. Flaky tests erode confidence in the test suite, and confidence in the code itself.

We want to test how our code handles the following situations:

* A 200 response is received
* A 400 response is received
* A connection error occurred
* An error occurred while parsing the response body

If our code is making actual HTTP requests during unit tests, we would need to provide HTTP endpoints that consistently respond with each expected behavior.

With dependency injection, we can provide "mock" HTTP request functions (dependencies) to our `makeCollectorRequest()` that return a particular response every time for those situations (200 response, 400 response, etc):

* In **our tests**, we provide a mock function: `mock200response()`, `mock400response()`, etc
* In **Production**, we provide the real deal: `httpHelper.MakeHTTPRequest()`

## Refactor to use dependency injection

Let's refactor `makeCollectorRequest()`. To do so, we need to:

1. Add a field to the task struct for each of our external depencies. 
2. Attach our `makeCollectorRequest()` function to the task struct as a struct method (so it can reference its dependencies defined as struct fields)
3. Define our "in-production" dependency when registering our task with the NR Diag task runner, i.e. creating the task struct instance.

### (1) allow dependencies to be attached to the task struct as struct fields

First, let's define a variable type that matches that of our actual (production) HTTP request function dependancy (`httpHelper.MakeHTTPRequest()`). This way, we can define task struct fields of this type. Our `makeCollectorRequest()` function will then reference those fields when making requests.

`httpHelper.MakeHTTPRequest`'s method signature (parameters and return types) looks like this:

```go
func MakeHTTPRequest(wrapper RequestWrapper) (*http.Response, error) {
    ...
}
```

So our new `httpRequestFunc` type should be a `func` with the same signature:

```go
type requestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)
```


Let's add a field for our HTTP request function dependency on our task struct. Our task struct is empty, and currently looks like this:

```go
type BaseCollectorConnectUS struct {

}
```

Let's add a field for our HTTP request function dependency:

```go
type BaseCollectorConnectUS struct {
    httpGetter httpRequestFunc
}
```

Here we have a struct field name `httpGetter` with the type: `httpRequestFunc`. From within any methods that have this struct as a receiver, `p`, we can access this function by calling `p.httpGetter`. 

### (2) Attach our `makeCollectorRequest()` to the task struct as a struct method (so it can reference its dependencies defined as struct fields)

Next, we attach our `makeCollectorRequest` function to our task `struct` so it can reference dependencies that are fields on the struct instance. We do this by adding the struct as a receiver (`(p BaseCollectorConnect)`) in the function declaration:

```go
func (p BaseCollectorConnectUS) makeCollectorRequest(collectorUrl string) (*http.Response, error) {
    ...
}
```

Finally, we modify the function code to use the dependency specified on the struct (`httpGetter`), instead of the hardcoded function call to `httpHelper.MakeHTTPRequest`:

```go
func (p BaseCollectorConnectUS) makeCollectorRequest(collectorUrl string) (*http.Response, error) {
    wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            collectorUrl,
		TimeoutSeconds: 30,
	}

    //Injected dependency method used here
    return p.httpGetter(wrapper)
}
```

### (3) Define our "in-production" dependency when registering our task (i.e. creating the task struct instance).

For **production**, you'll want to define the dependency when you register your task:

```go
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
    registrationFunc(BaseCollectorConnectUS{
        httpGetter: httpHelper.MakeHTTPRequest
    }, true)
}
```

In your unit tests, you can swap this dependency out for a dumb mock function that returns the same result every time, regardless of input.
You can then specify this mock function when your initilizing your test task struct.  

These tests will be using the Ginkgo library. See [unit-testing.md](unit-testing.md) for more information.

```go
 var _ = Describe("Base/Collector/ConnectUS", func() {
	var p BaseCollectorConnectUS
	Describe("Execute ", func() {

		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When the connection is successful", func() {

			BeforeEach(func() {
				p = BaseCollectorConnectUS{
					httpGetter: mockSuccessfulRequest200,
				}
				upstream = map[string]tasks.Result{
					"Base/Config/RegionDetect": tasks.Result{
						Status: tasks.Success,
						Payload: []string{
							"us01",
						},
					},
				}
			})

			It("Should return a Result instance with a status of Success", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("Should return a Result instance with the expected summary for Explain", func() {
				Expect(result.Summary).To(Equal("Status Code = 200 Body = test body"))
			})
		})
	})
})

```

The `mockSuccessfulRequest200` function specified in the test is a dummy function with the same method signature as the actual dependency, matching the `requestFunc` type.

```go
func mockSuccessfulRequest200(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}
```

This pattern can be applied to replace external dependencies in any helper functions. Thanks for reading!
