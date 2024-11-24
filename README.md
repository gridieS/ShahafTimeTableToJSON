A go program that converts a Shahaf Time Table website into a computer-readable JSON Format

# Usage

## Build the program binary:
```bash
go build ShahafTimeTableToJSON.go
```

## Run it with the flags you'd like:
```bash
./ShahafTimeTableToJSON --url <Your URL>
```

Then the program would create another file named "output.json" which will contain the parsed time table.

Available flags:

### --url (REQUIRED)
The website url of the shahaf time table

### --stdout
Instead of creating a output.json file, the program outputs the json to the standard output.

### --class
Checks the timetable of the specific class, default to 7-1 (7th grade, class 1)


