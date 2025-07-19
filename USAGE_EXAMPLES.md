# GoTo Usage Examples

This document provides practical examples of using GoTo's interactive task management features.

## Basic Task Management

### Adding Tasks
```bash
# Add a simple task
godo add "Write documentation"

# Add a task with description
godo add "Implement user authentication" -d "Add login/logout functionality with JWT tokens"
```

### Viewing Tasks

#### Static List View
```bash
# View all tasks in a formatted list
godo list
```

#### Interactive Mode
```bash
# Launch interactive mode with vim-like controls
godo list -i
```

## Interactive Mode Workflow Examples

### Quick Status Updates
1. Launch interactive mode: `godo list -i`
2. Use `j`/`k` to navigate to your task
3. Press `p` to start working (sets to "in-progress")
4. When done, press `d` to mark as complete
5. Press `q` to exit

### Editing Tasks on the Fly
1. Navigate to a task with `j`/`k`
2. Press `e` to edit the title
3. Type your changes (spaces work!) and press `Enter` to save
4. Press `o` to edit the description (multi-line supported)
5. In description mode: use `Enter` for new lines, `Ctrl+S` to save
6. Press `Escape` to cancel any edit, `Ctrl+U` to clear current field

### Organizing Your Workflow
```
Current workflow example:

🔄 IN-PROGRESS
► #3 Fix authentication bug ⏱ 2h 15m
     Investigate JWT token expiration issues

⭕ TODO  
  #4 Write unit tests
  #5 Update documentation
     Add examples for new API endpoints
     Include authentication examples
     Update installation guide

⏸️ PAUSED
  #1 Refactor database layer
     Waiting for team review

✅ DONE
  #2 Setup CI/CD pipeline ⌛ 4h 30m
```

### Keyboard Shortcuts Reference

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Move down | Navigate to next task |
| `k` / `↑` | Move up | Navigate to previous task |
| `g` | Go to top | Jump to first task |
| `G` | Go to bottom | Jump to last task |
| `t` | Set todo | Mark task as todo |
| `p` | Set progress | Start working on task |
| `d` | Set done | Mark task as complete |
| `s` | Set paused | Pause current work |
| `e` / `i` | Edit title | Modify task title (Enter to save) |
| `o` | Edit description | Modify multi-line description (Ctrl+S to save) |
| `x` | Delete | Remove task (careful!) |
| `h` / `?` | Help | Show help message |
| `q` | Quit | Exit interactive mode |

## Advanced Usage Tips

### Time Tracking
- When you set a task to "in-progress" (`p`), it automatically starts tracking time
- Switch to "paused" (`s`) to stop the timer while keeping your progress
- Completed tasks show total time worked

### Efficient Navigation
- Use `g` and `G` to quickly jump to the beginning or end of your task list
- The interface automatically groups tasks by status for better organization
- Currently selected task is highlighted and marked with `►`

### Batch Operations
While the interactive mode is great for individual tasks, you can also:
1. Use the regular CLI commands for bulk operations
2. Edit tasks in interactive mode, then use CLI for complex queries
3. Combine both approaches based on your current workflow needs

## Integration with Development Workflow

### Example: Bug Fix Workflow
```bash
# Add the bug as a task
godo add "Fix login redirect issue" -d "Users aren't redirected properly after login"

# Start working on it interactively
godo list -i
# Navigate to the task, press 'p' to start

# Add detailed investigation notes using 'o' for description:
# Press 'o', then type:
# "Investigation steps:
# - Check JWT token expiration
# - Verify redirect URL configuration  
# - Test with different browsers
# 
# Found issue: OAuth callback URL mismatch"
# Press Ctrl+S to save

# When you need to switch contexts
# Press 's' to pause, work on something else

# When ready to finish
# Press 'p' to resume, then 'd' when complete
```

### Example: Feature Development
```bash
# Break down the feature into tasks
godo add "Design user profile UI"
godo add "Implement profile API endpoints" 
godo add "Add profile tests"
godo add "Update documentation"

# Use interactive mode to manage the workflow
godo list -i
# Move through tasks with j/k, update status as you progress
```

## Best Practices

1. **Keep titles concise** - You can add details in the multi-line description
2. **Use status transitions** - todo → in-progress → done (or paused)
3. **Edit on the fly** - Use `e` for titles, `o` for detailed descriptions
4. **Multi-line descriptions** - Use Enter for line breaks, organize information clearly
5. **Regular cleanup** - Use `x` to remove completed or obsolete tasks
6. **Time awareness** - Pay attention to the time tracking to understand your productivity patterns

## Troubleshooting

### Common Issues
- **Can't see tasks**: Make sure you're in the right directory where you added tasks
- **Cursor not moving**: Check that you have tasks in your list
- **Edit mode stuck**: Press `Escape` to cancel edit mode
- **Wrong status**: Use the status keys (`t`/`p`/`d`/`s`) to fix task status

### Getting Help
- Press `h` or `?` in interactive mode for quick help
- Use `godo --help` for general CLI help
- Check the status bar at the bottom for current mode and available actions