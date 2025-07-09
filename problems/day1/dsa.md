# Day 1: DSA Challenge - Palindrome Checker

## Description

A palindrome is a word, phrase, or sequence that reads the same backward as forward, ignoring spaces, punctuation, and capitalization.

For example:
- "racecar" is a palindrome
- "A man, a plan, a canal: Panama" is a palindrome when ignoring spaces and punctuation
- "hello" is not a palindrome

## Task

Write a function that checks if a given string is a palindrome, ignoring spaces, punctuation, and capitalization.

## Function Signature

```python
def is_palindrome(s: str) -> bool:
    # Your code here
    pass
```

```javascript
function isPalindrome(s) {
    // Your code here
}
```

```go
func IsPalindrome(s string) bool {
    // Your code here
}
```

## Input

A string `s` consisting of printable ASCII characters.

## Output

Return `true` if the input string is a palindrome (ignoring spaces, punctuation, and capitalization). Otherwise, return `false`.

## Examples

**Example 1:**
```
Input: "racecar"
Output: true
```

**Example 2:**
```
Input: "A man, a plan, a canal: Panama"
Output: true
```

**Example 3:**
```
Input: "hello"
Output: false
```

## Constraints

- `0 <= s.length <= 10^5`
- `s` consists of printable ASCII characters

## Learning Resources

### String Manipulation

Strings are sequences of characters. In most programming languages, you can access individual characters by their index and perform various operations on them.

#### Python
```python
# Converting to lowercase
s = "Hello World"
lower_s = s.lower()  # "hello world"

# Removing characters
import re
cleaned_s = re.sub(r'[^a-zA-Z0-9]', '', s)  # Removes non-alphanumeric characters
```

#### JavaScript
```javascript
// Converting to lowercase
const s = "Hello World";
const lowerS = s.toLowerCase();  // "hello world"

// Removing characters
const cleanedS = s.replace(/[^a-zA-Z0-9]/g, '');  // Removes non-alphanumeric characters
```

#### Go
```go
// Converting to lowercase
import "strings"
s := "Hello World"
lowerS := strings.ToLower(s)  // "hello world"

// Removing characters
import "regexp"
re := regexp.MustCompile(`[^a-zA-Z0-9]`)
cleanedS := re.ReplaceAllString(s, "")  // Removes non-alphanumeric characters
```

### Two Pointers Technique

The two pointers technique is a common approach for solving problems involving sequences. It involves using two pointers (usually moving from the beginning and end of the sequence) to compare or manipulate elements.

```
1. Initialize two pointers: one at the beginning and one at the end of the string
2. While the pointers haven't crossed:
   a. Compare the characters at the two pointers
   b. If they match, move the pointers toward each other
   c. If they don't match, return false
3. If the loop completes without returning false, return true
```

## Hints

1. You'll need to ignore spaces, punctuation, and capitalization when checking for palindromes.
2. Consider cleaning the string first by removing non-alphanumeric characters and converting it to lowercase.
3. The two-pointer technique is an efficient way to check if a string is a palindrome.

## Submit Your Solution

Write your solution in your preferred language and submit it for evaluation. Your code should be clean, efficient, and properly commented.
