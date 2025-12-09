# 🚀 Human Git History (CLI Tool)

A clean, human-friendly way to explore your Git history — without scrolling through a wall of cryptic commits like you’re hacking the Pentagon.

This CLI gives you:
- readable commit logs  
- optional stats  
- author filters  
- ASCII graphs  
- changelog-style formatting  
- compact mode  
- date filtering  
- merge-only views  

Basically: **Git logs, but without the stress.**

---

## 📦 Prerequisites

Before running the tool, make sure you have these installed:

### **1. Go (version 1.20 or higher)**
Download Go: [https://go.dev/dl/](https://go.dev/dl/)  

Check your version:
```bash
go version


# Basic human-friendly format
git-history

# Show last 10 commits
git-history -n 10

# Show commits from specific author
git-history --author "John Doe"

# Show with statistics
git-history --stats

# Show with ASCII graph
git-history --graph

# Compact view
git-history --compact

# Changelog format
git-history --format changelog

# Detailed format
git-history --format detailed

# Show commits from last week
git-history --since "1 week ago"

# Show only merge commits
git-history --merges

# Combine options
git-history -n 20 --stats --graph --author "alice@example.com"