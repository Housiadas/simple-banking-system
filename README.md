# Simple Banking System (Tycoon)
A simple banking API built with Go `v1.22`.

## Setup

```
make docker/build
make docker/up
```

## Migration

```
make db/migrate/create
make db/migrate/up
make db/migrate/down
```

## Testing

```
make tests
```


## Database Schema
We used https://dbdiagram.io to define our database structure.

![Alt text](./docs/database/Simple_Banking_System.png "Simple_Banking_System")
