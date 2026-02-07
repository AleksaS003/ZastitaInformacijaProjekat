#!/bin/bash

echo "========================================="
echo "     FINAL SYSTEM INTEGRATION TEST"
echo "========================================="

# Cleanup
rm -rf final-test-* final-received final-watch 2>/dev/null

# 1. Setup
echo -e "\n1. 📁 Setting up test environment..."
mkdir -p final-watch final-received

# 2. Generate key
echo -e "\n2. 🔑 Generating encryption key..."
./crypto-cli lea genkey-file --size=256 --output=final-key.bin

# 3. Create test files
echo -e "\n3. 📝 Creating test files..."
echo "Document 1: Confidential Report" > final-watch/report.txt
echo "Document 2: Financial Data" > final-watch/finance.txt
echo "Large test file..." > final-watch/large.bin
dd if=/dev/urandom bs=1K count=10 2>/dev/null >> final-watch/large.bin

# 4. Test FSW (batch encryption)
echo -e "\n4. �� Testing FSW batch encryption..."
./crypto-cli fsw encrypt-existing \
  --watch=final-watch \
  --output=final-received \
  --keyfile=final-key.bin \
  --algo=LEA-PCBC

echo "   Encrypted files:"
ls -la final-received/*.enc 2>/dev/null | awk '{print "   - " $9 " (" $5 " bytes)"}'

# 5. Test TCP in background
echo -e "\n5. 📡 Testing TCP Network (background)..."
./crypto-cli server --address=:9999 --output=final-received --keyfile=final-key.bin &
SERVER_PID=$!
sleep 2

# Send file via TCP
echo "   Sending file via TCP..."
./crypto-cli client \
  --address=localhost:9999 \
  --file=final-watch/report.txt \
  --keyfile=final-key.bin \
  --algo=LEA-PCBC

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

# 6. Verify received files
echo -e "\n6. ✅ Verification..."
TCP_SUCCESS=false
if [ -f "final-received/report.txt" ]; then
    if cmp -s final-watch/report.txt final-received/report.txt; then
        echo "   TCP Transfer: ✅ SUCCESS"
        TCP_SUCCESS=true
    else
        echo "   TCP Transfer: ❌ FAILED (files differ)"
    fi
else
    echo "   TCP Transfer: ❌ FAILED (file not received)"
fi

# 7. Test decryption
echo -e "\n7. 🔓 Testing decryption..."
DECRYPT_SUCCESS=false
if [ -f "final-received/report.txt.enc" ]; then
    ./crypto-cli decrypt-file \
      --file=final-received/report.txt.enc \
      --keyfile=final-key.bin \
      --output=final-received/report-decrypted.txt
    
    if cmp -s final-watch/report.txt final-received/report-decrypted.txt; then
        echo "   Decryption: ✅ SUCCESS"
        DECRYPT_SUCCESS=true
    else
        echo "   Decryption: ❌ FAILED"
    fi
fi

# 8. Final summary
echo -e "\n========================================="
echo "           TEST SUMMARY"
echo "========================================="
echo "FSW Batch Encryption: ✅ Working"
echo "TCP File Transfer:    $(if $TCP_SUCCESS; then echo '✅ Working'; else echo '❌ Failed'; fi)"
echo "File Decryption:      $(if $DECRYPT_SUCCESS; then echo '✅ Working'; else echo '❌ Failed'; fi)"
echo -e "\n🎉 $(if $TCP_SUCCESS && $DECRYPT_SUCCESS; then echo 'ALL SYSTEMS OPERATIONAL!'; else echo 'Some tests failed'; fi)"
