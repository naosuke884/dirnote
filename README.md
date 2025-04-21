# dirnote

A CLI application to manage notes by directory.

## Features

- Save notes for each directory
- Add, view, edit, delete, and list notes
- Interactive directory selection for viewing notes
- Simple and user-friendly CLI interface

## Installation

1. Download the latest binary from the [Releases page](https://github.com/naosuke884/dirnote/releases).
2. Make the downloaded binary executable:
   ```bash
   chmod +x dirnote
   ```
3. Move the binary to a directory in your PATH:
   ```bash
   mv dirnote /usr/local/bin/
   ```

## Usage

### Add a Note

Add a note to the current directory:
```bash
dirnote add "Your note content"
```

### View a Note

View the note in the current directory:
```bash
dirnote view
```

View a note interactively by selecting a directory:
```bash
dirnote view --interactive
```

### Edit a Note

Edit the note in the current directory:
```bash
dirnote edit
```

### Delete a Note

Delete the note in the current directory:
```bash
dirnote delete
```

### List All Notes

List all notes across directories:
```bash
dirnote list
```

The note for the current directory is marked with `*`.

## Storage

The application creates a directory `~/.dirnote` in your home directory. Inside this directory, a file named `dirnote.db` is used to store all notes. This file is managed automatically by the application.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
