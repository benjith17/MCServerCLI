mkdir build

for file in $(find .github/scripts -name "build-*.sh" -maxdepth 1 -type f); do
  echo "Running $file..."
  sh "$file"
  echo "Finished $file."
  echo; echo
done