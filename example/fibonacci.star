def fibonacci(n):
    res = list(range(n + 1))
    for i in res[2:]:
        res[i] = res[i - 2] + res[i - 1]
    return res[1:]

def fib_last(n):
    return fibonacci(n)[-1]

# print("tf", fibonacci(100))
