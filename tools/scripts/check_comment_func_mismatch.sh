#!/bin/bash

#========================================================================================
# Script that would run in current dir or path if specified to detect Go files
# with mismatches between function names and the comments above them
# Usage: ./check_comment_func_mismatch.sh ./path/to/project
#========================================================================================

# Directory to scan (default: current directory)
SCAN_DIR="${1:-.}"

# Configurable list of directories and files to ignore (relative to SCAN_DIR)
IGNORE_PATHS=(
    # "tests"
    "crypto/ssi-sdk.go"

    # Note: Need to handle globs properly
    "*_test.go"
)

# Build find command with ignore patterns
find_args=("$SCAN_DIR" -type f -name '*.go')
for path in "${IGNORE_PATHS[@]}"; do
    # Check if the path is a glob pattern for file names
    if [[ "$path" == *_test.go ]]; then
        find_args+=(-not -name "$path")
    else
        find_args+=(-not -path "$SCAN_DIR/$path" -not -path "$SCAN_DIR/$path/*")
    fi
done

# Temporary file to capture mismatches
tmpfile=$(mktemp)

# Find comment blocks and function declarations, excluding ignored paths
# Capture output to tmpfile while displaying it
find "${find_args[@]}" -exec awk '
BEGIN { comment = ""; comment_line_num = 0; in_comment = 0; full_comment = "" }
/^[[:space:]]*\/\/[[:space:]]*[[:alnum:]]/ && !/^[[:space:]]*\/\/go:/ {
    if (!in_comment) {
        in_comment = 1
        comment_line_num = FNR
        comment = $0
        sub(/^[[:space:]]*\/\//, "", comment)
        sub(/^[[:space:]]*/, "", comment)
        full_comment = comment
    }
    next
}
/^[[:space:]]*func[[:space:]]+[[:alnum:]]/ {
    if (in_comment && full_comment != "") {
        func_name = $0
        sub(/.*func[[:space:]]+/, "", func_name)
        # Extract function name up to [ or (
        sub(/[\[\(].*/, "", func_name)
        # Clean up any remaining punctuation or spaces
        sub(/[[:space:]]*[[:punct:]]*$/, "", func_name)
        sub(/[[:space:]].*$/, "", func_name)
        # Check if full_comment starts with func_name
        if (index(full_comment, func_name) == 1) {
            # Comment starts with function name, no mismatch
        } else {
            # Extract first word for reporting
            comment = full_comment
            sub(/[[:space:]].*$/, "", comment)
            if (comment != func_name) {
                printf "Mismatch in %s:%d: Comment: %s, Function: %s\n", FILENAME, comment_line_num, comment, func_name
            }
        }
    }
    in_comment = 0
    comment = ""
    full_comment = ""
    next
}
/^[[:space:]]*$|^[[:space:]]*\/\/go:/ || !/^[[:space:]]*\/\// {
    in_comment = 0
    comment = ""
    full_comment = ""
    next
}
' {} + | tee "$tmpfile"

# Check if mismatches were found
if [ -s "$tmpfile" ]; then
    # Non-empty tmpfile means mismatches were found
    rm -f "$tmpfile"
    exit 1
else
    # Empty tmpfile means no mismatches
    rm -f "$tmpfile"
    exit 0
fi
