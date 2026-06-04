mkdir build

for file in $(find scripts -name "build-*.sh" -maxdepth 1 -type f); do
  sh "$file"
done