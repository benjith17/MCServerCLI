mkdir build

for file in $(find .github/scripts -name "build-*.sh" -maxdepth 1 -type f); do
  sh "$file"
done