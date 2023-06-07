coins = {
    "dime": 10,
    "nickel": 5,
    "penny": 1,
    "quarter": 25,
}

print("By name:\t" + ", ".join(sorted(coins.keys())))
print("By value:\t" + ", ".join(sorted(coins.keys(), key = coins.get)))
