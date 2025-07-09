# Day 1: Build Challenge - Simple Command-Line Todo List

## Description

For today's build challenge, you'll create a simple command-line todo list application. This project will help you practice essential programming concepts like file I/O, command parsing, and data structures, while building a useful tool you can actually use.

## Requirements

Build a command-line todo list application with the following features:

1. Add tasks to the list
2. List all tasks with their status (completed or pending)
3. Mark tasks as completed
4. Delete tasks
5. Save tasks to a file so they persist between program runs

## Specifications

### Command-Line Interface

Your application should respond to the following commands:

```
todo add "Task description"        # Add a new task
todo list                         # List all tasks
todo done NUMBER                  # Mark task NUMBER as done
todo undone NUMBER                # Mark task NUMBER as not done
todo delete NUMBER                # Delete task NUMBER
todo help                         # Show help message
```

### Data Storage

Tasks should be saved to a file called `tasks.json` (or similar format) in the user's home directory. Each task should have:

1. A unique identifier or position number
2. A description
3. A completed status (true/false)
4. (Optional) A creation timestamp

### Example Usage

```
$ todo add "Buy groceries"
Task added: "Buy groceries"

$ todo add "Finish homework"
Task added: "Finish homework"

$ todo list
1. [ ] Buy groceries
2. [ ] Finish homework

$ todo done 1
Task marked as done: "Buy groceries"

$ todo list
1. [✓] Buy groceries
2. [ ] Finish homework

$ todo delete 2
Task deleted: "Finish homework"

$ todo list
1. [✓] Buy groceries
```

## Implementation Options

You can implement this project in any language of your choice:

### Python

```python
import json
import os
import sys

# Code structure
def add_task(description):
    # Implementation

def list_tasks():
    # Implementation
    
def mark_task_done(task_number):
    # Implementation
    
# And so on...
```

### JavaScript (Node.js)

```javascript
const fs = require('fs');
const path = require('path');

// Code structure
function addTask(description) {
    // Implementation
}

function listTasks() {
    // Implementation
}

// And so on...
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    // Other imports
)

// Code structure
func addTask(description string) {
    // Implementation
}

func listTasks() {
    // Implementation
}

// And so on...
```

## Learning Resources

### File I/O

#### Python
```python
# Reading from a file
with open('filename.txt', 'r') as f:
    content = f.read()
    
# Writing to a file
with open('filename.txt', 'w') as f:
    f.write('Hello, world!')
```

#### JavaScript
```javascript
const fs = require('fs');

// Reading from a file
const content = fs.readFileSync('filename.txt', 'utf8');

// Writing to a file
fs.writeFileSync('filename.txt', 'Hello, world!');
```

#### Go
```go
// Reading from a file
content, err := os.ReadFile("filename.txt")
if err != nil {
    // Handle error
}

// Writing to a file
err := os.WriteFile("filename.txt", []byte("Hello, world!"), 0644)
if err != nil {
    // Handle error
}
```

### JSON Handling

#### Python
```python
import json

# Parse JSON
data = json.loads('{"name": "John", "age": 30}')

# Convert to JSON
json_string = json.dumps(data)
```

#### JavaScript
```javascript
// Parse JSON
const data = JSON.parse('{"name": "John", "age": 30}');

// Convert to JSON
const jsonString = JSON.stringify(data);
```

#### Go
```go
import "encoding/json"

// Parse JSON
var data map[string]interface{}
json.Unmarshal([]byte(`{"name": "John", "age": 30}`), &data)

// Convert to JSON
jsonBytes, err := json.Marshal(data)
```

## Bonus Challenges

Once you've completed the basic requirements, try implementing these additional features:

1. Add priorities to tasks (high, medium, low)
2. Add due dates to tasks
3. Filter tasks by status (completed/pending)
4. Add task categories/tags
5. Create a simple terminal UI with colors

## Submission

Submit your complete source code along with a README.md file explaining how to build and run your application.
