# Assert
[![Build Status](https://travis-ci.org/AMekss/assert.svg?branch=master)](https://travis-ci.org/AMekss/assert)
[![Maintainability](https://api.codeclimate.com/v1/badges/1fc2f9f7b3058063795d/maintainability)](https://codeclimate.com/github/AMekss/assert/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/1fc2f9f7b3058063795d/test_coverage)](https://codeclimate.com/github/AMekss/assert/test_coverage)

Go testing micro library. This library is small yet opinionated as it aims to be as simple as possible while provide testing utilities for 80% of the cases in order to remove boilerplate from the testing code, improve test expressiveness and standardize test failure messages. All the edge cases (the rest 20%) are not covered and have to be tested the way you would do without this library. Before you decide on using it please read and consider [this](https://golang.org/doc/faq#testing_framework). On the other hand if you're looking for a fully flagged all in one testing solution, this is not the place, however [testify](https://github.com/stretchr/testify) might be it.

## Usage example
If you are still here cool! Bare with me and I'll show what this library can bring you :)

Let's assume we have some `db` method receiver implementing `FetchRecord` method which in turn accepts `id` and returns some record along with the error if there is any. Now we'd like to test this behavior. So by using `assert` library the test would look something like this:

```go
func TestRecordFetching(t *testing.T){
    record, err := db.FetchRecord(123)

    assert.NoError(t.Fatalf, err)
    assert.EqualStrings(t, "Bart", record.FirstName)
    assert.EqualStrings(t, "Simpson", record.LastName)
}
```

Which is effectively the same as following test by using only inbuilt testing facilities:
```go
func TestRecordFetching(t *testing.T){
    record, err := db.FetchRecord(123)

    if err != nil {
        t.Fatalf("\nEquality assertion failed:\n\twant: no error \n\t got: error '%s'", err)
    }

    expectedName := "Bart"
    if record.FirstName != expectedName {
        t.Errorf("\nEquality assertion failed:\n\twant: '%s' \n\t got: '%s'", expectedName, record.FirstName)
    }

    expectedLastName := "Simpson"
    if record.LastName != expectedLastName {
        t.Errorf("\nEquality assertion failed:\n\twant: '%s' \n\t got: '%s'", expectedLastName, record.LastName)
    }
}
```
As we can see even in this pretty simple and small test above there are a lot of boilerplate and repetition. And since software developers are extremely lazy people (like I am :D) which in practice brings us to short and ambiguous failure messages in the production test suits (The best I've seen so far `t.Error("write something useful here")` :))

I hope that `assert` library will help to reduce boilerplate code and consumption of cognitive energy in making up failure messages while guarantee consistent and descriptive messages across the test suite. So you can focus on the test itself and the logic you want to test.

For more examples please check [Docs](#docs) section below

## Installation
```
$ go get github.com/amekss/assert
```

(optional) To run unit tests:
```
$ cd $GOPATH/src/github.com/amekss/assert
$ go test -cover
```

## Docs

### Equality
```go
func TestEquality(t *testing.T) {
    // asserts that to values are the same
    assert.EqualStrings(t, expectedStr, "foo")
    assert.EqualErrors(t, expectedErr, errors.New("bar"))
    assert.EqualInt(t, expectedInt, 10)
    assert.EqualFloat32(t, expectedFloat32, float32(2.5))
    assert.EqualFloat64(t, expectedFloat64, float64(2.5))
    assert.EqualTime(t, expectedTime, time.Now())
}

```
**Note:** There is no `assert.Equal` method and no current plans on implementing one. IMHO well written tests serves as live documentation and `assert.Equal(t, a, b)` doesn't reads as good as for example `assert.EqualStrings(t, a, b)` since just by looking at it its quite clear that we're testing strings here.

### Equality within a tolerance
```go
func TestEqualityTol(t *testing.T) {
    // asserts that two values are the same within a relative tolerance (f.ex. 2%)
    assert.EqualFloat32Tol(t, expectedFloat32, float32(101.0), float32(0.02))
    assert.EqualFloat64Tol(t, expectedFloat64, float64(101.0), float64(0.02))
}
```

### Boolean
```go
func TestBoolean(t *testing.T) {
    // asserts that expression is evaluated to true or false
    assert.True(t, true==true)
    assert.False(t, true!=true)
}
```

### Inclusion
```go
func TestInclusion(t *testing.T) {
    // asserts that first string is a substring of the other (substring of error message in the case of errors)
    assert.IncludesString(t, "fo", "foo")
    assert.ErrorIncludesMessage(t, "foo", err)
}
```

### Nils & Panics
```go
func TestNils(t *testing.T) {
    // asserts that error is not produced
    assert.NoError(t, err)
    // asserts panic is produced
    defer assert.Panic(t, "foo")
    // assert that nil is received
    assert.IsNil(t, nil)
}
```

### Advanced reporters
By default assert will use `Errorf()` of the `*testing.T` to report error and that's why tests are not interrupted on the first encountered failures. Which helps to get full picture and list of all failed assertions in the test report, while sometimes it's better to interrupt tests when the failure is encountered. For example when function call under test are returning some value along with error, so one might want to stop further assertions on value if the error is not `nil`, as it might cause some panics and add some unnecessary noise to the test results. For this very reason you can customize reporter function by passing in the one you need as follows:
```go
func TestSomeBehavior(t *testing.T) {
    // fail fast behavior, stop current test further assertions on failure
    assert.NoError(t.Fatalf, err)
    // default shortcut - Errorf() on t will be called under the hood current test won't be interrupted on failure
    assert.EqualStrings(t, "foo", foo)
    // same as default reporting behavior
    assert.EqualStrings(t.Errorf, "bar", bar)
```

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
