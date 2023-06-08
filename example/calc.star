
fibonacci = 123
x = fibonacci * 2
print("Z", x)

load("fibonacci.star", "fibonacci")
# load("fibonacci.star", "fibonacci")
load("fibonacci.star", fib="fibonacci")

f = fibonacci(10)
print("A", f[-1])
print("B", fib(10)[-1])
