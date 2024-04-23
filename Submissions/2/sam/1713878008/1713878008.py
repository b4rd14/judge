def ELMOS(num):
    if num == 0:
        return 5
    if num % 2 == 0:
        return ELMOS(num - 1) - 21
    else:
        return ELMOS(num - 1) ** 2

n = int(input())
result = ELMOS(n)
print(result)