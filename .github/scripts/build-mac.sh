# Build macOS (intel)
mkdir ./build/mcsm-osx-intel.app/Contents/MacOS -p

cat > ./build/mcsm-osx-intel.app/Contents/Info.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>MCServerManagerCLI</string>
    <key>CFBundleDisplayName</key>
    <string>MCServerManagerCLI</string>
    <key>CFBundleIdentifier</key>
    <string>uk.benjith.mcsm</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleExecutable</key>
    <string>mcsm-osx-intel</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
</dict>
</plist>
EOF

go env -w GOOS=darwin GOARCH=amd64
go build -o ./build/mcsm-osx-intel.app/Contents/MacOS/mcsm-osx-intel .

chmod +x ./build/mcsm-osx-intel.app/Contents/MacOS/mcsm-osx-intel
cp assets/icon.icns ./build/mcsm-osx-intel.app/Contents/Resources/icon.icns

zip -r ./build/mcsm-osx-intel.zip ./build/mcsm-osx-intel.app



# Build macOS (arm64)
mkdir ./build/mcsm-osx-arm.app/Contents/MacOS -p

cat > ./build/mcsm-osx-arm.app/Contents/Info.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>MCServerManagerCLI</string>
    <key>CFBundleDisplayName</key>
    <string>MCServerManagerCLI</string>
    <key>CFBundleIdentifier</key>
    <string>uk.benjith.mcsm</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleExecutable</key>
    <string>mcsm-osx-arm</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
</dict>
</plist>
EOF

go env -w GOOS=darwin GOARCH=arm64
go build -o ./build/mcsm-osx-arm.app/Contents/MacOS/mcsm-osx-arm .

chmod +x ./build/mcsm-osx-arm.app/Contents/MacOS/mcsm-osx-arm
cp assets/icon.icns ./build/mcsm-osx-arm.app/Contents/Resources/icon.icns

zip -r ./build/mcsm-osx-arm.zip ./build/mcsm-osx-arm.app