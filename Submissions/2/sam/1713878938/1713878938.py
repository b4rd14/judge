
def ELMOS(n):
    if n == 0:
        return 5
    else:
        if n % 2 == 0:
            return ELMOS(n-1) - 21
        else:
            return ELMOS(n-1) ** 2

input_num = int(input())
result = ELMOS(input_num)
print(result)
