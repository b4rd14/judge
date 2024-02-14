s = input()
open = []
error = False

for i in range(len(s)):
    if s[i] == 'F' or s[i] == 'I':
        open.append(s[i])
    elif s[i] == 'W' or s[i] == 'E':
        if len(open) == 0:
            error = True
            break
        else:
            top = open[-1]
            open.pop()
            if (s[i] == 'W' and top != 'F') or (s[i] == 'E' and top != 'I'):
                error = True
                break

if len(open) > 0:
    error = True

if error:
    print("syntax error")
else:
    print("compile shod")