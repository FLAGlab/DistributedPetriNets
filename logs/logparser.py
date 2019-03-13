import os
import glob

for filepath in glob.iglob('./*.log'):
    print(filepath)
    with open(filepath, 'r') as f:
        with open(filepath + ".csv", 'w') as out:
            for line in f:
                if line.startswith("_"):
                    out.write(line)
