def fizzbuzz(n):
    res = ""
    i = 1
    while i <= n:
        if i % 3 == 0 and i % 5 == 0:
            res = res + "FizzBuzz"
        elif i % 3 == 0:
            res = res + "Fizz"
        elif i % 5 == 0:
            res = res + "Buzz"
        i = i + 1
    return res


fizzbuzz(100000)
print("done")