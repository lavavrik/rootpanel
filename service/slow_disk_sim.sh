#!/bin/bash

# ==============================================================================
# Disk Read/Write Simulation Script
#
# This script automates the process of:
# 1. Creating a large test file to simulate a write operation.
# 2. Enabling a slow disk profile using macOS's 'dmc' tool on the
#    directory containing the file.
# 3. Reading the large test file back to simulate a throttled read operation.
# 4. Cleaning up all created files, directories, and dmc profiles.
#
# It will prompt for your password for 'sudo' commands.
# ==============================================================================

# --- Configuration ---
# The directory where the test will be performed. /tmp is a good choice.
TEST_DIR="/tmp/disk_speed_test"

# The name of the large file to be created.
TEST_FILE="large_file.dat"

# The full path to the test file.
FILE_PATH="${TEST_DIR}/${TEST_FILE}"

# The size of the test file in Megabytes (MB). 1024 = 1GB.
FILE_SIZE_MB=1024

# The dmc profile index to use for throttling.
# Run 'dmc list' to see all available options.
# 0: Faulty 5400 HDD (very slow)
# 1: 5400 HDD
# 3: Slow SSD
DMC_PROFILE_INDEX=0
DMC_PROFILE_NAME="Faulty 5400 HDD"

# --- Helper function for colored output ---
print_step() {
  printf "\n\033[1;34m===> %s\033[0m\n" "$1"
}

# --- Cleanup Function ---
# This function is called when the script exits to ensure we clean up.
cleanup() {
  print_step "Cleaning up..."
  # Stop the dmc profile on the test directory.
  # The '&>/dev/null || true' part suppresses errors if it wasn't running.
  if [ -d "$TEST_DIR" ]; then
    sudo dmc stop "$TEST_DIR" &>/dev/null || true
  fi
  # Remove the test directory and the large file.
  rm -rf "$TEST_DIR"
  echo "Cleanup complete."
}

# --- Main Script ---

# Trap the EXIT signal to ensure the cleanup function is always called.
trap cleanup EXIT

# Start by cleaning up any potential leftovers from a previous run.
cleanup

# Create the test directory.
mkdir -p "$TEST_DIR"
echo "Test directory created at: $TEST_DIR"

# Step 1: Write the large file (un-throttled write speed).
print_step "Step 1: Writing a ${FILE_SIZE_MB}MB test file (measuring normal write speed)"
# Use 'dd' to create a file of a specific size from /dev/zero.
# The output of this command will show you the un-throttled write speed.
dd if=/dev/zero of="$FILE_PATH" bs=1m count="$FILE_SIZE_MB"

# Step 2: Enable the slow dmc profile.
print_step "Step 2: Enabling DMC profile '${DMC_PROFILE_NAME}' on '$TEST_DIR'"
# This command requires admin privileges.
sudo dmc start "$TEST_DIR" "$DMC_PROFILE_INDEX"

# Verify that the profile is active.
echo "Verifying DMC status:"
dmc status "$TEST_DIR"

# Step 3: Read the file back (throttled read speed).
print_step "Step 3: Reading the ${FILE_SIZE_MB}MB file (measuring throttled read speed)"
echo "Your monitoring software should now report a significant drop in read speed."
# Use 'dd' to read the file and send its contents to /dev/null.
# The output of this command will show you the throttled read speed.
dd if="$FILE_PATH" of=/dev/null bs=1m

print_step "Simulation Complete"
# The 'trap' will automatically call the cleanup function now.
