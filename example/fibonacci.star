def fibonacci(n):
    res = list(range(n))
    for i in res[2:]:
        res[i] = res[i - 2] + res[i - 1]
    return res

# print("tf", fibonacci(100))
