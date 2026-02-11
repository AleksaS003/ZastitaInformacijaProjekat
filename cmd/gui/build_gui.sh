echo " Build-ujem GUI aplikaciju..."

go get fyne.io/fyne/v2@latest
go install fyne.io/fyne/v2/cmd/fyne@latest

echo " Build za Linux..."
GOOS=linux GOARCH=amd64 go build -o bin/crypto-gui-linux ./cmd/gui

echo " Build za Windows..."
GOOS=windows GOARCH=amd64 go build -o bin/crypto-gui-windows.exe ./cmd/gui

echo " Build za macOS..."
GOOS=darwin GOARCH=amd64 go build -o bin/crypto-gui-macos ./cmd/gui


if command -v fyne &> /dev/null; then
    echo " Pakujem GUI aplikaciju..."
    fyne package -os darwin -icon icon.png -name "Crypto App" -executable ./bin/crypto-gui-macos
    fyne package -os windows -icon icon.png -name "Crypto App" -executable ./bin/crypto-gui-windows.exe
    fyne package -os linux -icon icon.png -name "Crypto App" -executable ./bin/crypto-gui-linux
fi

echo " Build zavrsen! Binarni fajlovi su u ./bin/"