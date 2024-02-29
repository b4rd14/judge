import sys
sys.setrecursionlimit(10000)

def custom_function(i):
    if i == 0:
        return 5
    elif i % 2 == 0:
        return custom_function(i - 1) - 21
    else:
        return custom_function(i - 1) ** 2

user_input = int(input())
result = custom_function(user_input)
print(result)

