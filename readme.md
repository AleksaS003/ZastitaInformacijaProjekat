# Crypto App - CLI Tool

A command-line tool for cryptographic operations implementing:
- Foursquare cipher (classical)
- LEA (Lightweight Encryption Algorithm)
- PCBC (Propagating Cipher Block Chaining)
- SHA-256 (hash function)

## Installation


# Build CLI (prvo)
make build

# Build GUI
go build -o crypto-gui ./cmd/gui

# Pokreni GUI
./crypto-gui


primer za foursquare
# Testiraj foursquare encrypt sa tekstom
./crypto-cli foursquare encrypt --text="HELLO WORLD" --key1=keyword --key2=example

# Testiraj sa fajlom
echo "SECRET MESSAGE" > test.txt
./crypto-cli foursquare encrypt --file=test.txt --key1=mykey --key2=yourkey

# Proveri round-trip
./crypto-cli foursquare encrypt --text="TESTING" --key1=a --key2=b > encrypted.txt
./crypto-cli foursquare decrypt --file=encrypted.txt --key1=a --key2=b


lea

# 2. Generiši LEA ključ
./crypto-cli lea genkey-file --size=128 --output=key.bin

# 3. Proveri veličinu
wc -c key.bin  # Mora biti 16

# 4. Napravi test fajl
echo "This is a test message for LEA encryption." > test.txt

# 5. Šifruj
./crypto-cli lea encrypt --file=test.txt --keyfile=key.bin --output=encrypted.bin

# 6. Dešifruj
./crypto-cli lea decrypt --file=encrypted.bin --keyfile=key.bin --output=decrypted.txt

# 7. Uporedi
diff test.txt decrypted.txt
echo "Exit code: $?"  # Treba da bude 0

# 8. Testiraj sa hex ključem direktno
./crypto-cli lea genkey 128
# Kopiraj hex string (bez "Hex:" prefiksa)
# Npr: 7a5c3f2e1d0a9b8c7d6e5f4a3b2c1d0e
./crypto-cli lea encrypt --file=test.txt --key=7a5c3f2e1d0a9b8c7d6e5f4a3b2c1d0e --output=encrypted2.bin


PCBC

# 1. Generiši ključ
./crypto-cli lea genkey-file --size=128 --output=pcbc-key.bin

# 2. Napravi test fajl
cat > pcbc-test.txt << 'EOF'
PCBC (Propagating Cipher Block Chaining) test.
This mode provides better error propagation than CBC.
Each block depends on two previous blocks.
EOF

# 3. Šifruj sa PCBC
echo "Encrypting with PCBC mode..."
./crypto-cli pcbc encrypt --file=pcbc-test.txt --keyfile=pcbc-key.bin --output=pcbc-encrypted.bin

# 4. Dešifruj
echo "Decrypting with PCBC mode..."
./crypto-cli pcbc decrypt --file=pcbc-encrypted.bin --keyfile=pcbc-key.bin --output=pcbc-decrypted.txt

# 5. Uporedi
if cmp -s pcbc-test.txt pcbc-decrypted.txt; then
    echo " PCBC round-trip PASSED"
else
    echo " PCBC round-trip FAILED"
    echo "Differences:"
    diff -u pcbc-test.txt pcbc-decrypted.txt
fi

# 6. Test error propagation
echo -e "\nTesting PCBC error propagation..."
# Kopiraj šifrovani fajl
cp pcbc-encrypted.bin pcbc-corrupted.bin
# Izmeni jedan bajt negde u sredini
python3 -c "
data = open('pcbc-corrupted.bin', 'rb').read()
# Izmeni bajt na poziciji 50
data = bytearray(data)
data[50] ^= 0x01
open('pcbc-corrupted.bin', 'wb').write(data)
print('Corrupted one byte at position 50')
"

# Pokušaj dekripciju korumpiranog fajla
./crypto-cli pcbc decrypt --file=pcbc-corrupted.bin --keyfile=pcbc-key.bin --output=pcbc-corrupted-decrypted.txt 2>&1

echo "PCBC error propagation characteristic:"
echo "- In PCBC, error in block n corrupts blocks n and n+1"
echo "- But recovers at block n+2 (unlike CBC)"




echo "=== Final SHA-256 Test ==="

# Test praznog stringa
echo "1. Testing empty string:"
./crypto-cli sha256 hash --text=""
echo "Expected: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

# Test know vectors
echo -e "\n2. Testing known vectors:"
echo "'abc':"
./crypto-cli sha256 hash --text="abc"

echo -e "\n'hello':"
./crypto-cli sha256 hash --text="hello"

# File hash round-trip
echo -e "\n3. File hash round-trip:"
cat > test-file.txt << 'EOF'
This is a test file.
Multiple lines.
SHA-256 should work consistently.
EOF

echo "Original file created"
./crypto-cli sha256 hash --file=test-file.txt --output=file-hash.txt

echo "Verifying..."
./crypto-cli sha256 verify --file=test-file.txt --hashfile=file-hash.txt

# Modify and verify fails
echo -e "\n4. Testing corruption detection:"
echo "MODIFIED" >> test-file.txt
./crypto-cli sha256 verify --file=test-file.txt --hashfile=file-hash.txt





echo "=== Testing Build ==="
./crypto-cli help

# Test metadata komande
echo -e "\n=== Testing Metadata Commands ==="
cat > test-doc.txt << 'EOF'
Test document for metadata system.
This should be encrypted with metadata header.
EOF

./crypto-cli lea genkey-file --size=256 --output=test-meta.key

echo "Testing encrypt-file command..."
./crypto-cli encrypt-file --file=test-doc.txt --keyfile=test-meta.key --algo=LEA-PCBC

echo -e "\nTesting decrypt-file command..."
./crypto-cli decrypt-file --file=test-doc.txt.enc --keyfile=test-meta.key

# Proveri da li su fajlovi isti
if cmp -s test-doc.txt test-doc.txt.enc.dec; then
    echo " Metadata system working!"
else
    echo " Metadata system failed"
fi




# Kompletan test
cat > test-metadata-full.sh << 'EOF'
#!/bin/bash

echo "=== COMPREHENSIVE METADATA SYSTEM TEST ==="

# Cleanup
rm -f *.txt *.enc *.dec *.key *.bin

# 1. Build
echo "1. Building..."
make build

# 2. Generate key
echo -e "\n2. Generating key..."
./crypto-cli lea genkey-file --size=256 --output=test.key

# 3. Create test files
echo -e "\n3. Creating test files..."
cat > doc1.txt << 'DOC1'
Confidential Report
===================
Date: 2024-01-17
Author: John Doe
Content: This is sensitive information.
DOC1

cat > doc2.txt << 'DOC2'
Financial Data
==============
Quarter: Q4 2023
Revenue: $1,234,567
Expenses: $987,654
DOC2

# 4. Encrypt with metadata
echo -e "\n4. Encrypting files with metadata..."
./crypto-cli encrypt-file --file=doc1.txt --keyfile=test.key --algo=LEA-PCBC --output=doc1.enc
./crypto-cli encrypt-file --file=doc2.txt --keyfile=test.key --algo=LEA --output=doc2.enc

echo "Encrypted files created:"
ls -la *.enc

# 5. Show metadata structure
echo -e "\n5. Metadata header structure:"
echo "doc1.enc (first 200 bytes):"
hexdump -C doc1.enc | head -10

# 6. Decrypt files
echo -e "\n6. Decrypting files..."
./crypto-cli decrypt-file --file=doc1.enc --keyfile=test.key --output=doc1-decrypted.txt
./crypto-cli decrypt-file --file=doc2.enc --keyfile=test.key --output=doc2-decrypted.txt

# 7. Verify
echo -e "\n7. Verifying..."
if cmp -s doc1.txt doc1-decrypted.txt && cmp -s doc2.txt doc2-decrypted.txt; then
    echo " ALL TESTS PASSED - Metadata system working correctly!"
else
    echo " TESTS FAILED"
    echo "doc1 comparison:" && cmp doc1.txt doc1-decrypted.txt
    echo "doc2 comparison:" && cmp doc2.txt doc2-decrypted.txt
fi

# 8. Test error cases
echo -e "\n8. Testing error cases..."
echo "Testing wrong key..."
./crypto-cli lea genkey-file --size=256 --output=wrong.key
./crypto-cli decrypt-file --file=doc1.enc --keyfile=wrong.key --output=should-fail.txt 2>&1 | grep -q "failed" && echo " Correctly failed with wrong key"

echo -e "\n Metadata System Test Complete!"
EOF

chmod +x test-metadata-full.sh
./test-metadata-full.sh



echo "=== Testing File System Watcher ==="

# 1. Kreiraj test direktorijume
mkdir -p test-watch test-encrypted

# 2. Generiši ključ
./crypto-cli lea genkey-file --size=128 --output=fsw-key.bin

# 3. Šifruj postojeće fajlove (ako ih ima)
echo "Test file 1" > test-watch/file1.txt
echo "Test file 2" > test-watch/file2.txt

echo "Encrypting existing files..."
./crypto-cli fsw encrypt-existing --watch=test-watch --output=test-encrypted --keyfile=fsw-key.bin

# 4. Pokreni FSW u background-u
echo -e "\nStarting FSW in background..."
./crypto-cli fsw start --watch=test-watch --output=test-encrypted --keyfile=fsw-key.bin --algo=LEA-PCBC &
FSW_PID=$!

# 5. Dodaj nove fajlove dok FSW radi
sleep 2
echo "Creating new files while FSW is running..."
echo "New file 3" > test-watch/file3.txt
echo "New file 4" > test-watch/file4.txt

# 6. Proveri da li su fajlovi šifrovani
sleep 3
echo -e "\nChecking encrypted files..."
ls -la test-encrypted/

# 7. Zaustavi FSW
kill $FSW_PID 2>/dev/null || true

# 8. Dešifruj i proveri
echo -e "\nDecrypting to verify..."
./crypto-cli decrypt-file --file=test-encrypted/file1.txt.enc --keyfile=fsw-key.bin --output=decrypted1.txt
./crypto-cli decrypt-file --file=test-encrypted/file3.txt.enc --keyfile=fsw-key.bin --output=decrypted3.txt

if cmp test-watch/file1.txt decrypted1.txt && cmp test-watch/file3.txt decrypted3.txt; then
    echo " FSW test PASSED - Auto-encryption works!"
else
    echo " FSW test FAILED"
fi

# 9. Proveri log
echo -e "\nFSW Log (last 10 lines):"
tail -10 fsw.log 2>/dev/null || echo "No log file"



# Obriši sve test direktorijume
rm -rf test-* fsw-* debug-* watch-* encrypted-* decrypted-*

# Obriši sve .txt, .enc, .dec, .bin, .key fajlove (osim crypto-cli)
rm -f *.txt *.enc *.dec *.bin *.key *.log 2>/dev/null

# Obriši sve sa prefiksima
rm -f foursquare-* lea-* pcbc-* sha256-* document-* secret-* largefile-* hex-* out-* simple-* tajna-* sifrovano-*




# 1. Terminal 1 - Server
./crypto-cli lea genkey-file --size=128 --output=test.key
echo "Test" > test.txt

./crypto-cli server --address=:1234 --output=out --keyfile=test.key

# 2. Terminal 2 - Klijent  
./crypto-cli client --address=localhost:1234 --file=test.txt --keyfile=test.key --algo=LEA-PCBC



./crypto-cli server --address=0.0.0.0:5555 --output=out --keyfile=test.key

./crypto-cli client --address=192.168.1.10:5555 --file=test.txt --keyfile=test.key --algo=LEA-PCBC./crypto-cli client --address=192.168.1.10:5555 --file=test.txt --keyfile=test.key --algo=LEA-PCBC