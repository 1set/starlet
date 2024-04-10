#import sys

print("cnt:", len(sys.argv))
print("argv:", sys.argv)
print("platform:", sys.platform)
print("cwd:", path.getcwd())
print("path:", runtime.getenv("PATH"))