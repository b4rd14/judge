
def ELMOS(n):
    if n == 0:
        return 5
    if n % 2 == 0:
        return ELMOS(n - 1) - 21
    else:
        return ELMOS(n - 1) ** 2

value = int(input())
result = ELMOS(value)
print(result)
