def minsky_even(n):
    r0 = n
    pc = 0
    halt = False

    while not halt:
        if pc == 0:
            if r0 == 0:
                pc = 2
            else:
                r0 = r0 - 1
                pc = 1
        elif pc == 1:
            if r0 == 0:
                pc = 3
            else:
                r0 = r0 - 1
                pc = 0
        elif pc == 2:
            r0 = r0 + 1
            pc = 3
        elif pc == 3:
            halt = True

    return r0


print(minsky_even(200000))