import time

m = list(map(int, input().split()))
for num in m:
    print(num , end=" ")
    time.sleep(num)