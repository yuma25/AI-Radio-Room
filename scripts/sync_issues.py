import re
import subprocess
import sys

def create_issue(title, body):
    cmd = ["gh", "issue", "create", "--title", title, "--body", body]
    try:
        result = subprocess.run(cmd, check=True, capture_output=True, text=True)
        print(f"Created issue: {title}")
        print(result.stdout)
    except subprocess.CalledProcessError as e:
        print(f"Failed to create issue: {title}")
        print(e.stderr)

def main():
    with open("docs/issues.md", "r", encoding="utf-8") as f:
        content = f.read()

    # Split by horizontal rule
    sections = content.split("---")
    
    for section in sections:
        # Find title
        title_match = re.search(r"## Issue #\d+: (.*)", section)
        if not title_match:
            continue
        
        title = title_match.group(1).strip()
        
        # Everything after the title line is the body
        body = section.split(title_match.group(0))[1].strip()
        
        create_issue(title, body)

if __name__ == "__main__":
    main()
