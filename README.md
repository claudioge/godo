# godo

_godo_ is a command-line todo list manager that helps you track and organize tasks.

## Features

- **Interactive mode with vim-like keybindings** (default when running `godo`)
- Add, complete, and remove todo items
- List all outstanding tasks
- Mark tasks as done
- Track time spent on tasks
- Organize tasks by priority
- Save tasks persistently
- Simple command-line interface

## Quick Start

### Interactive Mode (Default)

Simply run `godo` without any arguments to enter the interactive mode:

```bash
godo
```

### Interactive Mode Keybindings

- `j/k` or `↓/↑`: Navigate through tasks
- `a`: Add a new task (or add existing task to project if on project)
- `A`: Add a new project
- `m`: Move existing task to a project
- `e`: Edit task title
- `o`: Edit task description
- `g`: Add/edit git links
- `t`: Toggle task status between todo and in-progress
- `s`: Mark task as paused
- `p`: Mark task as in-progress
- `d`: Mark task as done
- `x`: Delete task or project
- `q` or `Ctrl+C`: Quit

### Command Line Mode

You can also use godo with traditional command-line arguments:

```bash
# Add a new task
godo add "Task title" "Optional description"

# List all tasks (non-interactive)
godo list

# Start working on a task
godo start <task-id>

# Mark task as done
godo set <task-id> done

# Delete a task
godo del <task-id>
```
