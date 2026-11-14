import os
import subprocess
import sys

LOX_BINARY = sys.argv[1].split() if len(sys.argv) > 1 else ["plox"]

currentdir = os.path.dirname(os.path.abspath(__file__))

total_failed = 0
tests = list(filter(lambda f: f.endswith(".lox"), sorted(os.listdir(currentdir))))

if len(sys.argv) > 2:
    wanted = sys.argv[2] if sys.argv[2].endswith(".lox") else sys.argv[2] + ".lox"
    if wanted not in tests:
        print(f"No existe el archivo {wanted} en {currentdir}")
        sys.exit(2)
    tests = [wanted]

for lox_file in tests:
    result = subprocess.run(
        [*LOX_BINARY, os.path.join(currentdir, lox_file)],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        stdin=subprocess.DEVNULL,
    )

    out = result.stdout.decode().strip()
    err = result.stderr.decode().strip()
    failed = result.returncode != 0 or "error" in out.lower()

    if not failed:
        print(f"{lox_file} -> OK")
        print()
    else:
        total_failed += 1
        print(f"{lox_file} -> ERROR")
        if out:
            print(f"\tgot: <{out}>")
        err_lines = [line for line in err.splitlines() if not line.startswith("exit status")]
        if err_lines:
            print(f"\tgot: <{chr(10).join(err_lines)}>")
        print()

print()
if total_failed > 0:
    pct_failed = total_failed / len(tests) * 100
    print(" -------------------------- ")
    print(f"| Algo falló: {pct_failed:.0f}% de fallo |")
    print(" -------------------------- ")
    sys.exit(1)
else:
    print(" ---------------- ")
    print("| Todo OK: 100% |")
    print(" ---------------- ")