# Day 1: Linux Basics - File Navigation and Manipulation

## Description

Understanding how to navigate and manipulate files in a Linux system is a fundamental skill for any developer. In this challenge, you'll learn and practice basic Linux commands for file system operations.

## Tasks

Complete the following tasks using Linux command line tools:

1. Navigate and explore directories
2. Create, copy, move, and delete files
3. Search for text within files
4. Change file permissions

## Environment

You'll be working in a Linux environment with the following structure:

```
/home/student/
├── documents/
│   ├── report.txt
│   └── notes/
│       └── meeting_notes.txt
├── pictures/
│   ├── vacation.jpg
│   └── profile.png
└── temp/
```

## Challenges

### 1. Directory Navigation (10 points)

- Print your current working directory
- List all files in your home directory, including hidden ones
- Navigate to the `documents/notes` directory
- Create a new directory called `archive` in your home directory

### 2. File Operations (30 points)

- Create a new file called `todo.txt` in your home directory with the content: "Learn Linux commands"
- Copy `documents/report.txt` to the `archive` directory
- Move `pictures/profile.png` to the `documents` directory
- Create a symbolic link in your home directory pointing to `documents/notes`
- Append the text "Meeting scheduled for Friday" to `documents/notes/meeting_notes.txt`

### 3. Text Search (30 points)

- Find all files in your home directory containing the word "meeting"
- Count the number of lines in `documents/report.txt`
- Display the first 5 lines of `documents/report.txt`
- Find all `.txt` files in your home directory and its subdirectories

### 4. File Permissions (30 points)

- Change the permissions of `todo.txt` to make it readable and writable by the owner only
- Make `documents/notes/meeting_notes.txt` readable by everyone, but writable only by the owner
- Check the current permissions of all files in the `documents` directory

## Learning Resources

### Basic Navigation Commands

```
pwd           # Print Working Directory - shows current directory
ls            # List files and directories
ls -la        # List all files (including hidden) in long format
cd directory  # Change Directory
cd ..         # Move up one directory
cd ~          # Move to home directory
mkdir name    # Make Directory
```

### File Operations

```
touch file            # Create an empty file or update timestamp
cat file              # Display file contents
cp source dest        # Copy file
mv source dest        # Move or rename file
rm file              # Remove file
rm -r directory      # Remove directory and its contents
ln -s source link    # Create symbolic link
```

### Text Processing

```
grep pattern file     # Search for pattern in file
grep -r pattern dir   # Search recursively in directory
wc -l file           # Count lines in file
head -n N file       # Display first N lines
tail -n N file       # Display last N lines
find dir -name pattern # Find files by name
```

### File Permissions

```
chmod permissions file  # Change file permissions
  Examples:
  chmod 600 file       # Owner: rw, Group: -, Others: -
  chmod 644 file       # Owner: rw, Group: r, Others: r
  chmod 755 file       # Owner: rwx, Group: rx, Others: rx
  
ls -l                   # List files with permissions
```

## Submission

To complete this challenge, you must write the exact commands you would use to accomplish each task. Submit your answers in a text file called `linux_commands.txt`.
