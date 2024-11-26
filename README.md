A go program that converts a Shahaf Time Table website into a computer-readable JSON File Format.

# Usage

## Build the program binary:
```bash
go build
```

## Run it with the flags you'd like:
```bash
./ShahafTimeTableToJSON --url <Your URL>
```
## Or build and run the program at once:
```bash
go run *.go <options>
```

Available flags:

### --url (REQUIRED)
The website url of the shahaf time table.

### --list
Instead of generating a time table json for the specified class, the program generates a json list of all of the classes available.

### --class <class-num>
Checks the timetable of $class-num, default to first class in class list.

### --output <file-name>
Instead of printing the json in stdout, the program modifies/creates a file $file-name.

