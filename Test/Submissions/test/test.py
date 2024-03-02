import sys

sys.setrecursionlimit(10000)

def custom_function(i):
    if i == 0:
        result = 5
    elif i % 2 == 0:
        result = custom_function(i - 1) - 21
    else:
        result = custom_function(i - 1) ** 2

    if i == 10000:
        print(result)

    return result

custom_function(int(input()))
