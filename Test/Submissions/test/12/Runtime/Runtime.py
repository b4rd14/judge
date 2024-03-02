def func(i):
    if i == 0:
        return 5
    if i % 2 == 0:
        return func(i - 1) - 21
    else:
        return func(i - 1) ** 2

n = int(input())
output = func(n)
print(output)

