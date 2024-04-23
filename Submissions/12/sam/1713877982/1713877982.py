
def ELMOS(n):
    if n == 0:
        return 5
    elif n % 2 == 0:
        return ELMOS(n-1) - 21
    else:
        return ELMOS(n-1) ** 2

# Get input from user
n = int(input())

# Call the function and print the result
print(ELMOS(n))
