# Interactive Task List

The interactive task list provides a vim-like interface for managing your tasks efficiently.

## Usage

Launch the interactive mode with:
```bash
godo list -i
# or
godo list --interactive
```

## Navigation

- `j` or `↓` - Move cursor down
- `k` or `↑` - Move cursor up  
- `g` - Go to first task
- `G` - Go to last task

## Status Changes

Change the status of the currently selected task:

- `t` - Set to **Todo** ⭕
- `p` - Set to **In Progress** 🔄 (automatically sets start time)
- `d` - Set to **Done** ✅
- `s` - Set to **Paused** ⏸️

## Editing

- `e` or `i` - Edit task title
- `o` - Edit task description (supports multi-line)

## Task Management

- `x` - Delete the currently selected task

### Edit Mode Controls

When editing **title**:
- Type to add characters (including spaces)
- `Backspace` - Delete previous character
- `Enter` - Save changes
- `Escape` - Cancel editing
- `Ctrl+U` - Clear entire line

When editing **description**:
- Type to add characters (including spaces)
- `Backspace` - Delete previous character
- `Enter` - Add new line (multi-line support)
- `Ctrl+S` - Save changes
- `Escape` - Cancel editing
- `Ctrl+U` - Clear entire text

## Other Commands

- `h` or `?` - Show help message
- `q` or `Ctrl+C` - Quit interactive mode

## Task Display

Tasks are organized by status in the following order:
1. **In Progress** 🔄 (with elapsed time)
2. **Todo** ⭕ 
3. **Paused** ⏸️
4. **Done** ✅ (with total time if available)

The currently selected task is highlighted and marked with `►`.

## Tips

- Changes are saved immediately when you confirm them
- Status messages appear at the bottom and auto-hide after 3 seconds
- Use `j`/`k` for quick navigation through your task list
- The interface updates in real-time when you make changes
- Be careful with the delete function (`x`) - there's no undo!
- **Multi-line descriptions**: Use `Enter` to add line breaks, `Ctrl+S` to save
- **Spaces work**: You can now add spaces in both titles and descriptions
- **Quick clear**: Use `Ctrl+U` to quickly clear the current edit field
- **Enhanced editing UI**: Description editor has a bordered box with line/character count
- **Visual cursor**: Animated cursor shows your current position while editing
- **Line wrapping**: Long lines are automatically wrapped in the description editor
- **Multi-line display**: Descriptions show with proper line breaks and visual indicators