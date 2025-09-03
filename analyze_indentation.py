#!/usr/bin/env python3

import re
import sys

def analyze_terminal_output(output):
    """Analyze terminal output for cumulative indentation problems"""
    
    print("=== TERMINAL OUTPUT ANALYSIS ===")
    print(f"Raw output length: {len(output)} characters")
    
    # Remove ANSI escape sequences for cleaner analysis
    ansi_escape = re.compile(r'\x1b\[[0-9;]*[a-zA-Z]')
    clean_output = ansi_escape.sub('', output)
    
    print("=== CLEANED OUTPUT ===")
    lines = clean_output.split('\n')
    for i, line in enumerate(lines):
        if line.strip():  # Only non-empty lines
            print(f"Line {i+1}: '{line}' (len: {len(line)})")
    
    # Look for our test content
    test_lines = []
    for i, line in enumerate(lines):
        if any(content in line for content in ['aaaa', 'bbbb', 'cccc', 'dddd']):
            # Count leading spaces
            leading_spaces = len(line) - len(line.lstrip(' '))
            content = line.strip()
            test_lines.append({
                'line_num': i+1,
                'content': content,
                'leading_spaces': leading_spaces,
                'raw_line': repr(line)
            })
    
    print("\n=== TEST CONTENT ANALYSIS ===")
    if not test_lines:
        print("No test content found!")
        return False
    
    for tl in test_lines:
        print(f"Line {tl['line_num']}: '{tl['content']}' -> {tl['leading_spaces']} spaces")
        print(f"  Raw: {tl['raw_line']}")
    
    # Check for cumulative indentation problem
    continuation_lines = [tl for tl in test_lines if not '✦' in tl['content']]
    
    print(f"\n=== INDENTATION CHECK ===")
    print(f"Found {len(continuation_lines)} continuation lines")
    
    if len(continuation_lines) >= 2:
        expected_indent = continuation_lines[0]['leading_spaces']
        print(f"Expected indentation: {expected_indent} spaces")
        
        cumulative_problem = False
        for i, tl in enumerate(continuation_lines):
            if i > 0 and tl['leading_spaces'] > continuation_lines[i-1]['leading_spaces']:
                print(f"❌ CUMULATIVE INDENT FOUND: Line '{tl['content']}' has {tl['leading_spaces']} spaces, more than previous line's {continuation_lines[i-1]['leading_spaces']} spaces")
                cumulative_problem = True
            elif tl['leading_spaces'] != expected_indent:
                print(f"❌ INCONSISTENT INDENT: Line '{tl['content']}' has {tl['leading_spaces']} spaces, expected {expected_indent}")
            else:
                print(f"✅ Line '{tl['content']}' has correct {tl['leading_spaces']} spaces")
        
        if not cumulative_problem:
            print("✅ No cumulative indentation problem detected")
        
        return cumulative_problem
    else:
        print("Not enough continuation lines for analysis")
        return False

if __name__ == "__main__":
    if len(sys.argv) > 1:
        # Read from file
        with open(sys.argv[1], 'r') as f:
            output = f.read()
    else:
        # Read from stdin
        output = sys.stdin.read()
    
    has_problem = analyze_terminal_output(output)
    sys.exit(1 if has_problem else 0)