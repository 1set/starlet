load("fibonacci.star", "fibonacci", fib = "fibonacci")

fibonacci = 123
x = fibonacci * 2
print("Z", x)

f = fibonacci(10)
print("A", f[-1])
print("B", fib(10)[-1])
